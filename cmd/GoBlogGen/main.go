package main

import (
	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
)

func main() {
	// Read Config

	filename := "config.yaml"
	config, err := config.ParseConfig(filename)
	if err != nil {
		panic(err)
	}
	// Print Config
	logger := log.NewCLILogger("main", false)

	logger.Info("Config loaded successfully %+v", config)
}
