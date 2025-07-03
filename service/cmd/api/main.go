package main

import (
	"os"

	"github.com/scienceol/studio/service/cmd/api/app"
)

func main() {
	app := app.NewWeb()
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
