package main

type ConfigValidator interface {
	ValidateConfig(config *Config) error
}

type defaultConfigValidator struct {
}

func NewDefaultConfigValidator() ConfigValidator {
	return &defaultConfigValidator{}
}

func (d defaultConfigValidator) ValidateConfig(config *Config) error {
	return nil
}
