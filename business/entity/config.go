// Package entity provides entities for business logic.
package entity

import "github.com/forest33/honeybee/pkg/config"

type Config struct {
	MQTT         *MQTT         `yaml:"MQTT"`
	Logger       *Logger       `yaml:"Logger"`
	Runtime      *Runtime      `yaml:"Runtime"`
	Scripts      *Scripts      `yaml:"Scripts"`
	Scheduler    *Scheduler    `yaml:"Scheduler"`
	Bot          *Bot          `yaml:"Bot"`
	Notification *Notification `yaml:"Notification"`
}

type MQTT struct {
	Host                 string `yaml:"Host" default:"127.0.0.1"`
	Port                 int    `yaml:"Port" default:"1883"`
	ClientID             string `yaml:"ClientID" default:"honeybee"`
	User                 string `yaml:"User" default:""`
	Password             string `yaml:"Password" default:""`
	UseTLS               bool   `yaml:"UseTLS"  default:"false"`
	ServerTLS            bool   `yaml:"ServerTLS"  default:"false"`
	CACert               string `yaml:"CACert"  default:""`
	Cert                 string `yaml:"Cert"  default:""`
	Key                  string `yaml:"Key" default:""`
	ConnectRetryInterval int    `yaml:"ConnectRetryInterval" default:"3"`
	Timeout              int    `yaml:"Timeout" default:"10"`
}

type Scripts struct {
	Folder           []string `yaml:"Folder" default:"./config/scripts"`
	RegistrySize     int      `yaml:"RegistrySize" default:"32768"`
	RegistryMaxSize  int      `yaml:"RegistryMaxSize" default:"65536"`
	RegistryGrowStep int      `yaml:"RegistryGrowStep" default:"32"`
}

type Scheduler struct {
	Enabled           bool `yaml:"Enabled" default:"true"`
	MaxTasksPerSender int  `yaml:"MaxTasksPerSender" default:"0"`
}

type Bot struct {
	Enabled       bool    `yaml:"Enabled" default:"false"`
	Token         string  `yaml:"Token"`
	ChatId        []int64 `yaml:"ChatId"`
	UpdateTimeout int     `yaml:"UpdateTimeout" default:"60"`
	PoolSize      int     `yaml:"PoolSize" default:"2"`
}

type Notification struct {
	Enabled  bool   `yaml:"Enabled" default:"false"`
	BaseURL  string `yaml:"BaseURL" default:"https://ntfy.sh"`
	Timeout  int    `yaml:"Timeout" default:"30"`
	Priority string `yaml:"Priority" default:"default"`
	PoolSize int    `yaml:"PoolSize" default:"2"`
}

type Logger struct {
	Level             string `yaml:"Level" default:"debug"`
	TimeFormat        string `yaml:"TimeFormat" default:"2006-01-02T15:04:05.000000"`
	PrettyPrint       bool   `yaml:"PrettyPrint" default:"false"`
	DisableSampling   bool   `yaml:"DisableSampling" default:"true"`
	RedirectStdLogger bool   `yaml:"RedirectStdLogger" default:"true"`
	ErrorStack        bool   `yaml:"ErrorStack" default:"true"`
}

type Runtime struct {
	GoMaxProcs int `yaml:"GoMaxProcs" default:"0"`
}

type ConfigHandler interface {
	Update(data interface{})
	Save()
	GetPath() string
	AddObserver(f func(interface{})) error
}

func GetConfig() (ConfigHandler, *Config, error) {
	cfg := &Config{}
	h, err := config.New(cfg)
	if err != nil {
		return nil, nil, err
	}
	return h, cfg, nil
}
