package storages

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/orange-cloudfoundry/statusetat/v2/common"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
	"github.com/orange-cloudfoundry/statusetat/v2/utils"
)

type s3Session struct {
	bucket string
	path   string
	client *s3.Client
	cfg    aws.Config
}

type S3 struct {
	sess *s3Session
}

func (s *S3) Create(incident models.Incident) (models.Incident, error) {
	if incident.Persistent {
		err := s.addPersistent(incident)
		return incident, err
	}
	b, _ := json.Marshal(incident)
	uploader := manager.NewUploader(s.sess.client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(incident.GUID),
		Body:   bytes.NewBuffer(b),
	})

	return incident, err
}

func (s *S3) retrieveSubscribers() ([]string, error) {
	obj, err := s.sess.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(subscriberFilename),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404") {
			return []string{}, os.ErrNotExist
		}
		return []string{}, err
	}
	defer utils.CloseAndLogError(obj.Body)
	subs := make([]string, 0)
	err = json.NewDecoder(obj.Body).Decode(&subs)
	if err != nil {
		return []string{}, err
	}
	return subs, err
}

func (s *S3) storeSubscribers(subscribers []string) error {
	b, _ := json.Marshal(subscribers)
	uploader := manager.NewUploader(s.sess.client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(subscriberFilename),
		Body:   bytes.NewBuffer(b),
	})
	return err
}

func (s *S3) addPersistent(incident models.Incident) error {
	incidents, err := s.Persistents()
	if err != nil {
		return err
	}
	incidents = models.Incidents(incidents).Filter(incident.GUID)
	incidents = append(incidents, incident)
	return s.storePersistents(incidents)
}

func (s *S3) removePersistent(guid string) error {
	incidents, err := s.Persistents()
	if err != nil {
		return err
	}
	incidents = models.Incidents(incidents).Filter(guid)
	return s.storePersistents(incidents)
}

func (s *S3) readPersistent(guid string) (models.Incident, error) {
	incidents, err := s.Persistents()
	if err != nil {
		return models.Incident{}, err
	}
	return models.Incidents(incidents).Find(guid), nil
}

func (s *S3) Subscribe(email string) error {
	subs, _ := s.retrieveSubscribers()
	if common.InStrSlice(email, subs) {
		return nil
	}
	subs = append(subs, email)
	return s.storeSubscribers(subs)
}

func (s *S3) Unsubscribe(email string) error {
	subs, err := s.retrieveSubscribers()
	if err != nil {
		return err
	}
	subs = common.FilterStrSlice(email, subs)
	return s.storeSubscribers(subs)
}

func (s *S3) storePersistents(incidents []models.Incident) error {
	sort.Sort(models.Incidents(incidents))
	b, _ := json.Marshal(incidents)
	uploader := manager.NewUploader(s.sess.client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(persistentFilename),
		Body:   bytes.NewBuffer(b),
	})
	return err
}

func (s *S3) Subscribers() ([]string, error) {
	return s.retrieveSubscribers()
}

func (s *S3) Update(guid string, incident models.Incident) (models.Incident, error) {
	if incident.Persistent {
		_ = s.Delete(guid) // nolint
		err := s.addPersistent(incident)
		return incident, err
	}
	_ = s.removePersistent(guid) // nolint
	incident.GUID = guid
	return s.Create(incident)
}

func (s *S3) Delete(guid string) error {
	err := s.removePersistent(guid)
	if err != nil {
		return err
	}
	_, err = s.sess.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(guid),
	})
	return err
}

func (s *S3) Read(guid string) (models.Incident, error) {
	var incident models.Incident
	obj, err := s.sess.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(guid),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404") {
			incident, err := s.readPersistent(guid)
			if err != nil {
				return models.Incident{}, os.ErrNotExist
			}
			if incident.GUID == guid {
				return incident, nil
			}
			return models.Incident{}, os.ErrNotExist
		}
		return models.Incident{}, err
	}
	defer utils.CloseAndLogError(obj.Body)
	err = json.NewDecoder(obj.Body).Decode(&incident)
	if err != nil {
		return models.Incident{}, err
	}
	sort.Sort(models.Messages(incident.Messages))
	return incident, nil
}

func (s *S3) ByDate(from, to time.Time) ([]models.Incident, error) {
	objs, err := s.sess.client.ListObjects(context.TODO(), &s3.ListObjectsInput{
		Bucket: aws.String(s.sess.bucket),
	})
	if err != nil {
		return []models.Incident{}, err
	}
	incidents := make([]models.Incident, 0)
	for _, obj := range objs.Contents {
		if *obj.Key == subscriberFilename ||
			*obj.Key == persistentFilename {
			continue
		}

		incident, err := s.Read(*obj.Key)
		if err != nil {
			return incidents, err
		}
		if incident.CreatedAt.Before(from) || incident.CreatedAt.After(to) {
			continue
		}
		sort.Sort(models.Messages(incident.Messages))
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (s *S3) Ping() error {
	_, err := s.sess.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(s.sess.bucket),
	})
	return err
}

func (s *S3) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		s := &S3{}
		sess, err := s.urlToSession(u)
		if err != nil {
			return nil, err
		}
		s.sess = sess
		return s, nil
	}
}

func (s *S3) Detect(u *url.URL) bool {
	return u.Scheme == "s3"
}

func (s *S3) urlToSession(u *url.URL) (*s3Session, error) {
	bucket, path := s.extractBucketPath(u)

	// Use default config if connecting to standard AWS S3
	if strings.HasSuffix(u.Host, "s3.amazonaws.com") && (u.User == nil || u.User.Username() == "") {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
		return &s3Session{
			bucket: bucket,
			path:   path,
			client: s3.NewFromConfig(cfg),
			cfg:    cfg,
		}, nil
	}

	// Handle custom S3-compatible endpoints
	pathStyle := true
	host := u.Host
	if strings.HasSuffix(host, ".s3.amazonaws.com") {
		host = strings.TrimPrefix(host, bucket+".")
		pathStyle = false
	}

	var credProvider aws.CredentialsProvider
	if u.User != nil && u.User.Username() != "" {
		pass, _ := u.User.Password()
		credProvider = credentials.NewStaticCredentialsProvider(u.User.Username(), pass, "")
	} else {
		// Use default credentials chain (environment variables, IAM role, etc.)
		credProvider = nil
	}

	region := "us-east-1"
	if u.Query().Get("region") != "" {
		region = u.Query().Get("region")
	}

	var cfg aws.Config
	var err error
	if credProvider != nil {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credProvider),
			config.WithRegion(region),
			config.WithHTTPClient(makeHttpClient(u)),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithHTTPClient(makeHttpClient(u)),
		)
	}
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = pathStyle
		o.BaseEndpoint = aws.String("https://" + host)
	})

	return &s3Session{
		bucket: bucket,
		path:   path,
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *S3) Persistents() ([]models.Incident, error) {
	obj, err := s.sess.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.sess.bucket),
		Key:    aws.String(persistentFilename),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404") {
			return []models.Incident{}, nil
		}
		return []models.Incident{}, err
	}
	defer utils.CloseAndLogError(obj.Body)
	subs := make([]models.Incident, 0)
	err = json.NewDecoder(obj.Body).Decode(&subs)
	if err != nil {
		return []models.Incident{}, err
	}
	return subs, err
}

func (s *S3) extractBucketPath(u *url.URL) (bucket string, path string) {
	if strings.HasSuffix(u.Host, ".s3.amazonaws.com") {
		return strings.TrimSuffix(u.Host, ".s3.amazonaws.com"), u.Path
	}
	split := strings.Split(u.Path, "/")
	return split[1], strings.Join(split[2:], "/")
}
