package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	Admin    AdminConfig    `yaml:"admin"`
	App      AppConfig      `yaml:"app"`
}

type ServerConfig struct {
	Mode string `yaml:"mode"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type JWTConfig struct {
	Secret              string `yaml:"secret"`
	AccessExpireMinutes int    `yaml:"access_expire_minutes"`
	RefreshExpireHours  int    `yaml:"refresh_expire_hours"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	Filename string `yaml:"filename"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	RealName string `yaml:"real_name"`
}

type AppConfig struct {
	DemoMode bool `yaml:"demo_mode"`
}

func Load(path string) (*Config, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return nil, err
	}
	applyDefault(&cfg)
	return &cfg, nil
}

func applyDefault(cfg *Config) {
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "data/erp.db?_foreign_keys=on&_busy_timeout=5000"
	}
	if cfg.JWT.AccessExpireMinutes <= 0 {
		cfg.JWT.AccessExpireMinutes = 30
	}
	if cfg.JWT.RefreshExpireHours <= 0 {
		cfg.JWT.RefreshExpireHours = 24
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-this-jwt-secret-before-use"
	}
	if cfg.Admin.Username == "" {
		cfg.Admin.Username = "admin"
	}
	if cfg.Admin.Password == "" {
		cfg.Admin.Password = "admin888"
	}
}
