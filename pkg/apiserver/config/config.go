package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/xutao1989103/oam-go/pkg/client/k8s"
	"reflect"
	"strings"
)

const (
	defaultConfigurationName = "oam"
	defaultConfigurationPath = "/Users/diyi/config"
)

type Config struct {
	K8sOptions *k8s.K8sOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
}

func NewConfig() *Config {
	return &Config{
		K8sOptions: k8s.NewK8sOptions(),
	}
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	result := make(map[string]bool, 0)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}

func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := NewConfig()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
