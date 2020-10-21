package app

/*
	get environment variable in kanban server
*/

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path"
)

type Env struct {
	AionHome   string `envconfig:"AION_HOME" default:"/var/lib/aion"`             // aion home path
	RedisAddr  string `envconfig:"REDIS_HOST" default:"localhost:6379"`           // redis address
	ServerPort int    `envconfig:"SERVER_PORT" default:"11010"`                   // port of kanban server
	SBAddr     string `envconfig:"SERVICE_BROKER_HOST" default:"localhost:11001"` // address of service broker
}

// get environment variable and check existence of aion home
func GetConfig() *Env {
	var env Env
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("cant get env yaml: %v", err)
	}
	// check aion home
	if _, err := os.Stat(env.AionHome); os.IsNotExist(err) {
		log.Fatalf("AION_HOME does not exist : %s", env.AionHome)
	}
	return &env
}

// return aion home path
func (e *Env) GetAionHome() string {
	return e.AionHome
}

// return data path
func (e *Env) GetDataDir() string {
	return path.Join(e.AionHome, "Data")
}

// return service broker address
func (e *Env) GetSBAddr() string {
	return e.SBAddr
}

// return redis address
func (e *Env) GetRedisAddr() string {
	return e.RedisAddr
}

// return port of kanban server
func (e *Env) GetServerPort() int {
	return e.ServerPort
}
