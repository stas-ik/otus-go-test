package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger   LoggerConf   `yaml:"logger"`
	Server   ServerConf   `yaml:"server"`
	Storage  StorageConf  `yaml:"storage"`
	Database DatabaseConf `yaml:"database"`
	RabbitMQ RabbitMQConf `yaml:"rabbitmq"`
	Schedule ScheduleConf `yaml:"schedule"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type ServerConf struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	GRPCHost string `yaml:"grpcHost"`
	GRPCPort string `yaml:"grpcPort"`
}

type StorageConf struct {
	Type string `yaml:"type"` // "memory" or "sql"
}

type DatabaseConf struct {
	DSN string `yaml:"dsn"`
}

type RabbitMQConf struct {
	URL   string `yaml:"url"`
	Queue string `yaml:"queue"`
}

type ScheduleConf struct {
	ScanInterval    time.Duration `yaml:"scanInterval"`
	CleanupInterval time.Duration `yaml:"cleanupInterval"`
}

func NewConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}
