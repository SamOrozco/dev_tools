package main

type Downloader struct {
	Url        string `yaml:"url"`
	Method     string `yaml:"method"`
	Body       string `yaml:"body"`
	Auth       *Auth  `yaml:"auth"`
	FilePrefix string `yaml:"file_prefix"`
}

type Auth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
