package internal

import (
	"testing"
)

func TestNewWhoisResponse(t *testing.T) {
	resp := NewWhoisResponse()

	if resp.target != "" {
		t.Errorf("new whoisResponse should have an empty target")
	} else if resp.hostPort != "" {
		t.Errorf("new whoisResponse should have an empty hostPort value")
	} else if resp.raw != "" {
		t.Errorf("new whoisResponse should have an empty raw value")
	} else if resp.status != ResponseUnknown {
		t.Errorf("new whoisResponse should be default to unknown")
	} else if resp.err != nil {
		t.Errorf("new whoisResponse should default to not having an error")
	} else if resp.refer != "" {
		t.Errorf("new whoisResponse should have an empty refer value")
	} else if resp.hasExpiration != false {
		t.Errorf("new whoisResponse should be default to not having an expiration")
	}
}
