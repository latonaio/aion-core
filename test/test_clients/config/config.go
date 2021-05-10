package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	WatchPeriod int `envconfig:"WATCH_PERIOD" default:"15"`
	MaxAlertNum int `envconfig:"MAX_ALERT_NUM" default:"3"`
}


func New() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
