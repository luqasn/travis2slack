package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Filters map[string]string

type Template struct {
	Message string // `yaml:"message"`
}

type Slack struct {
	OAuthAccessToken string
}

type Travis struct {
	PublicKeyURL        string
	DisableVerification bool
}

type Templates map[string]Template

type HTTP struct {
	ListenAddress string
}

type Config struct {
	Slack           Slack
	Travis          Travis
	Templates       Templates
	Filters         Filters
	DefaultTemplate string
	DefaultFilter   string
	HTTP            HTTP
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("cfg")
	viper.AutomaticEnv()

	viper.SetDefault("Travis.PublicKeyURL", "https://api.travis-ci.com/config")
	viper.SetDefault("HTTP.ListenAddress", ":8080")
	var configuration Config

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	return configuration
}
