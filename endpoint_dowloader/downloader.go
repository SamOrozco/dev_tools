package main

type Downloader struct {
	Endpoint       *Endpoint `yaml:"endpoint"`
	FilePrefix     string    `yaml:"file_prefix"`
	FileExtension  string    `yaml:"file_extension"`
	IntervalMillis int       `yaml:"interval_millis"`
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
