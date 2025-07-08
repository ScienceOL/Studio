package main

import (
	"os"

	"github.com/scienceol/studio/service/cmd/api/app"
)

// @title Studio API
// @version 1.0
// @description Studio 实验室管理系统 API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	app := app.NewWeb()
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
