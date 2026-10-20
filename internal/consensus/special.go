package consensus

import "net/netip"

// BuiltinSpecialUse is the offline safety net for the IANA filter feed.
// It intentionally excludes reserved and documentation ranges but keeps
// ordinary public CDN and hosting provider address space.
func BuiltinSpecialUse() []netip.Prefix {
	raw := []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8",
		"169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24",
		"192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24",
		"203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
		"::/96", "::ffff:0:0/96", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64",
		"2001:1::/32", "2001:2::/31", "2001:4::/30", "2001:8::/29", "2001:10::/28",
		"2001:20::/27", "2001:40::/26", "2001:80::/25", "2001:100::/24", "2001:db8::/32",
		"2002::/16", "fc00::/7", "fe80::/10", "ff00::/8",
	}
	out := make([]netip.Prefix, 0, len(raw))
	for _, value := range raw {
		out = append(out, netip.MustParsePrefix(value))
	}
	return out
}
