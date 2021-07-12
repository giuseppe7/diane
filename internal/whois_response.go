package internal

import "time"

// Non-exhaustive list of values from a whois query.
type WhoisResponse struct {
	isAvailable bool
	refer       string
	domain      string
	// What a stupid inconsisent spec.
	hasExpiration bool
	expiration    time.Time
	// Raw response.
	raw string
}

func NewWhoisResponse() WhoisResponse {
	resp := WhoisResponse{}
	resp.isAvailable = false
	resp.hasExpiration = false
	return resp
}
