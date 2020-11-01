package main

import "dev_tools/files"

type FileLoader interface {
	LoadFile(location string) ([]byte, error)
	LoadFileString(location string) (string, error)
}

type diskFileLoader struct {
}

func NewFileLoader() FileLoader {
	return &diskFileLoader{}
}

func (d diskFileLoader) LoadFile(location string) ([]byte, error) {
	return files.ReadBytesFromFile(location)
}

func (d diskFileLoader) LoadFileString(location string) (string, error) {
	data, err := d.LoadFile(location)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
