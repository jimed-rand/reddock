package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultGPUMode = "guest"
)

type RedroidImage struct {
	Name string
	URL  string
}

var AvailableImages = []RedroidImage{
	{"Android 16", "redroid/redroid:16.0.0-latest"},
	{"Android 16 64bit only", "redroid/redroid:16.0.0_64only-latest"},
	{"Android 15", "redroid/redroid:15.0.0-latest"},
	{"Android 15 64bit only", "redroid/redroid:15.0.0_64only-latest"},
	{"Android 14", "redroid/redroid:14.0.0-latest"},
	{"Android 14 64bit only", "redroid/redroid:14.0.0_64only-latest"},
	{"Android 13", "redroid/redroid:13.0.0-latest"},
	{"Android 13 64bit only", "redroid/redroid:13.0.0_64only-latest"},
	{"Android 12", "redroid/redroid:12.0.0-latest"},
	{"Android 12 64bit only", "redroid/redroid:12.0.0_64only-latest"},
	{"Android 11", "redroid/redroid:11.0.0-latest"},
	{"Android 10", "redroid/redroid:10.0.0-latest"},
	{"Android 9", "redroid/redroid:9.0.0-latest"},
	{"Android 8.1", "redroid/redroid:8.1.0-latest"},
}

type Container struct {
	Name        string `json:"name"`
	ImageURL    string `json:"image_url"`
	DataPath    string `json:"data_path"`
	LogFile     string `json:"log_file"`
	Port        int    `json:"port"`
	GPUMode     string `json:"gpu_mode"`
	Initialized bool   `json:"initialized"`
}

type Config struct {
	Containers map[string]*Container `json:"containers"`
}

func GetConfigDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "reddock")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func GetDefaultDataPath(containerName string) string {
	home := os.Getenv("HOME")
	return filepath.Join(home, "data-"+containerName)
}

func GetDefault() *Config {
	return &Config{
		Containers: make(map[string]*Container),
	}
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefault(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	if cfg.Containers == nil {
		cfg.Containers = make(map[string]*Container)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir := GetConfigDir()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := GetConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

func (c *Container) GetDataPath() string {
	if c.DataPath != "" {
		return c.DataPath
	}
	return GetDefaultDataPath(c.Name)
}

func (cfg *Config) GetContainer(name string) *Container {
	if container, exists := cfg.Containers[name]; exists {
		return container
	}
	return nil
}

func (cfg *Config) AddContainer(container *Container) {
	if cfg.Containers == nil {
		cfg.Containers = make(map[string]*Container)
	}
	cfg.Containers[container.Name] = container
}

func (cfg *Config) RemoveContainer(name string) {
	delete(cfg.Containers, name)
}

func (cfg *Config) ListContainers() []*Container {
	var containers []*Container
	for _, container := range cfg.Containers {
		containers = append(containers, container)
	}
	return containers
}
