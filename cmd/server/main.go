package main

import (
	"github.com/practice/multi_resource/cmd/server/app"
	"os"
)

func main() {
	cmd := app.NewServerCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
