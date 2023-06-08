package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"primerbitcoin/pkg/utils"
)

type Config struct {
	Port     int `yaml:"port"`
	Database struct {
		Host string `yaml:"host"`
	} `yaml:"database"`
	Scheduler struct {
		Schedule string `yaml:"schedule"`
	}
	Order struct {
		Side     string `yaml:"side"`
		Major    string `yaml:"major"`
		Minor    string `yaml:"minor"`
		Symbol   string `yaml:"symbol"`
		Quantity string `yaml:"quantity"`
	}
	Version string `yaml:"version"`
}

func DecodeConfig(config *Config) {
	configFile, err := os.Open("application.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			utils.Logger.Error("Unable to open application yaml config file, ", err)
		}
	}(configFile)

	yamlParser := yaml.NewDecoder(configFile)
	if err := yamlParser.Decode(&config); err != nil {
		utils.Logger.Panic("Unable to decode config from application.yaml, ", err)
	}
}
