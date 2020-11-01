package main

import (
	"bytes"
	logger2 "dev_tools/endpoint_watcher/logger"
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
	"sync"
	"time"
)

var vm = otto.New()
var httpClient = http.DefaultClient
var globalLogger = logger2.NewStdOutLogger("global")

func main() {
	registerJavascriptFunctions()

	if len(os.Args) < 2 {
		panic("no yaml file passed")
	}

	yamlFiles := os.Args[1:]
	var wg sync.WaitGroup
	for i := range yamlFiles {
		wg.Add(1)
		var idx = i
		go func() {
			defer wg.Done()
			handleYamlFileLocation(yamlFiles[idx])
		}()
	}

	// wait for all to be done
	wg.Wait()
}

// handle yaml file string
func handleYamlFileLocation(location string) {
	// expand env vars to their actual end vars
	yamlFile := os.ExpandEnv(location)
	data, err := files.ReadBytesFromFile(yamlFile)
	if err != nil {
		panic(err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(err)
	}

	// read config from file if applicable if file is set on config
	if len(config.ConfigFile) > 0 {
		config = readConfigFromFileIfApplicable(config)
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
	log := logger2.NewStdOutLogger(config.Name)

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
	log.Debug(fmt.Sprintf("testing [%s] endpoint with %d attempt", config.Endpoint.Url, config.Limit))
	for i := 0; i < config.Limit; i++ {
		resp, err := httpClient.Do(&request)
		if err != nil {
			println(err.Error())
		}
		if handleResponse(resp, js) {
			executeSuccess(config)

			if len(config.Name) > 0 {
				log.Debug(fmt.Sprintf("%s has finished successfully", config.Name))
			} else {
				log.Debug("finished successfully")
			}
			return
		}

		// every 10 print which req we're on
		if i%10 == 0 {
			log.Debug(fmt.Sprintf("on request %d", i))
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

		// if there is a condition on success
		if currentSuccess.If != nil {
			switch currentSuccess.If.Type {
			case "js":
				if !handleJsIf(currentSuccess.If) {
					continue
				}
				break
			default: // if no type specified defaults to js
				if !handleJsIf(currentSuccess.If) {
					continue
				}
				break
			}
		}

		// desktop notification
		if successType == "desktop" {
			handleDesktopSuccess(currentSuccess.Message)
		} else if successType == "webhook" {
			handleWebhookSuccess(currentSuccess.Endpoint)
		} else if successType == "watcher" {
			handleWatcherSuccess(currentSuccess.Config)
		} else if successType == "js" {
			handleJsSuccess(currentSuccess.Js)
		} else {
			handleDesktopSuccess(currentSuccess.Message)
		}
	}
}

func handleJsIf(jsIf *If) bool {
	if jsIf.Js == nil {
		panic("js if must have js field defined")
	}
	js := getJsContents(jsIf.Js)
	return executeJsAndGetDefValue(js)
}

func handleWatcherSuccess(config *Config) {
	// read config from file if applicable
	config = readConfigFromFileIfApplicable(config)
	prepareAndRunConfig(config)
}

func handleJsSuccess(js *Js) {
	javascript := getJsContents(js)
	val, err := vm.Run(javascript)
	if err != nil {
		println(err.Error())
		return
	}
	if val.IsDefined() {
		println(val.String())
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

	if err := vm.Set("statusCode", resp.StatusCode); err != nil {
		println(err.Error())
	}

	if err := vm.Set("responseBody", string(respBytes)); err != nil {
		println(err.Error())
	}
	return executeJsAndGetDefValue(js)
}

func executeJsAndGetDefValue(js string) bool {
	if _, err := vm.Run(js); err != nil {
		panic(err)
	}

	resultVal, err := vm.Get("def")
	if err != nil {
		globalLogger.Error("unable to get def value for executeJsAndGetDefValue")
		globalLogger.Error(js)
		return false
	}

	if val, err := resultVal.ToBoolean(); err != nil {
		globalLogger.Error("unable to convert def value to boolean")
		globalLogger.Error(js)
		return false
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

	var javascript string
	if jsType == "file" {
		javascript = readJSStringFromFile(js.Js)
	} else if jsType == "script" {
		javascript = js.Js
	} else {
		javascript = readJSStringFromFile(js.Js)
	}
	return os.ExpandEnv(javascript)
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
		prepareJs(success.Js)
	}
}

// if file is set read from file else return current config
func readConfigFromFileIfApplicable(config *Config) *Config {
	if len(config.ConfigFile) > 0 {
		return readConfigFromFile(config.ConfigFile)
	}
	return config
}

// read config from yaml to config struct
func readConfigFromFile(fileLocation string) *Config {
	data, err := files.ReadBytesFromFile(fileLocation)
	if err != nil {
		panic(err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(err)
	}
	return config
}

func registerJavascriptFunctions() {
	// set env variable
	err := vm.Set("setEnv", func(call otto.FunctionCall) otto.Value {
		key := call.Argument(0).String()
		value := call.Argument(1).String()
		err := os.Setenv(key, value)
		if err != nil {
			println(err.Error())
		}
		return otto.Value{}
	})
	if err != nil {
		panic(err)
	}

	// getEnv variable
	err = vm.Set("getEnv", func(call otto.FunctionCall) otto.Value {
		key := call.Argument(0).String()
		val := os.Getenv(key)
		jsVal, err := vm.ToValue(val)
		if err != nil {
			return otto.Value{}
		}
		return jsVal
	})

	if err != nil {
		panic(err)
	}
}
