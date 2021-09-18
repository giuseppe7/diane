package internal

import (
	"log"

	"github.com/spf13/viper"
)

const ApplicationNamespace = "diane"
const ApplicationMetricsEndpoint = "/metrics"
const ApplicationMetricsEndpointPort = ":2112"

// Structure for parsed yaml configuration.
type configuration struct {
	Domains []string `yaml:"domains"`
}

func InitConfiguration() configuration {
	viper.SetConfigName("diane")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/diane/")
	viper.AddConfigPath("./configs/")
	viper.AddConfigPath("../configs/")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("could not read the configuration file", err)
	}
	var c configuration
	err = viper.Unmarshal(&c)
	if err != nil {
		log.Fatal("could not unmarshal the configuration file", err)
	}
	return c
}
