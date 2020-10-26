package main

import (
	"dev_tools/files"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/go-yaml/yaml"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

var vm = otto.New()
var httpClient = http.DefaultClient

func main() {

	if len(os.Args) < 2 {
		panic("no yaml file passed")
	}

	data, err := files.ReadBytesFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(err)
	}

	// validate config
	if !validateConfig(config) {
		panic("keys `endpoint` and `js_file` must be filled out")
	}

	// read js file
	js := readJSStringFromFile(config.JsFile)

	// call endpoint
	request := buildRequest(config.Endpoint)

	// set delay time
	interval := config.IntervalMillis
	if interval < 1 {
		interval = 100
	}

	// call endpoint for limit or condition met
	println(fmt.Sprintf("testing endpoint with %d attemtps", config.Limit))
	for i := 0; i < config.Limit; i++ {
		resp, err := httpClient.Do(&request)
		if err != nil {
			println(err.Error())
		}
		if handleResponse(resp, js) {
			executeSuccess(config)
		}

		// wait for defined period of time
		// default 10 seconds
		<-time.After(time.Millisecond * time.Duration(interval))
	}
}

func executeSuccess(config *Config) {
	err := beeep.Notify("Test Passed", config.SuccessMessage, "assets/information.png")
	if err != nil {
		panic(err)
	}
}

func handleResponse(resp *http.Response, js string) bool {

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	vm.Set("statusCode", resp.StatusCode)
	vm.Set("responseBody", string(respBytes))
	if _, err := vm.Run(js); err != nil {
		panic(err)
	}

	resultVal, err := vm.Get("def")
	if err != nil {
		panic(err)
	}

	if val, err := resultVal.ToBoolean(); err != nil {
		panic(err)
	} else {
		return val
	}
}

func buildRequest(endpoint *Endpoint) http.Request {
	uri, err := url.Parse(endpoint.Url)
	if err != nil {
		panic(err)
	}
	return http.Request{
		Method: endpoint.Method,
		URL:    uri,
	}
}

func readJSStringFromFile(jsFile string) string {
	data, err := files.ReadBytesFromFile(jsFile)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func validateConfig(config *Config) bool {
	return config.Endpoint != nil && len(config.JsFile) > 0
}