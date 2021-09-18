package internal

import "testing"

func TestInitConfiguration(t *testing.T) {
	appConfig := InitConfiguration()
	if len(appConfig.Domains) < 4 {
		t.Errorf("expected at least four domains in configuration, found %d", len(appConfig.Domains))
	}
}
