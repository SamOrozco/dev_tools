package main

import "fmt"

type Config struct {
	Endpoint       *Endpoint `yaml:"endpoint"`
	JsFile         string    `yaml:"js_file"`
	Limit          int       `yaml:"limit"`
	IntervalMillis int       `yaml:"interval_millis"`
	SuccessMessage string    `yaml:"success_message"`
	Auth           *Auth     `yaml:"auth"`
}

func (config Config) String() string {
	return fmt.Sprintf("endpoint: %s, js_file: %s", config.Endpoint, config.JsFile)
}

type Endpoint struct {
	Url    string `yaml:"url"`
	Method string `yaml:"method"`
}

type Auth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
