package internal

import (
	"log"
	"regexp"
	"strings"
	"time"
)

type WhoisResponseType int

const (
	ResponseUnknown WhoisResponseType = iota
	ResponseOk
	ResponseError
	ResponseAvailable
	ResponseUnauthorized
	ResponseExceededRate
)

func (wrt WhoisResponseType) String() string {
	return [...]string{"Unknown", "OK", "Error", "Available", "Unauthorized", "ExceededRate"}[wrt]
}

// Non-exhaustive list of values from a whois query.
type WhoisResponse struct {
	target   string            // What we queried.
	hostPort string            // Who we queried.
	raw      string            // Raw response.
	status   WhoisResponseType // Response status using enums above.
	err      error             // Error exception caught for ResponseError use cases.
	// Parsed values below.
	refer         string    // Parsed refer response in case of another query is required.
	domain        string    // Parsed domain in the final response.
	hasExpiration bool      // Determines if expiry was parsed.
	expiration    time.Time // Actual expiry that was parsed.
}

func NewWhoisResponse() WhoisResponse {
	resp := WhoisResponse{}
	resp.status = ResponseUnknown
	resp.hasExpiration = false
	return resp
}

func (r *WhoisResponse) ParseRawResponse(raw string) {
	r.raw = raw
	r.status = ResponseOk // Default to OK at this point unless we have a value below.

	if hasRefer(raw) {
		r.refer = getRefer(raw)
	}
	if noMatchFound(raw) {
		r.status = ResponseAvailable
	}
	if hasDomain(raw) {
		r.domain = getDomain(raw)
	}
	if hasExpiration(raw) {
		r.hasExpiration = true
		r.expiration = getExpiration(raw)
	}
	if hasExceededQueries(raw) {
		r.status = ResponseExceededRate
	}
	if notAuthorized(raw) {
		r.status = ResponseUnauthorized
	}
}

func hasRefer(text string) bool {
	re := regexp.MustCompile(`(?i)refer:`)
	return re.MatchString(strings.TrimSpace(text))
}

func getRefer(text string) string {
	result := ""
	if hasRefer(text) {
		re := regexp.MustCompile(`(?is)refer:\s+(.*?)\s+`)
		match := re.FindStringSubmatch(text)
		if match != nil {
			result = match[1]
		}
	}
	return result
}

func noMatchFound(text string) bool {
	re := regexp.MustCompile(`(?im)((no match for)|(not found)|(no data found))`)
	return re.MatchString(strings.TrimSpace(text))
}

func hasDomain(text string) bool {
	re := regexp.MustCompile(`(?i)\s*((domain)|(domain name)):\s+(.*?)\s+`)
	return re.MatchString(strings.TrimSpace(text))
}

func getDomain(text string) string {
	result := ""
	if hasDomain(text) {
		re := regexp.MustCompile(`(?i)\s*?((domain)|(domain name)):\s+(.*?)\s+`)
		match := re.FindStringSubmatch(strings.ToLower(text))
		if match != nil {
			result = strings.TrimSpace(match[4])
		}
	}
	return result
}

func hasExpiration(text string) bool {
	re := regexp.MustCompile(`(?i)((domain expires)|(registry expiry date)|(expiry date)|(expire date)|(expires)):\s+.*?\s+`)
	return re.MatchString(strings.TrimSpace(text))
}

func getExpiration(text string) time.Time {
	if hasExpiration(text) {
		re := regexp.MustCompile(`(?i).*?expir.*?:\s+(.*?)\s+`)
		match := re.FindStringSubmatch(text)
		if match != nil {
			value := strings.TrimSpace(match[1])
			result, err := time.Parse(time.RFC3339, value)
			if err != nil {
				// OK, RFC3339 parse fail, try one of two short forms.
				const shortForm = "02-Jan-2006"
				result, err = time.Parse(shortForm, value)
				if err != nil {
					// OK, one more time.
					const shortForm2 = "2006-01-02"
					result, err = time.Parse(shortForm2, value)
					if err != nil {
						log.Println("error in parsing both short forms", err.Error())
						result = time.Now()
					}
				}
			}
			return result
		}
	}
	return time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)
}

func notAuthorized(text string) bool {
	re := regexp.MustCompile(`(?i)( not authorised )`)
	return re.MatchString(strings.TrimSpace(text))
}

func hasExceededQueries(text string) bool {
	re := regexp.MustCompile(`(?i)^.*( queries exceeded.)$`)
	return re.MatchString(strings.TrimSpace(text))
}
