package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultGPUMode = "auto"
)

type RedroidImage struct {
	Name      string
	URL       string
	Is64Only  bool
	IsARMOnly bool
}

var AvailableImages = []RedroidImage{
	{"Android 8.1", "redroid/redroid:8.1.0-latest", false, false},
	{"Android 9", "redroid/redroid:9.0.0-latest", false, false},
	{"Android 10", "redroid/redroid:10.0.0-latest", false, false},
	{"Android 11", "redroid/redroid:11.0.0-latest", false, false},
	{"Android 11 (64bit only)", "redroid/redroid:11.0.0_64only-latest", true, false},
	{"Android 11 (ARM64 only)", "abing7k/redroid:a11_arm", false, true},
	{"Android 11 (Magisk - ARM64)", "abing7k/redroid:a11_magisk_arm", false, true},
	{"Android 11 (GApps - ARM64)", "abing7k/redroid:a11_gapps_arm", false, true},
	{"Android 11 (GApps & Magisk - ARM64)", "abing7k/redroid:a11_gapps_magisk_arm", false, true},
	{"Android 11 (LibNDK only - AMD64/x86_64)", "abing7k/redroid:a11_ndk_amd", true, false},
	{"Android 11 (Magisk & LibNDK - AMD64/x86_64)", "abing7k/redroid:a11_magisk_ndk_amd", true, false},
	{"Android 11 (GApps & LibNDK - AMD64/x86_64)", "abing7k/redroid:a11_gapps_ndk_amd", true, false},
	{"Android 11 (GApps & Magisk & LibNDK - AMD64/x86_64)", "abing7k/redroid:a11_gapps_magisk_ndk_amd", true, false},
	{"Android 11 (GApps & Libhoudini - AMD64/x86_64)", "teddynight/redroid:latest", true, false},
	{"Android 11 (NDK ChromeOS - AMD64/x86_64)", "erstt/redroid:11.0.0_ndk_ChromeOS", true, false},
	{"Android 12", "redroid/redroid:12.0.0-latest", false, false},
	{"Android 12 (64bit only)", "redroid/redroid:12.0.0_64only-latest", true, false},
	{"Android 12 (Fahaddz - GApps & Magisk)", "fahaddz/redroid:13", false, false},
	{"Android 12 (NDK ChromeOS - AMD64/x86_64)", "erstt/redroid:12.0.0_ndk_ChromeOS", true, false},
	{"Android 13", "redroid/redroid:13.0.0-latest", false, false},
	{"Android 13 (64bit only)", "redroid/redroid:13.0.0_64only-latest", true, false},
	{"Android 13 (NDK ChromeOS - AMD64/x86_64)", "erstt/redroid:13.0.0_ndk_ChromeOS", true, false},
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
		return nil, fmt.Errorf("Failed to read config: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Failed to parse config: %v", err)
	}

	if cfg.Containers == nil {
		cfg.Containers = make(map[string]*Container)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir := GetConfigDir()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("Failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	configPath := GetConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("Failed to write config: %v", err)
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

func Is64OnlyImage(imageURL string) bool {
	return strings.Contains(imageURL, "64only") ||
		strings.Contains(imageURL, "ndk_amd") ||
		strings.Contains(imageURL, "ndk_ChromeOS") ||
		strings.EqualFold(imageURL, "teddynight/redroid:latest")
}

func ExtractVersionFromImage(imageURL string) string {
	parts := strings.Split(imageURL, ":")
	if len(parts) < 2 {
		return ""
	}

	versionPart := parts[1]

	versionPart = strings.TrimSuffix(versionPart, "-latest")

	if strings.Contains(versionPart, "_") {
		return versionPart
	}

	knownVersions := []string{
		"8.1.0",
		"9.0.0",
		"10.0.0",
		"11.0.0",
		"a11_arm",
		"a11_magisk_arm",
		"a11_gapps_arm",
		"a11_gapps_magisk_arm",
		"a11_ndk_amd",
		"a11_magisk_ndk_amd",
		"a11_gapps_ndk_amd",
		"a11_gapps_magisk_ndk_amd",
		"11.0.0_ndk_ChromeOS",
		"12.0.0",
		"12.0.0_64only",
		"12.0.0_ndk_ChromeOS",
		"13.0.0",
		"13.0.0_64only",
		"13.0.0_ndk_ChromeOS",
	}

	for _, v := range knownVersions {
		if strings.HasPrefix(versionPart, v) {
			return v
		}
	}

	return versionPart
}
