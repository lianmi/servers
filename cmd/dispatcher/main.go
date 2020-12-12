package main

import (
	"flag"
	"log"
)

var configFile = flag.String("f", "dispatcher.yml", "set config file which viper will loading.")

func main() {
	flag.Parse()

	app, err := CreateApp(*configFile)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	if err := app.Start(); err != nil {
		log.Println(err)
		panic(err)
	}

	app.AwaitSignal()
}
