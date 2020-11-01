package main

import (
	logger2 "dev_tools/endpoint_watcher/logger"
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var vm = otto.New()
var httpClient = http.DefaultClient
var globalLogger = logger2.NewStdOutLogger("global")

func main() {
	registerJavascriptFunctions()

	// setup dependencies
	valuePreparer := NewSuppliedPreparer(func(val string) string {
		return os.ExpandEnv(val)
	})
	fileLoader := NewFileLoader()
	jsLoader := NewDefaultJsLoader(fileLoader)
	configConverter := NewYamlConfigConverter()
	configFileLoader := NewConfigFileLoader(fileLoader, configConverter)
	configValidator := NewDefaultConfigValidator()
	endpointRequestBuilder := NewDefaultRequestBuilder()
	ifHandler := NewJsIfHandler(vm, jsLoader)
	httpClient := http.DefaultClient

	conditionResponseHandler := NewDefaultConditionResponseHandler(NewFuncConditionResponseHandler(func(resp *http.Response, cond *Condition) bool {
		if err := vm.Set("statusCode", resp.StatusCode); err != nil {
			panic(err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := vm.Set("responseBody", string(data)); err != nil {
			panic(err)
		}
		return ifHandler.If(cond)
	}))

	// setup success handler
	successHandler := NewDefaultSuccessHandler(
		endpointRequestBuilder,
		fileLoader,
		vm,
		jsLoader,
		ifHandler,
	)

	// run a config for each file given
	asyncWaitGroup := sync.WaitGroup{}
	fileLocations := os.Args[1:]
	for i := range fileLocations {
		currentFile := fileLocations[i]
		curLogger := logger2.NewStdOutLogger(fmt.Sprintf("root logger [%d]", i))
		// run every file concurrently

		// tell the wait group to wait
		asyncWaitGroup.Add(1)
		go func() {
			// create new config runner
			configRunner := NewLocalConfigRunner(
				valuePreparer,
				configFileLoader,
				configValidator,
				endpointRequestBuilder,
				httpClient,
				conditionResponseHandler,
				successHandler,
				curLogger,
			)
			// list for any child builds async
			if err := configRunner.RunConfigFromFile(currentFile); err != nil {
				panic(err)
			}
			asyncWaitGroup.Done()
		}()
	}
	asyncWaitGroup.Wait()

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
