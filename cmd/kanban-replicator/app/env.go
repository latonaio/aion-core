package app

/*
	get environment variable in kanban server
*/

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"flag"
	"github.com/kelseyhightower/envconfig"
	"os"
)

type Config struct {
	env        Env
	configPath *string
}

type Env struct {
	AionHome         string `envconfig:"AION_HOME" default:"/var/lib/aion"`    // aion home path
	RedisAddr        string `envconfig:"REDIS_HOST" default:"localhost:6379"`  // redis address
	MongoAddr        string `envconfig:"MONGO_HOST" default:"localhost:27017"` // mongo address
	KanbanDB         string `envconfig:"KANBAN_DB" default:"AionCore"`         // mongo db name
	KanbanCollection string `envconfig:"KANBAN_COLLECTION" default:"kanban"`   // mongo collection name
}

func GetConfig() *Config {
	conf := Config{}

	// load environment variable
	if err := envconfig.Process("", &conf.env); err != nil {
		log.Fatalf("cant load environment variable: %v", err)
	}

	conf.configPath = flag.String("c", "./yaml/sample.yml", "AION Config File")
	flag.Parse()

	// check aion home
	if _, err := os.Stat(conf.env.AionHome); os.IsNotExist(err) {
		log.Fatalf("AION_HOME does not exist : %s", conf.env.AionHome)
	}

	return &conf
}

func (c *Config) getAionHome() string {
	return c.env.AionHome
}

func (c *Config) getRedisAddr() string {
	return c.env.RedisAddr
}

func (c *Config) getMongoAddr() string {
	return c.env.MongoAddr
}

func (c *Config) getConfigPath() string {
	return *c.configPath
}

func (c *Config) getKanbanDB() string {
	return c.env.KanbanDB
}

func (c *Config) getKanbanCollection() string {
	return c.env.KanbanCollection
}
