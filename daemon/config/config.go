package config

import (
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	configPath string
}

// Open configuration from disk.
func (a *Configuration) Open() (*Config, error) {
	file, err := os.Open(a.configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return a.OpenReader(file)
}

// Open configuration from a reader.
func (a *Configuration) OpenReader(r io.Reader) (*Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return getRawConfig(data)
}

func getRawConfig(data []byte) (*Config, error) {
	//log.Print("\r\n", string(data))
	config := &Config{}
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewConfiguration(configPath string) *Configuration {
	return &Configuration{
		configPath: configPath}
}
