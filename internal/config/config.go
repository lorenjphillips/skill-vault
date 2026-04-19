package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ToolConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Categories []string `yaml:"categories"`
}

type GitHubConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Repo      string `yaml:"repo"`
	LocalPath string `yaml:"local_path"`
}

type S3Config struct {
	Enabled bool   `yaml:"enabled"`
	Bucket  string `yaml:"bucket"`
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
}

type ScheduleConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}

type Config struct {
	Tools    map[string]ToolConfig `yaml:"tools"`
	GitHub   GitHubConfig          `yaml:"github"`
	S3       S3Config              `yaml:"s3"`
	Schedule ScheduleConfig        `yaml:"schedule"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skill-vault")
}

func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	os.MkdirAll(Dir(), 0755)
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0644)
}

func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}
