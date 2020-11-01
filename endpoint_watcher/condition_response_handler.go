package main

import "net/http"

type ConditionRegistry interface {
	Register(typ string, handler ConditionResponseHandler)
}

type ConditionResponseHandler interface {
	Handle(resp *http.Response, condition *Condition) bool
}

type ConditionRegistryHandler interface {
	ConditionRegistry
	ConditionResponseHandler
}

type defaultConditionResponseHandler struct {
	defaultHandler   ConditionResponseHandler
	typeConditionMap map[string]ConditionResponseHandler
}

type funConditionResponseHandler struct {
	fun func(resp *http.Response, cond *Condition) bool
}

func NewFuncConditionResponseHandler(fun func(resp *http.Response, cond *Condition) bool) ConditionResponseHandler {
	return &funConditionResponseHandler{
		fun: fun,
	}
}

func (f funConditionResponseHandler) Handle(resp *http.Response, condition *Condition) bool {
	return f.fun(resp, condition)
}

func NewDefaultConditionResponseHandler(defaultHandler ConditionResponseHandler) ConditionRegistryHandler {
	return &defaultConditionResponseHandler{
		typeConditionMap: make(map[string]ConditionResponseHandler, 0),
		defaultHandler:   defaultHandler,
	}
}

func NewResponseHandlerFromFunc(fun func(resp *http.Response, cond *Condition) bool) ConditionResponseHandler {
	return &defaultConditionResponseHandler{}
}

func (d defaultConditionResponseHandler) Handle(resp *http.Response, condition *Condition) bool {
	handler := d.typeConditionMap[condition.Type]
	if handler != nil {
		return handler.Handle(resp, condition)
	}
	return d.defaultHandler.Handle(resp, condition)
}

func (d defaultConditionResponseHandler) Register(typ string, handler ConditionResponseHandler) {
	d.typeConditionMap[typ] = handler
}
