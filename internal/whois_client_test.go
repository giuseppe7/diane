package internal

import (
	"testing"
	"time"
)

const testApplicationNamespace = "test_diane"

var whois = NewWhoisClient(testApplicationNamespace)

type whoisTestData struct {
	target        string
	domain        string
	hasExpiration bool
	expiration    time.Time
}

func TestWhoisClientAvailable(t *testing.T) {

	var tests = []whoisTestData{
		{target: "somethingmadeup123.com"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := whois.Query(tt.target)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.raw == "" {
				t.Errorf("whoisQuery(%s) expected non empty raw response.", tt.target)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			} else if resp.status != ResponseAvailable {
				t.Errorf("whoisQuery(%s) expected to be available.", tt.target)
			}
		})
	}
}

func TestWhoisClientNotAvailable(t *testing.T) {

	var tests = []whoisTestData{
		{target: "example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := whois.Query(tt.target)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.raw == "" {
				t.Errorf("whoisQuery(%s) expected non empty raw response.", tt.target)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.status == ResponseAvailable {
				t.Errorf("whoisQuery(%s) expected to be not available. %+v", tt.target, resp.raw)
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			}
		})
	}
}

func TestWhoisClientExampleNoRefer(t *testing.T) {

	var tests = []whoisTestData{
		{target: "example.com", domain: "example.com"},
		{target: "example.net", domain: "example.net"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := whois.Query(tt.target)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.domain != tt.domain {
				t.Errorf("whois.Query(%s) got domain %v, expected %v", tt.target, resp.domain, tt.domain)
			}
		})
	}
}

func TestWhoisClientExampleWithRefer(t *testing.T) {

	var tests = []whoisTestData{
		{target: "example.edu", domain: "example.edu"},
		{target: "example.org", domain: "example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			resp := whois.Query(tt.domain)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.domain != tt.domain {
				t.Errorf("whois.Query(%s) got domain %v, expected %v", tt.target, resp.domain, tt.domain)
			}
		})
	}
}

func TestWhoisClientForExpirations(t *testing.T) {

	var tests = []whoisTestData{
		{target: "example.com", domain: "example.com", hasExpiration: false},
		{target: "example.net", domain: "example.net", hasExpiration: false},
		{target: "example.edu", domain: "example.edu", hasExpiration: true, expiration: time.Date(2023, 7, 31, 0, 0, 0, 0, time.UTC)},
		{target: "example.org", domain: "example.org", hasExpiration: true, expiration: time.Date(2010, 8, 30, 0, 0, 0, 0, time.UTC)},
		{target: "github.com", domain: "github.com", hasExpiration: true, expiration: time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC)},
		{target: "gitlab.com", domain: "gitlab.com", hasExpiration: true, expiration: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := whois.Query(tt.target)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.domain != tt.domain {
				t.Errorf("whois.Query(%s) domain was %v, expected %v", tt.target, resp.domain, tt.domain)
			} else if resp.hasExpiration != tt.hasExpiration {
				t.Errorf("whois.Query(%s) hasExpiration was %v, expected %v", tt.target, resp.hasExpiration, tt.hasExpiration)
			} else if resp.hasExpiration {
				// TODO: Check year, month, day only for now.
				if resp.expiration.Year() != tt.expiration.Year() ||
					resp.expiration.Month() != tt.expiration.Month() ||
					resp.expiration.Day() != tt.expiration.Day() {
					t.Errorf("whois.Query(%s) expiration was %v, expected %v", tt.target, resp.expiration.Local(), tt.expiration.Local())
				}

			}
		})
	}
}

func TestWhoisClientForNotAuthorized(t *testing.T) {
	// Apparently .es uses an unconventional whois server.
	// https://en.wikipedia.org/wiki/.es
	var tests = []whoisTestData{
		{target: "example.es", domain: "example.es"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := whois.Query(tt.target)

			if resp.err != nil {
				t.Errorf("whoisQuery(%s) error %s", tt.target, resp.err.Error())
			} else if resp.hostPort == "" {
				t.Errorf("whoisQuery(%s) expected non empty hostPort. %v", tt.target, resp.hostPort)
			} else if resp.target == "" {
				t.Errorf("whoisQuery(%s) expected non empty target.", tt.target)
			} else if resp.status != ResponseUnauthorized {
				t.Errorf("whois.Query(%s) expected to return unauthorized.", tt.target)
			}
		})
	}
}
