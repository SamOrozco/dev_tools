package main

type Config struct {
	Name           string     `yaml:"name"`        // optional
	ConfigFile     string     `yaml:"config_file"` // optional, only if you want to load config from file
	Endpoint       *Endpoint  `yaml:"endpoint"`
	Cond           *Condition `yaml:"cond"`
	Limit          int        `yaml:"limit"`
	IntervalMillis int        `yaml:"interval_millis"`
	Success        []*Success `yaml:"success"`
}

type Endpoint struct {
	Url     string              `yaml:"url"`
	Method  string              `yaml:"method"`
	Body    string              `yaml:"body"`
	Headers map[string][]string `yaml:"headers"`
	Auth    *Auth               `yaml:"auth"`
}

type Auth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Success struct {
	If       *Condition `yaml:"if"`
	Type     string     `yaml:"type"`
	Message  string     `yaml:"message"`
	Endpoint *Endpoint  `yaml:"endpoint"`
	Watcher  *Config    `yaml:"watcher"`
	Cond     *Condition `yaml:"js"`
}

type Condition struct {
	Type string `yaml:"type"`
	Js   string `yaml:"javascript"`
}
