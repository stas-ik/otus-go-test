package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	content := `
logger:
  level: info
server:
  host: 127.0.0.1
  port: 8080
  grpcHost: 127.0.0.1
  grpcPort: 50051
storage:
  type: sql
database:
  dsn: postgres://user:pass@localhost:5432/db
rabbitmq:
  url: amqp://guest:guest@localhost:5672/
  queue: events
schedule:
  scanInterval: 1m
  cleanupInterval: 1h
`
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := NewConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Logger.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logger.Level)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Storage.Type != "sql" {
		t.Errorf("expected type sql, got %s", cfg.Storage.Type)
	}
	if cfg.Schedule.ScanInterval != time.Minute {
		t.Errorf("expected scanInterval 1m, got %v", cfg.Schedule.ScanInterval)
	}
}

func TestNewConfig_FileNotFound(t *testing.T) {
	_, err := NewConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestNewConfig_InvalidYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("invalid: yaml: : content")); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = NewConfig(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
