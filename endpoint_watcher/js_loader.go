package main

import "strings"

type JsLoader interface {
	LoadJsForType(typ, value string) string
	LoadJsForCond(cond *Condition) string
}

type defaultJsLoader struct {
	fileLoader FileLoader
}

func NewDefaultJsLoader(fileLoader FileLoader) JsLoader {
	return &defaultJsLoader{fileLoader: fileLoader}
}

func (d defaultJsLoader) LoadJsForType(typ, value string) string {
	newType := strings.ToLower(typ)
	if newType == "script" {
		return value
	}
	js, err := d.fileLoader.LoadFileString(value)
	if err != nil {
		return ""
	}
	return js
}

func (d defaultJsLoader) LoadJsForCond(cond *Condition) string {
	if cond == nil {
		return ""
	}
	return d.LoadJsForType(cond.Type, cond.Js)
}
