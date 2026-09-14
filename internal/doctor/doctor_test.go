package doctor

import "testing"

func TestHealthy(t *testing.T) {
	checks := []Check{{Name: "one", Found: true}, {Name: "two", Found: true}}
	if !Healthy(checks) {
		t.Fatal("expected healthy checks")
	}
}

func TestHealthyRejectsMissingCheck(t *testing.T) {
	checks := []Check{{Name: "one", Found: true}, {Name: "two", Found: false}}
	if Healthy(checks) {
		t.Fatal("expected unhealthy checks")
	}
}
