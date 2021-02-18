package storages

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type s3Session struct {
	bucket  string
	path    string
	svc     *s3.S3
	awsSess *session.Session
}

type S3 struct {
	sess *s3Session
}

func (s *S3) Create(incident models.Incident) (models.Incident, error) {
	b, _ := json.Marshal(incident)
	uploader := s3manager.NewUploader(s.sess.awsSess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &s.sess.bucket,
		Key:    aws.String(incident.GUID),
		Body:   bytes.NewBuffer(b),
	})

	return incident, err
}

func (s S3) retrieveSubscribers() ([]string, error) {
	obj, err := s.sess.svc.GetObject(&s3.GetObjectInput{
		Bucket: &s.sess.bucket,
		Key:    aws.String(subscriberFilename),
	})
	if err != nil {
		if aerr, ok := err.(awserr.RequestFailure); ok && aerr.StatusCode() == 404 {
			return []string{}, os.ErrNotExist
		}
		return []string{}, err
	}
	defer obj.Body.Close()
	subs := make([]string, 0)
	err = json.NewDecoder(obj.Body).Decode(&subs)
	if err != nil {
		return []string{}, err
	}
	return subs, err
}

func (s S3) storeSubscribers(subscribers []string) error {
	b, _ := json.Marshal(subscribers)
	uploader := s3manager.NewUploader(s.sess.awsSess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &s.sess.bucket,
		Key:    aws.String(subscriberFilename),
		Body:   bytes.NewBuffer(b),
	})
	return err
}

func (s S3) Subscribe(email string) error {
	subs, _ := s.retrieveSubscribers()
	if common.InStrSlice(email, subs) {
		return nil
	}
	subs = append(subs, email)
	return s.storeSubscribers(subs)
}

func (s S3) Unsubscribe(email string) error {
	subs, err := s.retrieveSubscribers()
	if err != nil {
		return err
	}
	subs = common.FilterStrSlice(email, subs)
	return s.storeSubscribers(subs)
}

func (s S3) Subscribers() ([]string, error) {
	return s.retrieveSubscribers()
}

func (s *S3) Update(guid string, incident models.Incident) (models.Incident, error) {
	incident.GUID = guid
	return s.Create(incident)
}

func (s *S3) Delete(guid string) error {
	_, err := s.sess.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.sess.bucket,
		Key:    aws.String(guid),
	})
	return err
}

func (s *S3) Read(guid string) (models.Incident, error) {
	obj, err := s.sess.svc.GetObject(&s3.GetObjectInput{
		Bucket: &s.sess.bucket,
		Key:    aws.String(guid),
	})
	if err != nil {
		if aerr, ok := err.(awserr.RequestFailure); ok && aerr.StatusCode() == 404 {
			return models.Incident{}, os.ErrNotExist
		}
		return models.Incident{}, err
	}
	defer obj.Body.Close()
	var incident models.Incident
	sort.Sort(models.Messages(incident.Messages))
	err = json.NewDecoder(obj.Body).Decode(&incident)
	return incident, err
}

func (s *S3) ByDate(from, to time.Time) ([]models.Incident, error) {
	objs, err := s.sess.svc.ListObjects(&s3.ListObjectsInput{
		Bucket: &s.sess.bucket,
	})
	if err != nil {
		return []models.Incident{}, err
	}
	incidents := make([]models.Incident, 0)
	for _, obj := range objs.Contents {
		if *obj.Key == subscriberFilename {
			continue
		}
		// we can be in the future but not in the past inside an incident
		// se we check for the past for earning time (to not retrieve file content)
		if obj.LastModified.Before(to) {
			continue
		}
		incident, err := s.Read(*obj.Key)
		if err != nil {
			return incidents, err
		}
		if incident.CreatedAt.After(from) || incident.CreatedAt.Before(to) {
			continue
		}
		sort.Sort(models.Messages(incident.Messages))
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (s *S3) Ping() error {
	_, err := s.sess.svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: &s.sess.bucket,
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

func (s S3) Detect(u *url.URL) bool {
	return u.Scheme == "s3"
}

func (s S3) urlToSession(u *url.URL) (*s3Session, error) {
	if strings.HasSuffix(u.Host, "s3.amazonaws.com") && (u.User == nil || u.User.Username() == "") {
		bucket, path := s.extractBucketPath(u)
		sess, err := session.NewSession()
		if err != nil {
			return nil, err
		}
		return &s3Session{
			bucket:  bucket,
			path:    path,
			svc:     s3.New(sess),
			awsSess: sess,
		}, nil
	}
	pathStyle := true
	bucket, path := s.extractBucketPath(u)
	host := u.Host
	if strings.HasSuffix(host, ".s3.amazonaws.com") {
		host = strings.TrimPrefix(host, bucket+".")
		pathStyle = false
	}
	creds := credentials.NewEnvCredentials()
	if u.User != nil || u.User.Username() != "" {
		pass, _ := u.User.Password()
		creds = credentials.NewStaticCredentials(u.User.Username(), pass, "")
	}
	region := "us-east-1"
	if u.Query().Get("region") != "" {
		region = u.Query().Get("region")
	}
	u.User = nil
	sess, err := session.NewSession(&aws.Config{
		Credentials:      creds,
		Endpoint:         &host,
		Region:           &region,
		HTTPClient:       makeHttpClient(u),
		S3ForcePathStyle: &pathStyle,
	})
	if err != nil {
		return nil, err
	}
	return &s3Session{
		bucket:  bucket,
		path:    path,
		svc:     s3.New(sess),
		awsSess: sess,
	}, nil
}

func (s S3) extractBucketPath(u *url.URL) (bucket string, path string) {
	if strings.HasSuffix(u.Host, ".s3.amazonaws.com") {
		return strings.TrimSuffix(u.Host, ".s3.amazonaws.com"), u.Path
	}
	split := strings.Split(u.Path, "/")
	return split[1], strings.Join(split[2:], "/")
}
