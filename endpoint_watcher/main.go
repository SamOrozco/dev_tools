package main

import (
	"bytes"
	"dev_tools/files"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/go-yaml/yaml"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	request = addAuthToRequest(request, config.Auth)

	// set delay time
	interval := config.IntervalMillis
	if interval < 1 {
		interval = 100
	}

	// if no limit set or is less than zero set to 10_000
	if config.Limit < 1 {
		config.Limit = 10_000
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
			println("success")
			return
		}

		// every 10 print which req we're on
		if i%10 == 0 {
			println(fmt.Sprintf("on request %d", i))
		}

		// wait for defined period of time
		// default 10 seconds
		<-time.After(time.Millisecond * time.Duration(interval))
	}
}

func executeSuccess(config *Config) {

	if config.Success == nil {
		handleBasicSuccess()
	}

	successType := strings.ToLower(config.Success.Type)
	// desktop notification
	if successType == "desktop" {
		handleDesktopSuccess(config.Success.Message)
	} else if successType == "webhook" {
		handleWebhookSuccess(config.Endpoint)
	} else {
		handleDesktopSuccess(config.Success.Message)
	}
}

func handleWebhookSuccess(endpoint *Endpoint) {
	if endpoint == nil {
		panic("success configured for webhook but no endpoint supplied")
	}
	successRequest := buildRequest(endpoint)
	_, err := httpClient.Do(&successRequest)
	if err != nil {
		panic(err)
	}
}

func handleDesktopSuccess(message string) {
	err := beeep.Alert("Test Passed", message, "assets/information.png")
	if err != nil {
		panic(err)
	}
}

func handleBasicSuccess() {
	err := beeep.Alert("Test Passed", "Endpoint response passed condition", "assets/information.png")
	if err != nil {
		panic(err)
	}
}

func handleResponse(resp *http.Response, js string) bool {

	if resp == nil {
		return false
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println(err.Error())
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
		Header: http.Header{},
		Body:   ioutil.NopCloser(bytes.NewReader([]byte(endpoint.Body))),
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

	// has endpoint and js file
	if !(config.Endpoint != nil && len(config.JsFile) > 0) {
		return false
	}

	// has auth and username and pass
	if config.Auth != nil {
		return len(config.Auth.Username) > 0 && len(config.Auth.Password) > 0
	}
	return true
}

func addAuthToRequest(request http.Request, auth *Auth) http.Request {
	if auth == nil {
		return request
	}
	if strings.ToLower(auth.Type) == "basic" {
		request.SetBasicAuth(auth.Username, auth.Password)
	}
	return request
}
