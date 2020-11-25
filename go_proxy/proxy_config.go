package main

type ProxyConfig struct {
	Port    int           `yaml:"port"`
	Proxies []*MatchProxy `yaml:"proxies"`
}

type MatchProxy struct {
	Match *Match `yaml:"match"`
	Proxy *Proxy `yaml:"proxy"`
}

type Match struct {
	Type      string     `yaml:"type"`
	PathMatch *PathMatch `yaml:"path_match"`
}

type PathMatch struct {
	MatchValue string    `yaml:"match_value"`
	MatchType  MatchType `yaml:"match_type"`
}

type Proxy struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type MatchType string

const (
	Equals     MatchType = "eq"
	Contains   MatchType = "cont"
	StartsWith MatchType = "sw"
	EndsWith   MatchType = "ew"
)
