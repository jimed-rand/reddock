package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultImageURL uses OCI registry format for LXC OCI template
	// Format: oci-registry://registry/image:tag (accessed via skopeo)
	DefaultImageURL      = "docker://redroid/redroid:16.0.0_64only-latest"
	DefaultContainerName = "redroid"
	DefaultGPUMode       = "guest"
)

type Config struct {
	ContainerName string `json:"container_name"`
	ImageURL      string `json:"image_url"`
	DataPath      string `json:"data_path"`
	LogFile       string `json:"log_file"`
	GPUMode       string `json:"gpu_mode"`
	Initialized   bool   `json:"initialized"`
}

func GetConfigDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "redway")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func GetDefaultDataPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, "data-redroid")
}

func GetDefault() *Config {
	return &Config{
		ContainerName: DefaultContainerName,
		ImageURL:      DefaultImageURL,
		DataPath:      GetDefaultDataPath(),
		LogFile:       "redroid.log",
		GPUMode:       DefaultGPUMode,
		Initialized:   false,
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

func (c *Config) GetContainerPath() string {
	return filepath.Join("/var/lib/lxc", c.ContainerName)
}

func (c *Config) GetConfigFilePath() string {
	return filepath.Join(c.GetContainerPath(), "config")
}

func (c *Config) GetRootfsPath() string {
	return filepath.Join(c.GetContainerPath(), "rootfs")
}
