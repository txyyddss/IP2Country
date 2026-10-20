package iprange

import (
	"net/netip"
	"testing"
)

func TestRangeToPrefixesHandlesIPv4Maximum(t *testing.T) {
	got := RangeToPrefixes(netip.MustParseAddr("255.255.255.254"), netip.MustParseAddr("255.255.255.255"))
	if len(got) != 1 || got[0].String() != "255.255.255.254/31" {
		t.Fatalf("unexpected prefix cover: %v", got)
	}
}

func TestRangeToPrefixesDoesNotExpandLargeIPv6Range(t *testing.T) {
	got := RangeToPrefixes(netip.MustParseAddr("2001:4860::"), netip.MustParseAddr("2001:4860:ffff:ffff:ffff:ffff:ffff:ffff"))
	if len(got) != 1 || got[0].String() != "2001:4860::/32" {
		t.Fatalf("unexpected prefix cover: %v", got)
	}
}

func TestSubtractPrefixes(t *testing.T) {
	got := SubtractPrefixes([]netip.Prefix{netip.MustParsePrefix("192.0.2.0/23")}, []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")})
	if len(got) != 1 || got[0].String() != "192.0.3.0/24" {
		t.Fatalf("unexpected subtraction: %v", got)
	}
}
