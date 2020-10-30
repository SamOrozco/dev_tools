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

	// expand env vars to their actual end vars
	yamlFile := os.ExpandEnv(os.Args[1])
	data, err := files.ReadBytesFromFile(yamlFile)
	if err != nil {
		panic(err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(err)
	}

	// run config
	prepareAndRunConfig(config)
}

func prepareAndRunConfig(config *Config) {
	prepareConfig(config)
	runConfig(config)
}

// run the endpoint watcher config
func runConfig(config *Config) {
	// validate config
	if !validateConfig(config) {
		panic("keys `endpoint` and `js_file` must be filled out")
	}

	// read js file
	js := getJsContents(config.Js)

	// call endpoint
	request := buildRequest(config.Endpoint)

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

	if config.Success == nil || len(config.Success) < 1 {
		handleBasicSuccess()
	}

	for i := range config.Success {
		currentSuccess := config.Success[i]
		successType := strings.ToLower(currentSuccess.Type)
		// desktop notification
		if successType == "desktop" {
			handleDesktopSuccess(currentSuccess.Message)
		} else if successType == "webhook" {
			handleWebhookSuccess(currentSuccess.Endpoint)
		} else if successType == "watcher" {
			handleWatcherSuccess(currentSuccess.Config)
		} else {
			handleDesktopSuccess(currentSuccess.Message)
		}
	}
}

func handleWatcherSuccess(config *Config) {
	prepareAndRunConfig(config)
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
	request := http.Request{
		Method: endpoint.Method,
		URL:    uri,
		Header: http.Header{},
		Body:   ioutil.NopCloser(bytes.NewReader([]byte(endpoint.Body))),
	}

	// if has auth set it
	if endpoint.Auth != nil {
		if len(endpoint.Auth.Password) == 0 || len(endpoint.Auth.Username) == 0 {
			panic("must supply a username and password with auth")
		}

		// default basic auth
		if endpoint.Auth.Type == "" {
			endpoint.Auth.Type = "basic"
		}

		return addAuthToRequest(request, endpoint.Auth)
	}
	return request
}

func readJSStringFromFile(jsFile string) string {
	data, err := files.ReadBytesFromFile(jsFile)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func validateConfig(config *Config) bool {
	// tood implement
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

func getJsContents(js *Js) string {
	jsType := strings.ToLower(js.Type)

	if len(js.Js) < 1 {
		panic("js needs to be supplied")
	}

	if jsType == "file" {
		return readJSStringFromFile(js.Js)
	} else if jsType == "script" {
		return js.Js
	} else {
		// default file
		return readJSStringFromFile(js.Js)
	}
}

// PREPARE FUNCS

func prepareConfig(config *Config) {
	if config != nil {
		// todo this way of preparing is very manual and prone to error
		// todo for any variables that might be added
		// prepare config strings
		prepareJs(config.Js)
		prepareEndpoint(config.Endpoint)
		prepareSuccesses(config.Success)
	}
}

func prepareJs(js *Js) {
	if js != nil {
		js.Js = os.ExpandEnv(js.Js)
		js.Type = os.ExpandEnv(js.Type)
	}

}

func prepareEndpoint(endpoint *Endpoint) {
	if endpoint != nil {
		endpoint.Method = os.ExpandEnv(endpoint.Method)
		endpoint.Body = os.ExpandEnv(endpoint.Body)
		endpoint.Url = os.ExpandEnv(endpoint.Url)
		if endpoint.Auth != nil {
			prepareAuth(endpoint.Auth)
		}
	}
}

func prepareAuth(auth *Auth) {
	if auth != nil {
		auth.Username = os.ExpandEnv(auth.Username)
		auth.Password = os.ExpandEnv(auth.Password)
		auth.Type = os.ExpandEnv(auth.Type)
	}
}

func prepareSuccesses(successes []*Success) {
	if successes != nil && len(successes) > 0 {
		for i := range successes {
			prepareSuccess(successes[i])
		}
	}
}

func prepareSuccess(success *Success) {
	if success != nil {
		success.Type = os.ExpandEnv(success.Type)
		success.Message = os.ExpandEnv(success.Message)
		prepareEndpoint(success.Endpoint)
		prepareConfig(success.Config)
	}
}
