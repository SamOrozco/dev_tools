package main

import (
	logger2 "dev_tools/endpoint_watcher/logger"
	"fmt"
	"math"
	"net/http"
	"time"
)

type ConfigExec func(config *Config)

type ConfigRunner interface {
	RunConfig(config *Config) error
	RunConfigFromFile(configFileLocation string) error
}

type localConfigRunner struct {
	preparer                 Preparer
	configFileLoader         ConfigFileLoader
	configValidator          ConfigValidator
	endpointRequestBuilder   EndpointRequestBuilder
	httpClient               *http.Client
	conditionResponseHandler ConditionResponseHandler
	successHandler           SuccessHandler
	log                      logger2.Logger
}

func NewLocalConfigRunner(
	preparer Preparer,
	configFileLoader ConfigFileLoader,
	configValidator ConfigValidator,
	endpointRequestBuilder EndpointRequestBuilder,
	httpClient *http.Client,
	conditionResponseHandler ConditionResponseHandler,
	successHandler SuccessHandler,
	logger logger2.Logger,
) ConfigRunner {
	return &localConfigRunner{
		preparer:                 preparer,
		configFileLoader:         configFileLoader,
		configValidator:          configValidator,
		endpointRequestBuilder:   endpointRequestBuilder,
		httpClient:               httpClient,
		conditionResponseHandler: conditionResponseHandler,
		successHandler:           successHandler,
		log:                      logger,
	}
}

func (l localConfigRunner) RunConfigFromFile(configFileLocation string) error {
	l.log.Debug(fmt.Sprintf("loading config file %s", configFileLocation))
	config, err := l.configFileLoader.LoadConfigFromFile(configFileLocation)
	if err != nil {
		return err
	}
	return l.RunConfig(config)
}

func (l localConfigRunner) RunConfig(config *Config) error {
	// prepare config or replace all ${var} with env var
	l.preparer.PrepareConfig(config)
	// get name from config
	configName := ""
	if len(config.Name) > 0 {
		configName = config.Name
	} else {
		configName = GetRandomName(0)
	}
	runLogger := logger2.NewStdOutLogger(configName)

	// we will always prioritize a config_file field over anything else
	if len(config.ConfigFile) > 0 {
		runLogger.Debug(fmt.Sprintf("loading and running config from file because `config_file` flag set %s", config.ConfigFile))
		return l.RunConfigFromFile(config.ConfigFile)
	}

	// validate config
	if err := l.configValidator.ValidateConfig(config); err != nil {
		runLogger.Error(fmt.Sprintf("invalid config %s", err.Error()))
		return err
	}

	if config.IntervalMillis < 1 {
		config.IntervalMillis = 100
	}

	limit := int64(config.Limit)
	if config.Limit < 1 {
		limit = math.MaxInt64
	}

	runLogger.Debug(fmt.Sprintf("starting requests at %s for %s", time.Now().String(), config.Endpoint.Url))
	for i := int64(0); i < limit; i++ {
		request := l.endpointRequestBuilder.BuildRequestFromEndpoint(config.Endpoint)
		resp, err := l.httpClient.Do(request)
		if err != nil {
			runLogger.Error(fmt.Sprintf("error executing http request %s", err.Error()))
		}

		// if condition passes handle successes
		if config.Cond == nil || l.conditionResponseHandler.Handle(resp, config.Cond) {
			if err := l.successHandler.Success(config, l); err != nil {
				panic(err)
			}
			runLogger.Debug("completed successfully")
			return nil
		}

		// print every 10th request so that everyone is on the same page
		if i%10 == 0 {
			runLogger.Debug(fmt.Sprintf("on request %d of %d", i, limit))
		}

		<-time.After(time.Millisecond * time.Duration(config.IntervalMillis))
	}
	// start running test
	// if limit is greater than 1, run that many times, else, run forever
	return nil
}
