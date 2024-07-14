package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	WebsiteUrl         string
	ScreenshotDir      string
	ScreenshotInterval time.Duration
	UseRemoteAllocator bool
	RemoteURL          string
	SupersaaScreenshot []SupersaaScreenshot
}

type SupersaaScreenshot struct {
	Filename  string
	Selection string
	Actions   []string
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
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}

		binPath := filepath.Dir(ex)

		_, err = toml.DecodeFile(binPath+"/config.toml", &conf)
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
