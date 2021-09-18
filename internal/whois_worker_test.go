package internal

import (
	"log"
	"math"
	"testing"
	"time"
)

func TestWhoisWorkerCheckDomains(t *testing.T) {
	appConfig := InitConfiguration()
	if len(appConfig.Domains) < 4 {
		t.Errorf("expected at least four domains in configuration, found %d", len(appConfig.Domains))
	}

	whoisWorker := NewWhoisWorker(ApplicationNamespace, appConfig.Domains)
	if len(whoisWorker.domains) < 4 {
		t.Errorf("expected at least four domains in configuration, found %d", len(appConfig.Domains))
	}

	queryChannel := make(chan WhoisResponse, len(whoisWorker.domains))
	whoisWorker.queryDomains(queryChannel)
	for i := 0; i < len(whoisWorker.domains); i++ {
		resp := <-queryChannel
		if resp.status == ResponseAvailable {
			log.Printf("queried %v, status is available", resp.target)
		} else if resp.status == ResponseError {
			log.Printf("queried %v, status is error, %v", resp.target, resp.status.String())
		} else if resp.status == ResponseExceededRate {
			log.Printf("queried %v, exceeded rate with %v", resp.target, resp.refer)
		} else if resp.status == ResponseOk {
			if resp.hasExpiration {
				delta := -(time.Since(resp.expiration))
				daysRemaining := math.Round((delta.Hours()/24)*100) / 100
				log.Printf("queried %v, expires in %v days", resp.target, daysRemaining)
			} else {
				log.Printf("queried %v, status is ok", resp.target)
			}
		} else if resp.status == ResponseUnauthorized {
			log.Printf("queried %v, unauthorized with %v", resp.target, resp.hostPort)
		} else if resp.status == ResponseUnknown {
			log.Printf("queried %v, unknown with %v", resp.target, resp.refer)
		} else {
			log.Printf("queried %v, unexpected status is %v", resp.target, resp.status.String())
		}

		if resp.status == ResponseUnknown {
			t.Errorf("queried %v, not expecting status %v", resp.target, ResponseUnknown)
		}
	}
}
