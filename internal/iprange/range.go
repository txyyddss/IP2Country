package iprange

import (
	"math/big"
	"net/netip"
	"sort"
)

type filterPrefix struct {
	prefix netip.Prefix
	end    netip.Addr
}
type PrefixFilter struct{ ipv4, ipv6 []filterPrefix }

func NewPrefixFilter(exclusions []netip.Prefix) *PrefixFilter {
	f := &PrefixFilter{}
	for _, p := range exclusions {
		if !p.IsValid() {
			continue
		}
		p = p.Masked()
		entry := filterPrefix{prefix: p, end: PrefixLast(p)}
		if p.Addr().Is4() {
			f.ipv4 = append(f.ipv4, entry)
		} else {
			f.ipv6 = append(f.ipv6, entry)
		}
	}
	less := func(items []filterPrefix) {
		sort.Slice(items, func(i, j int) bool { return items[i].prefix.Addr().Compare(items[j].prefix.Addr()) < 0 })
	}
	less(f.ipv4)
	less(f.ipv6)
	return f
}

func (f *PrefixFilter) SubtractPrefix(p netip.Prefix) []netip.Prefix {
	if !p.IsValid() {
		return nil
	}
	p = p.Masked()
	items := f.ipv4
	if p.Addr().Is6() {
		items = f.ipv6
	}
	last := PrefixLast(p)
	first := sort.Search(len(items), func(i int) bool { return items[i].end.Compare(p.Addr()) >= 0 })
	var relevant []netip.Prefix
	for i := first; i < len(items) && items[i].prefix.Addr().Compare(last) <= 0; i++ {
		relevant = append(relevant, items[i].prefix)
	}
	if len(relevant) == 0 {
		return []netip.Prefix{p}
	}
	return subtract(p, relevant)
}

func PrefixLast(p netip.Prefix) netip.Addr {
	p = p.Masked()
	width := p.Addr().BitLen()
	value := addrInt(p.Addr())
	hostBits := width - p.Bits()
	if hostBits > 0 {
		value.Or(value, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(hostBits)), big.NewInt(1)))
	}
	return intAddr(value, width)
}

func RangeToPrefixes(start, end netip.Addr) []netip.Prefix {
	if !start.IsValid() || !end.IsValid() || start.Is4() != end.Is4() || start.Compare(end) > 0 {
		return nil
	}
	width := start.BitLen()
	cur, limit := addrInt(start), addrInt(end)
	var out []netip.Prefix
	for cur.Cmp(limit) <= 0 {
		remaining := new(big.Int).Sub(limit, cur)
		remaining.Add(remaining, big.NewInt(1))
		maxHost := remaining.BitLen() - 1
		if maxHost > width {
			maxHost = width
		}
		align := trailingZeroBits(cur, width)
		if align > maxHost {
			align = maxHost
		}
		out = append(out, netip.PrefixFrom(intAddr(cur, width), width-align))
		next := new(big.Int).Add(cur, new(big.Int).Lsh(big.NewInt(1), uint(align)))
		if next.BitLen() > width {
			return out
		}
		cur = next
	}
	return out
}

func addrInt(addr netip.Addr) *big.Int {
	if addr.Is4() {
		b := addr.As4()
		return new(big.Int).SetBytes(b[:])
	}
	b := addr.As16()
	return new(big.Int).SetBytes(b[:])
}

func intAddr(n *big.Int, width int) netip.Addr {
	b := n.Bytes()
	if width == 32 {
		var raw [4]byte
		copy(raw[4-len(b):], b)
		return netip.AddrFrom4(raw)
	}
	var raw [16]byte
	copy(raw[16-len(b):], b)
	return netip.AddrFrom16(raw)
}

func trailingZeroBits(n *big.Int, width int) int {
	if n.Sign() == 0 {
		return width
	}
	for bit := 0; bit < width; bit++ {
		if n.Bit(bit) != 0 {
			return bit
		}
	}
	return width
}

func SubtractPrefixes(input, exclusions []netip.Prefix) []netip.Prefix {
	f := NewPrefixFilter(exclusions)
	out := make([]netip.Prefix, 0, len(input))
	for _, p := range input {
		out = append(out, f.SubtractPrefix(p)...)
	}
	return out
}

func subtract(p netip.Prefix, exclusions []netip.Prefix) []netip.Prefix {
	fragments := []netip.Prefix{p}
	for _, ex := range exclusions {
		if len(fragments) == 0 {
			break
		}
		next := make([]netip.Prefix, 0, len(fragments)+1)
		for _, fragment := range fragments {
			next = append(next, subtractOne(fragment, ex)...)
		}
		fragments = next
	}
	return fragments
}

func subtractOne(p, ex netip.Prefix) []netip.Prefix {
	if !p.IsValid() || !ex.IsValid() || p.Addr().Is4() != ex.Addr().Is4() {
		return []netip.Prefix{p}
	}
	if ex.Bits() <= p.Bits() && ex.Contains(p.Addr()) {
		return nil
	}
	if !p.Contains(ex.Addr()) {
		return []netip.Prefix{p}
	}
	if p.Bits() >= p.Addr().BitLen() {
		return []netip.Prefix{p}
	}
	left, right := children(p)
	return append(subtractOne(left, ex), subtractOne(right, ex)...)
}

func children(p netip.Prefix) (netip.Prefix, netip.Prefix) {
	p = p.Masked()
	nextBits := p.Bits() + 1
	left := netip.PrefixFrom(p.Addr(), nextBits)
	value := addrInt(left.Addr())
	value.SetBit(value, p.Addr().BitLen()-nextBits, 1)
	return left, netip.PrefixFrom(intAddr(value, p.Addr().BitLen()), nextBits)
}
