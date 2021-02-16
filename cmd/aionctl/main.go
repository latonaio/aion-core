package main

import (
	"fmt"
	"os"

	"bitbucket.org/latonaio/aion-core/cmd/aionctl/app"
)

func main() {
	if err := app.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(-1)
	}
}
