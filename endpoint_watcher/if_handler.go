package main

import "github.com/robertkrimen/otto"

type IfHandler interface {
	If(iff *Condition) bool
}

type jsIfHandler struct {
	vm       *otto.Otto
	jsLoader JsLoader
}

func NewJsIfHandler(
	vm *otto.Otto,
	jsLoader JsLoader,
) IfHandler {
	return &jsIfHandler{
		vm:       vm,
		jsLoader: jsLoader,
	}
}

func (j jsIfHandler) If(iff *Condition) bool {
	js := j.jsLoader.LoadJsForType(iff.Type, iff.Js)
	if _, err := j.vm.Run(js); err != nil {
		panic(err)
	}
	resultVal, err := j.vm.Get("def")
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
