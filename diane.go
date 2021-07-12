package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/giuseppe7/diane/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

const applicationNamespace = "diane"

// Variable to be set by the Go linker at build time.
var version string

// Structure for parsing yaml configuration.
type configuration struct {
	Domains []string `yaml:"domains"`
}

// Parse command line flags to determine config location and data.
func initConfiguration() configuration {
	var configFile string
	defaultConfigFile := fmt.Sprintf("./configs/%s.yaml", applicationNamespace)
	flag.StringVar(&configFile, "config", defaultConfigFile, "path to configuration file")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Configuration file error.", err)
	}

	var c configuration
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatal("Configuration file unmarshal error.", err)
	}
	return c
}

// Set up observability with Prometheus handler for metrics.
func initObservability() {

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	// Register a version gauge.
	versionGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "version_info",
			Help:      "Version of the application.",
		},
	)
	prometheus.MustRegister(versionGauge)
	versionValue, err := strconv.ParseFloat(version, 64)
	if err != nil {
		versionValue = 0.0
	}
	versionGauge.Set(versionValue)
}

// Obvious main function for the application.
func main() {
	log.Println("Coming online...")
	log.Print(fmt.Sprintf("Version: %v\n", version))

	// Channel to be aware of an OS interrupt like Control-C.
	var waiter sync.WaitGroup
	waiter.Add(1)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Load up configuration.
	appConfig := initConfiguration()
	log.Println(fmt.Sprintf("Loaded %d domains from configuration.", len(appConfig.Domains)))

	// Set up observability.
	initObservability()

	// Do the work.
	whoisWorker := internal.NewWhoisWorker(applicationNamespace, appConfig.Domains)
	go whoisWorker.DoWork()

	// Function and waiter to wait for the OS interrupt and do any clean-up.
	go func() {
		<-c
		fmt.Println("\r")
		log.Println("Interrupt captured.")
		waiter.Done()
	}()
	waiter.Wait()

	// Shut down the application.
	log.Println("Shutting down.")
}
