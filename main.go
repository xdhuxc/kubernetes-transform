package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"
	"k8s.io/klog/v2"

	"github.com/xdhuxc/kubernetes-transform/src/api"
	"github.com/xdhuxc/kubernetes-transform/src/config"
)

var cf = flag.String("config", "config.prod.yaml", "config path")

func main() {
	flag.Parse()
	err := config.InitConfig(*cf)
	if err != nil {
		klog.Fatalln(err)
	}

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		for s := range channel {
			switch s {
			case syscall.SIGTERM:
				fmt.Println("Start to exit, sleep 5s")
				time.Sleep(5 * time.Second)
				os.Exit(0)
			default:
				fmt.Println("Receive other signal")
			}
		}
	}()

	app, err := api.NewRouter()
	if err != nil {
		klog.Fatalln(err)
	}
	if err := app.Run(); err != nil {
		klog.Fatalln(err)
	}
}
