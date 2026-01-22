package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the structure of config.json
type Config struct {
	App struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Debug       bool   `json:"debug"`
		Language    string `json:"language"`
	} `json:"app"`
	Server struct {
		Host                string `json:"host"`
		Port                int    `json:"port"`
		ReadTimeoutSeconds  int    `json:"read_timeout_seconds"`
		WriteTimeoutSeconds int    `json:"write_timeout_seconds"`
	} `json:"server"`
	Database struct {
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
	} `json:"database"`
}

// Load reads config.json from path or from current directory if empty.
func Load(path string) (*Config, error) {
	if path == "" {
		path = "config.json"
	}
	// if relative path, make absolute relative to cwd
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err == nil {
			path = filepath.Join(wd, path)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// DatabaseURL builds a DSN for supported databases (postgres only for now).
// Returns empty string if unsupported or on error.
func (c *Config) DatabaseURL() string {
	if c == nil {
		return ""
	}
	switch c.Database.Type {
	case "postgres", "pg", "postgresql":
		// postgres://user:pass@host:port/dbname?sslmode=disable
		user := c.Database.User
		pass := c.Database.Password
		host := c.Database.Host
		port := c.Database.Port
		db := c.Database.DBName
		ssl := c.Database.SSLMode
		if ssl == "" {
			ssl = "disable"
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, pass, host, port, db, ssl)
	default:
		return ""
	}
}
