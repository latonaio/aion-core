package app

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path"
)

type Env struct {
	AionHome   string `envconfig:"AION_HOME" default:"/var/lib/aion"` // aion home path
	ServerPort int    `envconfig:"SERVER_PORT" default:"11011"`       // port of send anything
	ClientPort int    `envconfig:"CLIENT_PORT" default:"30100"`       // port of another client
}

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

func (e *Env) GetAionHome() string {
	return e.AionHome
}

func (e *Env) GetDataDir() string {
	return path.Join(e.AionHome, "Data")
}

func (e *Env) GetServerPort() int {
	return e.ServerPort
}

func (e *Env) GetClientPort() int {
	return e.ClientPort
}
