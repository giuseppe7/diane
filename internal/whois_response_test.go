package internal

import (
	"testing"
)

func TestNewWhoisResponse(t *testing.T) {
	resp := NewWhoisResponse()

	if resp.isAvailable != false {
		t.Errorf("new whoisResponse should be default to not available")
	} else if resp.hasExpiration != false {
		t.Errorf("new whoisResponse should be default to not having an expiration")
	} else if resp.refer != "" {
		t.Errorf("new whoisResponse should have an empty refer value")
	} else if resp.raw != "" {
		t.Errorf("new whoisResponse should have an empty raw value")
	}
}
