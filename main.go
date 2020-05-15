package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xdhuxc/kubernetes-transform/src/api"
	"github.com/xdhuxc/kubernetes-transform/src/config"
)

var cf = flag.String("config", "config.prod.yaml", "config path")

func main() {
	flag.Parse()
	err := config.InitConfig(*cf)
	if err != nil {
		_ = fmt.Errorf("init config error: %v", err)
		os.Exit(1)
	}

	app := api.NewRouter()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
