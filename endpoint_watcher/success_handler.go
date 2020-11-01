package main

import (
	"errors"
	"github.com/gen2brain/beeep"
	"github.com/robertkrimen/otto"
	"strings"
)

type SuccessHandler interface {
	Success(config *Config, configRunner ConfigRunner) error
	DesktopSuccess(message string) error
	WebhookSuccess(endpoint *Endpoint) error
	WatcherSuccess(config *Config, configRunner ConfigRunner) error
	JsSuccess(cond *Condition) error
}

type defaultSuccessHandler struct {
	endpointRequestBuilder EndpointRequestBuilder
	fileLoader             FileLoader
	vm                     *otto.Otto
	jsLoader               JsLoader
	ifHandler              IfHandler
}

func NewDefaultSuccessHandler(
	endpointRequestBuilder EndpointRequestBuilder,
	fileLoader FileLoader,
	vm *otto.Otto,
	jsLoader JsLoader,
	ifHandler IfHandler,
) SuccessHandler {
	return &defaultSuccessHandler{
		endpointRequestBuilder: endpointRequestBuilder,
		fileLoader:             fileLoader,
		vm:                     vm,
		jsLoader:               jsLoader,
		ifHandler:              ifHandler,
	}
}

func (d defaultSuccessHandler) Success(config *Config, configRunner ConfigRunner) error {
	if config.Success == nil || len(config.Success) < 1 {
		return d.DesktopSuccess("Endpoint response passed condition")
	}

	for i := range config.Success {
		currentSuccess := config.Success[i]
		successType := strings.ToLower(currentSuccess.Type)

		// if current success has an if ONLY execute if is true
		if currentSuccess.If != nil {
			if !d.ifHandler.If(currentSuccess.If) {
				continue
			}
		}

		if successType == "desktop" {
			if err := d.DesktopSuccess(currentSuccess.Message); err != nil {
				return err
			}
		} else if successType == "webhook" {
			if err := d.WebhookSuccess(currentSuccess.Endpoint); err != nil {
				return err
			}
		} else if successType == "watcher" {
			if err := d.WatcherSuccess(currentSuccess.Watcher, configRunner); err != nil {
				return err
			}
		} else if successType == "js" {
			if err := d.JsSuccess(currentSuccess.Cond); err != nil {
				return err
			}
		} else {
			if err := d.DesktopSuccess(currentSuccess.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

// sends a desktop notification for running computer with message
func (d defaultSuccessHandler) DesktopSuccess(message string) error {
	return beeep.Alert("Test Passed", message, "assets/information.png")
}

// execute endpoint on success
func (d defaultSuccessHandler) WebhookSuccess(endpoint *Endpoint) error {
	if endpoint == nil {
		return errors.New("success configured for webhook but no endpoint supplied")
	}
	successRequest := d.endpointRequestBuilder.BuildRequestFromEndpoint(endpoint)
	_, err := httpClient.Do(successRequest)
	if err != nil {
		return err
	}
	return nil
}

// executes another watcher on success
func (d defaultSuccessHandler) WatcherSuccess(config *Config, configRunner ConfigRunner) error {
	if config == nil {
		return errors.New("watcher success configured but watcher configuration was supplier")
	}
	return configRunner.RunConfig(config)
}

func (d defaultSuccessHandler) JsSuccess(cond *Condition) error {
	js := d.jsLoader.LoadJsForCond(cond)
	_, err := vm.Run(js)
	if err != nil {
		return err
	}
	return nil
}
