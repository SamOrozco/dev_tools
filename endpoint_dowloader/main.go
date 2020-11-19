package main

import (
	"dev_tools/files"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		panic("no config file provided")
	}

	//httpClient := http.DefaultClient

	data, err := files.ReadBytesFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	config := &Downloader{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
	runDownloader(config)
}

func runDownloader(config *Downloader) {
	println(config.Endpoint.Url)
}

func getHttpRequestFromEndpoint(endpoint *Endpoint) *http.Request {
	return &http.Request{}
}
