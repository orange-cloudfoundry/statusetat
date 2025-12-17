package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	VCAP_SERVICES    = "VCAP_SERVICES"
	VCAP_APPLICATION = "VCAP_APPLICATION"
)

type Config struct {
	Targets                      Targets    `yaml:"targets"`
	Listen                       string     `yaml:"listen"`
	Log                          *Log       `yaml:"log"`
	Components                   Components `yaml:"components"`
	BaseInfo                     *BaseInfo  `yaml:"base_info"`
	Username                     string     `yaml:"username"`
	Password                     string     `yaml:"password"`
	TlsConfig                    *TlsConfig `yaml:"tls"`
	CookieKey                    string     `yaml:"cookie_key"`
	Notifiers                    []Notifier `yaml:"notifiers"`
	DisableMaintenanceToIncident bool       `yaml:"disable_maintenance_to_incident"`

	Theme *Theme `yaml:"theme"`
}

type TlsConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func (c *TlsConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain TlsConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.CertFile == "" {
		return fmt.Errorf("tls cert_file is required")
	}
	if c.KeyFile == "" {
		return fmt.Errorf("tls key_file is required")
	}
	return nil
}

func (c *Config) Merge(other Config) {
	c.Targets = append(c.Targets, other.Targets...)
	c.Notifiers = append(c.Notifiers, other.Notifiers...)
	c.Components = append(c.Components, other.Components...)
	if len(c.Listen) == 0 {
		c.Listen = other.Listen
	}
	if len(c.Username) == 0 || len(c.Password) == 0 {
		c.Username = other.Username
		c.Password = other.Password
	}
	if len(c.CookieKey) == 0 {
		c.CookieKey = other.CookieKey
	}
	if c.Theme == nil {
		c.Theme = other.Theme
	}
	if c.BaseInfo == nil {
		c.BaseInfo = other.BaseInfo
	}
	if c.Log == nil {
		c.Log = other.Log
	}
	if !c.DisableMaintenanceToIncident {
		c.DisableMaintenanceToIncident = other.DisableMaintenanceToIncident
	}
}

func (c *Config) Validate() error {
	if len(c.Components) == 0 {
		return fmt.Errorf("at least one component must be define")
	}

	if c.Username == "" {
		c.Username = uuid.NewString()
		log.Infof("generated username (set username in config)")
	}

	if c.Password == "" {
		c.Password = uuid.NewString()
		log.Infof("generated password (set password in config)")
	}

	host := "0.0.0.0"
	port := "8080"
	splitListen := strings.Split(c.Listen, ":")
	if splitListen[0] != "" {
		host = splitListen[0]
	}
	if len(splitListen) == 2 {
		port = splitListen[1]
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port = envPort
	}
	c.Listen = host + ":" + port

	if c.CookieKey == "" {
		c.CookieKey = uuid.NewString()
	}

	if c.Theme == nil {
		c.Theme = &Theme{}
	}
	if err := c.Theme.Validate(); err != nil {
		return nil
	}

	if c.BaseInfo == nil {
		c.BaseInfo = &BaseInfo{}
	}
	if err := c.BaseInfo.Validate(c.Listen); err != nil {
		return nil
	}

	if c.Log == nil {
		c.Log = &Log{}
	}
	if err := c.Log.Validate(); err != nil {
		return err
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target must be define")
	}
	if err := c.Targets.Validate(); err != nil {
		return err
	}

	return nil
}

type Target string

func (t Target) Validate() (*url.URL, error) {
	log.Debugf("url: %s", string(t))
	u, err := url.Parse(string(t))
	if err != nil {
		return nil, err
	}
	log.Debugf("-> value: %+v", u)
	return u, nil
}

type Targets []Target

func (t Targets) Validate() error {
	for _, v := range t {
		if _, err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Log struct {
	Level   string `yaml:"level"`
	NoColor bool   `yaml:"no_color"`
	InJson  bool   `yaml:"in_json"`
}

func (c *Log) Validate() error {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: c.NoColor,
	})
	if c.Level != "" {
		lvl, err := log.ParseLevel(c.Level)
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
	}
	if c.InJson {
		log.SetFormatter(&log.JSONFormatter{})
	}
	return nil
}

type BaseInfo struct {
	BaseURL  string `yaml:"base_url"`
	Support  string `yaml:"support"`
	Contact  string `yaml:"contact"`
	Title    string `yaml:"title"`
	TimeZone string `yaml:"time_zone"`
}

func (b *BaseInfo) Validate(listen string) error {
	b.BaseURL = strings.TrimSuffix(b.BaseURL, "/")
	if b.Title == "" {
		b.Title = "Statusetat"
	}
	if b.BaseURL == "" {
		b.BaseURL = "http://" + listen
	}
	if b.Support == "" {
		b.Support = b.BaseURL
	}
	if b.Contact == "" {
		b.Contact = b.Support
	}
	if b.TimeZone == "" {
		b.TimeZone = "UTC"
	}

	return nil
}

type Notifier struct {
	For    ForComponent           `yaml:"for"`
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`
}

type Theme struct {
	PreStatus  string `yaml:"pre_status"`
	PostStatus string `yaml:"post_status"`

	PreTimeline  string `yaml:"pre_timeline"`
	PostTimeline string `yaml:"post_timeline"`

	PreMaintenance  string `yaml:"pre_maintenance"`
	PostMaintenance string `yaml:"post_maintenance"`

	PersistentDisplayName string `yaml:"persistent_display_name"`
	PrePersistent         string `yaml:"pre_persistent"`
	PostPersistent        string `yaml:"post_persistent"`

	Footer string `yaml:"footer"`
}

func (t *Theme) Validate() error {
	if t.PersistentDisplayName == "" {
		t.PersistentDisplayName = "persistent incident"
	}
	return nil
}

type Component struct {
	Name        string
	Description string
	Group       string
}

type Components []Component

func (cs Components) Regroups() map[string][]Component {
	regroups := make(map[string][]Component)
	for _, c := range cs {
		_, ok := regroups[c.Group]
		if ok {
			regroups[c.Group] = append(regroups[c.Group], c)
			continue
		}
		regroups[c.Group] = []Component{c}
	}
	return regroups
}

func (c Component) String() string {
	if c.Group == "" {
		return c.Name
	}
	return c.Group + " - " + c.Name
}

func (cs Components) Inline() []string {
	components := make([]string, len(cs))
	for i, co := range cs {
		components[i] = co.String()
	}
	return components
}

func LoadConfig(content []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func LoadConfigFromFile(filename string) (Config, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	//nolint:ineffassign
	config, err := LoadConfig(b)
	if err != nil {
		return Config{}, err
	}
	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config in file %s: %s", filename, err)
	}

	return config, nil
}

func LoadConfigFromEnv() (Config, error) {
	type config map[string]interface{}
	type configs []config
	type vcapservices map[string]configs

	var (
		services vcapservices
		res      Config
	)

	vcap := os.Getenv(VCAP_SERVICES)
	if vcap == "" {
		return Config{}, fmt.Errorf("empty %s env variable", VCAP_SERVICES)
	}

	if err := json.Unmarshal([]byte(vcap), &services); err != nil {
		return Config{}, fmt.Errorf("invalid json in %s env variable: %s", VCAP_SERVICES, err)
	}

	for _, cConfigs := range services {
		for _, cConfig := range cConfigs {
			name, ok := cConfig["name"]
			if !ok {
				continue
			}
			nameStr, ok := name.(string)
			if !ok {
				continue
			}

			creds, err := json.Marshal(cConfig["credentials"])
			if err != nil {
				return Config{}, fmt.Errorf("unable to marshal config '%s': %s", nameStr, err)
			}

			log.Infof("loading vcap service '%s'", nameStr)
			c, err := LoadConfig(creds)
			if err != nil {
				log.Errorf("error while loading vcap service '%s': %s", nameStr, err)
				continue
			}
			res.Merge(c)
		}
	}

	if err := res.Validate(); err != nil {
		return Config{}, fmt.Errorf("could not find any valid configuration in %s: %s", VCAP_SERVICES, err)
	}
	return res, nil
}
