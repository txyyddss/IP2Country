package model

import "net/netip"

type Record struct {
	Prefix    netip.Prefix
	Country   string
	SourceID  string
	VoteGroup string
	Order     int
	RowOrder  int
}

type SourceStatus struct {
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
	GeneratedAt  string         `json:"generated_at_utc"`
	Sources      []SourceStatus `json:"sources"`
	Countries    []CountryStats `json:"countries"`
	InputRecords int64          `json:"input_records"`
	OutputBlocks int64          `json:"output_blocks"`
}
