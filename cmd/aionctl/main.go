package main

import (
	"os"

	"bitbucket.org/latonaio/aion-core/cmd/aionctl/app"
)

func main() {
	if err := app.RootCmd.Execute(); err != nil {
		os.Exit(0)
	}
}
