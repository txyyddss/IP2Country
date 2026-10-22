package model

import "net/netip"

// The zero value preserves legacy geolocation records and configurations.
type EvidenceClass uint8

const (
	Geolocation EvidenceClass = iota
	Fallback
	Manual
)

func (e EvidenceClass) String() string {
	switch e {
	case Fallback:
		return "fallback"
	case Manual:
		return "manual"
	default:
		return "geolocation"
	}
}

func (e EvidenceClass) Priority() int {
	if e == Manual {
		return -1
	}
	return int(e)
}

type Override struct {
	CIDR    string `json:"cidr"`
	Country string `json:"country"`
	Reason  string `json:"reason,omitempty"`
}

type Record struct {
	Evidence  EvidenceClass
	Prefix    netip.Prefix
	Country   string
	SourceID  string
	VoteGroup string
	Order     int
	RowOrder  int
}

type SourceStatus struct {
	EvidenceClass string `json:"evidence_class,omitempty"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	URL           string `json:"url,omitempty"`
	Bytes         int64  `json:"bytes,omitempty"`
	ParsedRows    int64  `json:"parsed_rows"`
	RejectedRows  int64  `json:"rejected_rows"`
	NoCountryRows int64  `json:"no_country_rows"`
	IPv4Records   int64  `json:"ipv4_records"`
	IPv6Records   int64  `json:"ipv6_records"`
	Error         string `json:"error,omitempty"`
}

type CountryStats struct {
	Country     string `json:"country"`
	IPv4Blocks  int64  `json:"ipv4_blocks"`
	IPv6Blocks  int64  `json:"ipv6_blocks"`
	IPv4Cells   string `json:"ipv4_cells"`
	IPv6Cells   string `json:"ipv6_cells"`
	AverageConf string `json:"average_confidence"`
}

type RunSummary struct {
	Overrides        []OverrideStatus         `json:"overrides"`
	EvidenceCoverage map[string]CoverageStats `json:"evidence_coverage"`
	GeneratedAt      string                   `json:"generated_at_utc"`
	Sources          []SourceStatus           `json:"sources"`
	Countries        []CountryStats           `json:"countries"`
	InputRecords     int64                    `json:"input_records"`
	OutputBlocks     int64                    `json:"output_blocks"`
}

type CoverageStats struct {
	IPv4Blocks    int64  `json:"ipv4_blocks"`
	IPv6Blocks    int64  `json:"ipv6_blocks"`
	IPv4Addresses string `json:"ipv4_addresses"`
	IPv6Addresses string `json:"ipv6_addresses"`
}

type OverrideStatus struct {
	Override
	Applied CoverageStats `json:"applied"`
}
