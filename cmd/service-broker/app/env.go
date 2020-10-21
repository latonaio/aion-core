package app

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"flag"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path"
)

type Config struct {
	env  EnvironmentValue
	flag FlagValue
}

type EnvironmentValue struct {
	AionHome         string `envconfig:"AION_HOME" default:"/var/lib/aion"`
	RedisAddr        string `envconfig:"REDIS_HOST" default:"localhost:6379"`
	RepositoryPrefix string `envconfig:"REPOSITORY_PREFIX" default:"latonaio"`
	Namespace        string `envconfig:"NAMESPACE" default:"default"`
	RegistrySecret   string `envconfig:"REGISTRY_SECRET" default:"dockerhub"`
}

type FlagValue struct {
	ConfigPath *string
	IsDocker   *bool
}

func GetConfig() *Config {
	conf := Config{}
	// load environment variable
	if err := envconfig.Process("", &conf.env); err != nil {
		log.Fatalf("cant load environment variable: %v", err)
	}

	conf.flag.ConfigPath = flag.String("c", "./yaml/sample.yml", "AION Config File")
	conf.flag.IsDocker = flag.Bool("d", false, "use docker mode")
	flag.Parse()

	// check aion home
	if _, err := os.Stat(conf.env.AionHome); os.IsNotExist(err) {
		log.Fatalf("AION_HOME does not exist : %s", conf.env.AionHome)
	}
	return &conf
}

func (e *Config) GetConfigPath() string {
	return *e.flag.ConfigPath
}

func (e *Config) GetAionHome() string {
	return e.env.AionHome
}

func (e *Config) GetDataDir() string {
	return path.Join(e.env.AionHome, "Data")
}

func (e *Config) GetRedisAddr() string {
	return e.env.RedisAddr
}

func (e *Config) GetRepositoryPrefix() string {
	return e.env.RepositoryPrefix
}

func (e *Config) GetNamespace() string {
	return e.env.Namespace
}

func (e *Config) GetRegistrySecret() string {
	return e.env.RegistrySecret
}

func (e *Config) IsDocker() bool {
	return *e.flag.IsDocker
}
