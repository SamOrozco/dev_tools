package main

type Config struct {
	Name           string     `yaml:"name"`        // optional
	ConfigFile     string     `yaml:"config_file"` // optional, only if you want to load config from file
	Endpoint       *Endpoint  `yaml:"endpoint"`
	Js             *Js        `yaml:"js"`
	Limit          int        `yaml:"limit"`
	IntervalMillis int        `yaml:"interval_millis"`
	Success        []*Success `yaml:"success"`
}

type Endpoint struct {
	Url    string `yaml:"url"`
	Method string `yaml:"method"`
	Body   string `yaml:"body"`
	Auth   *Auth  `yaml:"auth"`
}

type Auth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Success struct {
	Type     string    `yaml:"type"`
	Message  string    `yaml:"message"`
	Endpoint *Endpoint `yaml:"endpoint"`
	Config   *Config   `yaml:"config"`
	Js       *Js       `yaml:"js"`
}

type Js struct {
	Type string `yaml:"type"`
	Js   string `yaml:"javascript"`
}
