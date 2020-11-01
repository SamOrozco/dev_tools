package main

import "gopkg.in/yaml.v2"

type ConfigConverter interface {
	ByteToConfig(data []byte) (*Config, error)
}

type yamlConfigConverter struct {
}

func NewYamlConfigConverter() ConfigConverter {
	return &yamlConfigConverter{}
}

func (y yamlConfigConverter) ByteToConfig(data []byte) (*Config, error) {
	if data == nil {
		return &Config{}, nil
	}
	config := &Config{}
	err := yaml.Unmarshal(data, config)
	return config, err
}
