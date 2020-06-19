package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	"github.com/xdhuxc/kubernetes-transform/src/model"
)

type Config struct {
	Address string `yaml:"address"`
	App     string `yaml:"app"`
	Env     string `yaml:"env"`

	Mode      string    `yaml:"mode"`
	Policy    string    `yaml:"policy"`
	Namespace Namespace `yaml:"namespace"`
	Resource  Resource  `yaml:"resource"`

	Source model.Cluster `yaml:"source"`
	Target model.Cluster `yaml:"target"`

	Database Database `yaml:"database"`

	Debug bool `yaml:"debug"`
}

type Namespace struct {
	Name       string   `yaml:"name"`
	Action     string   `yaml:"action"`
	Namespaces []string `yaml:"namespaces"`
}

type Resource struct {
	Name      string   `yaml:"name"`
	Action    string   `yaml:"action"`
	Resources []string `yaml:"resources"`
	Kinds     []string `yaml:"kinds"`
}

type Database struct {
	Host         string `yaml:"host"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Name         string `yaml:"name"`
	Log          bool   `yaml:"log"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

func (c *Config) String() string {
	if dataInBytes, err := json.Marshal(c); err == nil {
		return string(dataInBytes)
	}

	return ""
}

var config Config

func InitConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	if config.Debug {
		klog.Infof("the config is ", config.String())
	}

	return nil
}

func GetConfig() Config {
	return config
}
