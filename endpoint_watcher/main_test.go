package main

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestThatThisIsWorking(testing *testing.T) {
	config := Config{
		Endpoint: &Endpoint{
			Url:    "http://my.website.com",
			Method: "get",
			Body:   "",
			Auth: &Auth{ // default basic auth
				Username: "username",
				Password: "my_pass",
			},
		},
		Js: &Js{
			Type: "",
			Js:   "",
		},
		Limit:          100,
		IntervalMillis: 1000,
		Success: &Success{
			Type:    "webhook",
			Message: "",
			Endpoint: &Endpoint{
				Url:    "http://my.other.website.com",
				Method: "POST",
				Body:   `{"jobId":1000"}`,
				Auth: &Auth{
					Type:     "basic",
					Username: "samo",
					Password: "passo",
				},
			},
		},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		panic(err)
	}

	println(string(data))
}
