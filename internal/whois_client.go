package internal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
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
		[]string{"domain", "refer", "status"},
	)
	prometheus.MustRegister(whoisClient.histogram)

	return whoisClient
}

// Performs the whois query via port 43 protocol and returns a simplied single
// response intentionally because I'm a jerk and this is not meant to be
// exhaustive.
func (w *WhoisClient) Query(target string) (WhoisResponse, error) {
	hostPort := "whois.iana.org:43" // Default search before referrals.

	referral := "none"
	status := "ok"
	start := time.Now()
	whoisResponse, err := w.sendRequest(hostPort, target)
	if err != nil {
		status = "error"
	} else {
		if whoisResponse.refer != "" {
			// Refer was specified, replace it with the final results.
			referral = whoisResponse.refer
			referHostPort := fmt.Sprintf("%s:43", whoisResponse.refer)
			whoisResponse, err = w.sendRequest(referHostPort, target)
			if err != nil {
				status = "error"
			}
		}
	}

	duration := time.Since(start)
	w.histogram.WithLabelValues(whoisResponse.domain, referral, status).Observe(duration.Seconds())
	return whoisResponse, err
}

func (w *WhoisClient) sendRequest(hostPort string, target string) (WhoisResponse, error) {
	var resp WhoisResponse

	conn, err := net.Dial("tcp", hostPort) // Typically host:43
	if err != nil {
		return resp, err
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

	resp, err = w.parseResponse(string(result))
	return resp, err
}

func (w *WhoisClient) parseResponse(raw string) (WhoisResponse, error) {
	resp := NewWhoisResponse()
	resp.raw = raw

	// Iterate through the response to find key words to attributes.
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue // Skip empty lines.
		}
		if noMatchFound(text) {
			resp.isAvailable = true
		}
		if hasRefer(text) {
			resp.refer = getRefer(text)
		}
		if hasDomain(text) {
			resp.domain = getDomain(text)
		}
		if hasExpiration(text) {
			resp.hasExpiration = true
			resp.expiration = getExpiration(text)
		}
	}

	return resp, nil
}

func noMatchFound(text string) bool {
	return strings.HasPrefix(text, "No match for")
}

func hasRefer(text string) bool {
	return strings.HasPrefix(text, "refer:")
}

func getRefer(text string) string {
	result := ""
	if hasRefer(text) {
		values := strings.Split(text, "refer:")
		result = strings.ToLower(strings.TrimSpace(values[1]))
	}
	return result
}

func hasDomain(text string) bool {
	return strings.HasPrefix(text, "domain:") ||
		strings.Contains(strings.ToLower(text), "domain name:")
}

func getDomain(text string) string {
	result := ""
	if hasDomain(text) {
		re := regexp.MustCompile(`^domain.*?: (.*?)$`)
		match := re.FindStringSubmatch(strings.ToLower(text))
		if match != nil {
			result = strings.TrimSpace(match[1])
		}
	}
	return result
}

func hasExpiration(text string) bool {
	re := regexp.MustCompile(`(?i)^((domain expires)|(registry expiry date)):.*?$`)
	return re.MatchString(strings.TrimSpace(text))
}

func getExpiration(text string) time.Time {
	if hasExpiration(text) {
		re := regexp.MustCompile(`(?i)^.*?expir.*?: (.*?)$`)
		match := re.FindStringSubmatch(text)
		if match != nil {
			value := strings.TrimSpace(match[1])
			result, err := time.Parse(time.RFC3339, value)
			if err != nil {
				const shortForm = "02-Jan-2006"
				result, err = time.Parse(shortForm, value)
				if err != nil {
					log.Println(err.Error())
					result = time.Now()
				}
			}
			return result
		}
	}
	return time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)
}
