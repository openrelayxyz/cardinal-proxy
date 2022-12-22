package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	log "github.com/inconshreveable/log15"
)

type rpcService struct {
	Name        string            `yaml:"name"`
	HTTPPort    int64             `yaml:"ports.http"`
	WSPort      int64             `yaml:"ports.ws"`
	HCPort      int64             `yaml:"ports.hc"`
	Concurrency int               `yaml:"concurrency"`
	Backends    map[string]string `yaml:"backend.rpcmap"`
	BackendURLs map[string]string `yaml:"backend.urls"`
	DefaultBackendURL string      `yaml:"backend.default"`
}

type RPCService struct {
	rpcService
	unmarshal func(interface{}) error
}

func (msg *RPCService) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return unmarshal(&msg.rpcService)
}

func (msg *RPCService) Unmarshal(v any) error {
		return msg.unmarshal(v)
}

type Config struct {
	Services []*RPCService `yaml:"services"`
	PprofPort int `yaml:"pprof.port"`
	LogLevel string `yaml:"loggingLevel"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	var logLvl log.Lvl
	switch config.LogLevel {
	case "debug":
		logLvl = log.LvlDebug
	case "info":
		logLvl = log.LvlInfo
	case "warn":
		logLvl = log.LvlWarn
	case "error":
		logLvl = log.LvlError
	default:
		logLvl = log.LvlInfo
	}

	log.Root().SetHandler(log.LvlFilterHandler(logLvl, log.Root().GetHandler()))
	if len(config.Services) == 0 {
		return nil, errors.New("Must specify at least one service")
	}
	for _, service := range config.Services {
		if service.DefaultBackendURL == "" {
			return nil, errors.New("default.backend must be specified in config")
		}
		if service.Concurrency == 0 {
			service.Concurrency = 16
		}
	}
	return &config, nil
}
