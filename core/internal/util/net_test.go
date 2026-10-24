package util

import "testing"

func TestIPToCIDR(t *testing.T) {
	if got := IPToCIDR("192.0.2.1 255.255.255.0"); got != "192.0.2.1/24" {
		t.Fatalf("got %s", got)
	}
	if got := IPToCIDR("2001:db8::1 64"); got != "2001:db8::1/64" {
		t.Fatalf("v6 got %s", got)
	}
	if got := IPToCIDR("10.0.0.0/8"); got != "10.0.0.0/8" {
		t.Fatalf("passthrough %s", got)
	}
}
