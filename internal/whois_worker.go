package internal

import (
	"log"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type WhoisWorker struct {
	client       *WhoisClient
	domains      []string
	gaugeChannel *prometheus.GaugeVec
	gaugeDomain  *prometheus.GaugeVec
}

func NewWhoisWorker(applicationNamespace string, domains []string) *WhoisWorker {
	worker := new(WhoisWorker)
	worker.client = NewWhoisClient(applicationNamespace)
	worker.domains = domains

	labels := []string{"type"}
	worker.gaugeChannel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "worker_channel",
			Help:      "Gauge for size of query channel.",
		},
		labels,
	)
	prometheus.MustRegister(worker.gaugeChannel)

	labels = []string{"domain", "unit"}
	worker.gaugeDomain = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "worker_domain",
			Help:      "Gauge for days remaining before expiration for a domain.",
		},
		labels,
	)
	prometheus.MustRegister(worker.gaugeDomain)
	return worker
}

func (worker *WhoisWorker) DoWork() {
	// Construct a channel for running whois queries in parallel.
	queryChannel := make(chan WhoisResponse, len(worker.domains))

	// Have a metric to show how deep this buffer gets.
	go func() {
		for {
			time.Sleep(1 * time.Second)
			worker.gaugeChannel.WithLabelValues("query_channel").Set(float64(len(queryChannel)))
		}
	}()

	// Run the whois queries, capture how many days as a gauge per.
	for {
		for _, domain := range worker.domains {
			go worker.getWhoisResponse(domain, queryChannel)
		}
		for i := 0; i < len(worker.domains); i++ {
			resp := <-queryChannel
			if resp.hasExpiration {
				delta := -(time.Since(resp.expiration))
				yearsRemaining := math.Round((delta.Hours()/24/365)*100) / 100
				daysRemaining := math.Round((delta.Hours()/24)*100) / 100
				worker.gaugeDomain.WithLabelValues(resp.domain, "years").Set(yearsRemaining)
				worker.gaugeDomain.WithLabelValues(resp.domain, "days").Set(daysRemaining)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (worker *WhoisWorker) getWhoisResponse(domain string, channel chan WhoisResponse) {
	resp, err := worker.client.Query(domain)
	if err != nil {
		log.Println("error in query", domain, err.Error())
	} else {
		channel <- resp
	}
}
