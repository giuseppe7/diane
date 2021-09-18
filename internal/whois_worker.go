package internal

import (
	"log"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type WhoisWorker struct {
	client            *WhoisClient
	domains           []string
	gaugeChannel      *prometheus.GaugeVec
	gaugeDomainExpiry *prometheus.GaugeVec
}

func NewWhoisWorker(applicationNamespace string, domains []string) *WhoisWorker {
	worker := new(WhoisWorker)
	worker.client = NewWhoisClient(applicationNamespace)
	worker.domains = domains

	labels := []string{"type"}
	worker.gaugeChannel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "whois_worker_channel",
			Help:      "Gauge for size of query channel.",
		},
		labels,
	)
	prometheus.MustRegister(worker.gaugeChannel)

	labels = []string{"domain", "unit"}
	worker.gaugeDomainExpiry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "whois_worker_domain_expiry",
			Help:      "Gauge for days remaining before expiration for a domain.",
		},
		labels,
	)
	prometheus.MustRegister(worker.gaugeDomainExpiry)
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
		worker.queryDomains(queryChannel)
		for i := 0; i < len(worker.domains); i++ {
			resp := <-queryChannel
			//status := "unknown"
			if resp.hasExpiration {
				delta := -(time.Since(resp.expiration))
				yearsRemaining := math.Round((delta.Hours()/24/365)*100) / 100
				daysRemaining := math.Round((delta.Hours()/24)*100) / 100
				worker.gaugeDomainExpiry.WithLabelValues(resp.domain, "years").Set(yearsRemaining)
				worker.gaugeDomainExpiry.WithLabelValues(resp.domain, "days").Set(daysRemaining)
				log.Printf("Queried %v, it expires in %d days!\n", resp.target, int(daysRemaining))
			} else {
				log.Printf("Queried %v, status is %v", resp.target, resp.status.String())
			}
		}

		// TODO: Make this a configuration setting?
		pollingIntervalInMinutes := 5
		time.Sleep(time.Duration(pollingIntervalInMinutes) * time.Minute)
	}
}

func (worker *WhoisWorker) queryDomains(queryChannel chan WhoisResponse) {
	for _, domain := range worker.domains {
		go worker.getWhoisResponse(domain, queryChannel)
	}
}

func (worker *WhoisWorker) getWhoisResponse(target string, channel chan WhoisResponse) {
	resp := worker.client.Query(target)
	if resp.err != nil {
		log.Println("Error in query", target, resp.err.Error())
	}
	channel <- resp
}
