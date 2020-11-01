package main

type ConfigFileLoader interface {
	LoadConfigFromFile(fileLocation string) (*Config, error)
}

type configFileLoader struct {
	fileLoader      FileLoader
	configConverter ConfigConverter
}

func NewConfigFileLoader(fileLoader FileLoader, configConverter ConfigConverter) ConfigFileLoader {
	return &configFileLoader{
		fileLoader:      fileLoader,
		configConverter: configConverter,
	}
}

func (c configFileLoader) LoadConfigFromFile(fileLocation string) (*Config, error) {
	fileData, err := c.fileLoader.LoadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	return c.configConverter.ByteToConfig(fileData)
}
