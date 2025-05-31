package main

import (
	"flag"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
)

func main() {
	// Get the config file from the command line
	cfgFileNamePtr := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Read Config

	config, err := config.ParseConfig(*cfgFileNamePtr)
	if err != nil {
		panic(err)
	}
	// Print Config
	logger := log.NewCLILogger("main", false)

	logger.Info("Config loaded successfully %+v", config)
}
