package config

import (
	"bytes"
	"io"
	"log/slog"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Global     *GlobalConfig          `yaml:"global"`
	Server     *ServerType            `yaml:"server"`
	Clients    []*ClientConfig        `yaml:"clients"`
	Auths      []*AuthConfig          `yaml:"auths"`
	Collectors map[string]interface{} `yaml:"collectors"`

	loaded  bool
	success bool
	logger  *slog.Logger
}

func NewConfiguration() *Configuration {
	cfg := &Configuration{
		Global: &GlobalConfig{
			Server: &ServerConfig{
				Endpoint: new(string),
				Api_path: new(string),
				Insecure: new(bool),
				Mode:     new(string),
			},
			Client: &ClientConfig{
				Endpoint: new(string),
				Interval: new(time.Duration),
				Auth:     new(string),
				Insecure: new(bool),
			},
		},
		Server: &ServerType{
			Metrics: &ServerConfig{
				Enabled: true,
			},
			Logs: &ServerConfig{
				Enabled: true,
			},
			Traces: &ServerConfig{
				Enabled: false,
			},
		},
		Clients:    []*ClientConfig{},
		Auths:      []*AuthConfig{},
		Collectors: map[string]interface{}{},
		success:    true,
	}
	cfg.init()
	return cfg
}

func (_cfg *Configuration) init() {
	// Global -> Server
	*_cfg.Global.Server.Endpoint = "http://127.0.0.1:9090"
	*_cfg.Global.Server.Api_path = ""
	*_cfg.Global.Server.Insecure = true
	*_cfg.Global.Server.Mode = "http"

	// Global -> Client
	*_cfg.Global.Client.Endpoint = "https://127.0.0.1:8080"
	*_cfg.Global.Client.Interval = 1 * time.Second
	*_cfg.Global.Client.Insecure = true
}

func (_cfg *Configuration) LoadFile(filepath string, logger *slog.Logger) error {
	_cfg.logger = logger
	logger.Debug("loading config file", "path", filepath)
	f, err := os.Open(filepath)
	if err != nil {
		_cfg.success = false
		return err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		_cfg.success = false
		return err
	}
	err = yaml.NewDecoder(bytes.NewReader(content)).Decode(_cfg)
	if err != nil {
		_cfg.success = false
		return err
	}

	_cfg.applyGlobal()
	_cfg.checkDefects()
	_cfg.loaded = true

	return nil
}

func (_cfg *Configuration) applyGlobal() {

	_cfg.Server.Metrics.applyGlobal(_cfg.Global.Server)
	_cfg.Server.Logs.applyGlobal(_cfg.Global.Server)
	_cfg.Server.Traces.applyGlobal(_cfg.Global.Server)

	for _, client := range _cfg.Clients {
		client.applyGlobal(_cfg.Global.Client)
	}

}

func (_cfg *Configuration) checkDefects() {
	// Server Check
	if _cfg.Server.Metrics.Enabled {
		_, err := url.Parse(*_cfg.Server.Metrics.Endpoint)
		if err != nil {
			_cfg.success = false
			_cfg.logger.Error("failed to parse metrics endpoint", "error", err)
		}
	}
	if _cfg.Server.Logs.Enabled {
		_, err := url.Parse(*_cfg.Server.Logs.Endpoint)
		if err != nil {
			_cfg.success = false
			_cfg.logger.Error("failed to parse logs endpoint", "error", err)
		}
	}
	if _cfg.Server.Traces.Enabled {
		_, err := url.Parse(*_cfg.Server.Traces.Endpoint)
		if err != nil {
			_cfg.success = false
			_cfg.logger.Error("failed to parse traces endpoint", "error", err)
		}
	}

	// Client Check
	var endpointErr int
	for _, client := range _cfg.Clients {
		// Check Endpoint
		if client.Endpoint == nil {
			endpointErr++
		} else {
			// Url Check
			_, err := url.Parse(*client.Endpoint)
			if err != nil {
				endpointErr++
			}
		}
		var found bool

		// Check Auth
		for _, auth := range _cfg.Auths {
			if auth.Name == *client.Auth {
				found = true
			}
		}
		if !found {
			_cfg.logger.Error("auth not found", "auth", *client.Auth)
		}
	}
	if endpointErr > 0 {
		_cfg.logger.Error("invalid the endpoint of client", "error_count", endpointErr)
		_cfg.success = false
	}

	// Auth Check
	for _, auth := range _cfg.Auths {
		if auth.Name == "" {
			_cfg.success = false
			_cfg.logger.Error("auth name is not set")
		}
		if auth.Username == "" || auth.Password == "" {
			_cfg.success = false
			_cfg.logger.Error("auth username or password is not set", "auth", auth.Name)
		}
	}
}

type GlobalConfig struct {
	Server *ServerConfig `yaml:"server"`
	Client *ClientConfig `yaml:"client"`
}

type ServerConfig struct {
	Enabled  bool    `yaml:"enabled"`
	Endpoint *string `yaml:"endpoint"`
	Api_path *string `yaml:"api_path"`
	Insecure *bool   `yaml:"insecure"`
	Mode     *string `yaml:"mode"`
}

type ClientConfig struct {
	Endpoint *string           `yaml:"endpoint"`
	Auth     *string           `yaml:"auth"`
	Interval *time.Duration    `yaml:"interval"`
	Insecure *bool             `yaml:"insecure"`
	Labels   map[string]string `yaml:"labels"`
}

type AuthConfig struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ServerType struct {
	Metrics *ServerConfig `yaml:"metrics"`
	Logs    *ServerConfig `yaml:"logs"`
	Traces  *ServerConfig `yaml:"traces"`
}

func (_cfg *ServerConfig) applyGlobal(global *ServerConfig) {
	if _cfg.Endpoint == nil {
		_cfg.Endpoint = global.Endpoint
	}
	if _cfg.Api_path == nil {
		_cfg.Api_path = global.Api_path
	}
	if _cfg.Insecure == nil {
		_cfg.Insecure = global.Insecure
	}
	if _cfg.Mode == nil {
		_cfg.Mode = global.Mode
	}
}

func (_cfg *ClientConfig) applyGlobal(global *ClientConfig) {
	if _cfg.Auth == nil {
		_cfg.Auth = global.Auth
	}
	if _cfg.Interval == nil {
		_cfg.Interval = global.Interval
	}
	for gk, gv := range global.Labels {
		var found = false
		for ck, _ := range global.Labels {
			if gk == ck {
				found = true
			}
		}
		if !found {
			_cfg.Labels[gk] = gv
		}
	}
}
