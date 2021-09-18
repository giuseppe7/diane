package internal

import (
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type WhoisClient struct {
	histogram *prometheus.HistogramVec
}

func NewWhoisClient(applicationNamespace string) *WhoisClient {
	whoisClient := new(WhoisClient)

	// Capture metrics on the command execution times.
	whoisClient.histogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: applicationNamespace,
			Name:      "whois_client_command_duration_seconds",
			Help:      "Histogram of client calls in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 20, 30},
		},
		[]string{"target", "refer", "status"},
	)
	prometheus.MustRegister(whoisClient.histogram)

	return whoisClient
}

// Performs the whois query via port 43 protocol and returns a simplied single
// response intentionally because I'm a jerk and this is not meant to be
// exhaustive.
func (w *WhoisClient) Query(target string) WhoisResponse {
	hostPort := "whois.iana.org:43" // Default search before referrals.

	referral := "whois.iana.org" // No referral, then its iana.
	start := time.Now()
	whoisResponse := w.sendRequest(hostPort, target)
	if whoisResponse.err == nil {
		// No issues, check if a referral is sent.
		if whoisResponse.refer != "" {
			// Referral found, second invocation.
			referral = whoisResponse.refer
			referHostPort := fmt.Sprintf("%s:43", whoisResponse.refer)
			whoisResponse = w.sendRequest(referHostPort, target)
		}
	}
	duration := time.Since(start)
	w.histogram.WithLabelValues(target, referral, whoisResponse.status.String()).Observe(duration.Seconds())
	return whoisResponse
}

func (w *WhoisClient) sendRequest(hostPort string, target string) WhoisResponse {
	var resp WhoisResponse
	resp.target = target
	resp.hostPort = hostPort

	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.Dial("tcp", hostPort) // Typically host:43
	if err != nil {
		resp.status = ResponseError
		resp.err = err
		return resp
	}
	defer conn.Close()

	conn.Write([]byte(target + "\r\n"))
	buf := make([]byte, 1024)
	result := []byte{}
	for {
		numBytes, err := conn.Read(buf)
		sbuf := buf[0:numBytes]
		result = append(result, sbuf...)
		if err != nil {
			break
		}
	}
	if err != nil {
		resp.status = ResponseError
		resp.err = err
		return resp
	}

	resp.ParseRawResponse(string(result))
	return resp
}
