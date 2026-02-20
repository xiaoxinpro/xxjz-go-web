package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration (config.yaml + env override).
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	App      AppConfig      `yaml:"app"`
	User     UserConfig     `yaml:"user"`
	Mail     MailConfig     `yaml:"mail"`
	Wechat   WechatConfig   `yaml:"wechat"`
	Money    MoneyConfig   `yaml:"money"`
	Limits   LimitsConfig  `yaml:"limits"`
	Image    ImageConfig   `yaml:"image"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Driver      string `yaml:"driver"`
	DSN         string `yaml:"dsn"`
	TablePrefix string `yaml:"table_prefix"`
}

type AppConfig struct {
	Title       string `yaml:"title"`
	Keywords    string `yaml:"keywords"`
	Description string `yaml:"description"`
	Welcome     string `yaml:"welcome"`
	Version     string `yaml:"version"`
}

type UserConfig struct {
	LoginTimes int         `yaml:"login_times"`
	PageSize   int         `yaml:"page_size"`
	AdminUID   int         `yaml:"admin_uid"`
	Demo       DemoAccount `yaml:"demo"`
}

type DemoAccount struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type MailConfig struct {
	Host     string `yaml:"host"`
	Secure   string `yaml:"secure"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
}

type WechatConfig struct {
	Enable   bool   `yaml:"enable"`
	OpenIDKey string `yaml:"openid_key"`
	Secret   string `yaml:"secret"`
}

type MoneyConfig struct {
	FormatDecimals  int     `yaml:"format_decimals"`
	FormatPoint     string  `yaml:"format_point"`
	FormatThousands string  `yaml:"format_thousands"`
	MaxValue        float64 `yaml:"max_value"`
}

type LimitsConfig struct {
	MaxClassName  int `yaml:"max_class_name"`
	MaxFundsName  int `yaml:"max_funds_name"`
	MaxMarkValue  int `yaml:"max_mark_value"`
}

type ImageConfig struct {
	MaxSize    int      `yaml:"max_size"`
	MaxCount   int      `yaml:"max_count"`
	AllowedExt []string `yaml:"allowed_ext"`
	RootPath   string  `yaml:"root_path"`
	CacheURL   string  `yaml:"cache_url"`
}

// Load reads config from path and applies environment variable overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	applyEnv(&c)
	return &c, nil
}

func applyEnv(c *Config) {
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Server.Port = p
		}
	}
	if v := os.Getenv("GIN_MODE"); v != "" {
		c.Server.Mode = v
	}
	if v := os.Getenv("DB_DRIVER"); v != "" {
		c.Database.Driver = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("DB_TABLE_PREFIX"); v != "" {
		c.Database.TablePrefix = v
	}
}
