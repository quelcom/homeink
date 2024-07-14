package config

import (
	"log/slog"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Port           string
	KindleIP       string
	KindlePassword string
	PathToFBInk    string
}

var (
	conf        Config
	once        sync.Once
	alreadyInit bool
)

func GetConfig() Config {
	if alreadyInit {
		return conf
	}

	once.Do(func() {
		slog.Info("Config: init")
		_, err := toml.DecodeFile("config.toml", &conf)
		if err != nil {
			panic(err)
		}
		alreadyInit = true
	})

	return conf
}

func EmbedTestConfig(testConfig string) {
	if _, err := toml.Decode(testConfig, &conf); err != nil {
		panic("Could not decode test configuration. Error: " + err.Error())
	}
	alreadyInit = true
}
