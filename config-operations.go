package main

import (
	"os"

	stanlog "github.com/Biloute271/stan-log"
	"gopkg.in/yaml.v2"
)

const (
	localConfigFile = "config.yaml"
)

type Config struct {
	App struct {
		LogLevel string `yaml:"loglevel"`
	} `yaml:"app"`
	Clickhouse struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		Login    string `yaml:"login"`
		Password string `yaml:"password"`
	} `yaml:"clickhouse"`
}

func readConfig() error {
	stanlog.Info("Reading configuration file")
	// Read configuration
	f, err := os.Open(localConfigFile)
	if err != nil {
		stanlog.Error("Error reading configuration file")
		return err
	}
	defer f.Close()
	stanlog.Info("Decoding configuration file")
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		stanlog.Error("Error decoding config file: " + err.Error())
		return err
	}
	stanlog.Info("Configuration file successfully read and decoded")
	stanlog.SetLogLevel(config.App.LogLevel)
	stanlog.Debug("Clickhouse server : " + config.Clickhouse.Server + ":" + config.Clickhouse.Port)
	return nil
}
