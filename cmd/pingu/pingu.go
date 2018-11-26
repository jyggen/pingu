package main

import (
	"github.com/jyggen/pingu"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger := logrus.New()
	config := viper.New()

	config.AutomaticEnv()
	config.SetConfigName("pingu")
	config.AddConfigPath(".")

	if err := config.ReadInConfig(); err == nil {
		logger.WithField("file", config.ConfigFileUsed()).Info("Configuration file loaded")
	}

	p := pingu.New(config, logger)

	p.Run()
}
