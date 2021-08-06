package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Log struct {
	Level   string `yaml:"level"`
	NoColor bool   `yaml:"no_color"`
	InJson  bool   `yaml:"in_json"`
}

func (c *Log) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Log
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
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

type target struct {
	Raw string
	URL *url.URL
}

func (d *target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	err := unmarshal(&raw)
	if err != nil {
		return err
	}
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	*d = target{
		Raw: raw,
		URL: u,
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

func (c *BaseInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain BaseInfo
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")
	if c.Title == "" {
		c.Title = "Statusetat"
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

type Config struct {
	Targets    []target   `yaml:"targets"`
	Listen     string     `yaml:"listen"`
	Log        Log        `yaml:"log"`
	Components Components `yaml:"components"`
	BaseInfo   *BaseInfo  `yaml:"base_info"`
	Username   string     `yaml:"username"`
	Password   string     `yaml:"password"`

	CookieKey string     `yaml:"cookie_key"`
	Notifiers []Notifier `yaml:"notifiers"`

	Theme *Theme `yaml:"theme"`
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

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("At least one target must be define")
	}

	if len(c.Components) == 0 {
		return fmt.Errorf("At least one component must be define")
	}

	if c.Username == "" {
		c.Username = uuid.NewString()
		log.Infof("Generated username (set username in config): %s", c.Username)
	}

	if c.Password == "" {
		c.Password = uuid.NewString()
		log.Infof("Generated password (set password in config): %s", c.Password)
	}
	if c.BaseInfo == nil {
		c.BaseInfo = &BaseInfo{
			Title: "Statusetat",
		}
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
	if c.BaseInfo.BaseURL == "" {
		c.BaseInfo.BaseURL = "http://" + c.Listen
	}
	if c.BaseInfo.Support == "" {
		c.BaseInfo.Support = c.BaseInfo.BaseURL
	}
	if c.BaseInfo.Contact == "" {
		c.BaseInfo.Contact = c.BaseInfo.Support
	}
	if c.CookieKey == "" {
		c.CookieKey = uuid.NewString()
	}
	if c.BaseInfo.TimeZone == "" {
		c.BaseInfo.TimeZone = "UTC"
	}
	if c.Theme == nil {
		c.Theme = &Theme{}
	}
	if c.Theme.PersistentDisplayName == "" {
		c.Theme.PersistentDisplayName = "persistent incident"
	}

	return nil
}

func LoadConfig(filename string) (Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
