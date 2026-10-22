package config

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/biter777/countries"
	"github.com/txyyddss/ip2country/internal/model"
)

type File struct {
	Overrides []model.Override `json:"overrides,omitempty"`
	Version   int              `json:"version"`
	Settings  Settings         `json:"settings"`
	Sources   []Source         `json:"sources"`
}

type Settings struct {
	IPv4Prefix int    `json:"ipv4_prefix"`
	IPv6Prefix int    `json:"ipv6_prefix"`
	SpecialUse string `json:"special_use_source"`
}

type Source struct {
	EvidenceClass   string   `json:"evidence_class,omitempty"`
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Enabled         *bool    `json:"enabled"`
	VoteGroup       string   `json:"vote_group"`
	Role            string   `json:"role"`
	Format          string   `json:"format"`
	URLs            []string `json:"urls"`
	FallbackURLs    []string `json:"fallback_urls"`
	Family          string   `json:"family"`
	IPEncoding      string   `json:"ip_encoding,omitempty"`
	CountryOverride string   `json:"country_override"`
	TextRegex       string   `json:"text_regex"`
	Delimiter       string   `json:"delimiter"`
	Header          *bool    `json:"header"`
	CIDRColumn      string   `json:"cidr_column"`
	StartColumn     string   `json:"start_column"`
	EndColumn       string   `json:"end_column"`
	CountryColumn   string   `json:"country_column"`
	ArchiveMember   string   `json:"archive_member"`
}

func Load(path string) (File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return File{}, fmt.Errorf("decode config: %w", err)
	}
	if f.Version != 1 {
		return File{}, fmt.Errorf("unsupported config version %d", f.Version)
	}
	if f.Settings.IPv4Prefix != 24 || f.Settings.IPv6Prefix != 64 {
		return File{}, fmt.Errorf("this generator requires /24 IPv4 and /64 IPv6 vote cells")
	}
	seenOverrides := make(map[string]string)
	uniqueOverrides := make([]model.Override, 0, len(f.Overrides))
	for i, override := range f.Overrides {
		prefix, err := netip.ParsePrefix(override.CIDR)
		if err != nil || prefix.Addr().Is4In6() || prefix.Masked().String() != override.CIDR {
			return File{}, fmt.Errorf("override %d requires a canonical IPv4 or IPv6 CIDR: %q", i, override.CIDR)
		}
		override.Country = strings.ToUpper(strings.TrimSpace(override.Country))
		if override.Country == "UK" {
			override.Country = "GB"
		}
		if len(override.Country) != 2 || override.Country[0] < 'A' || override.Country[0] > 'Z' || override.Country[1] < 'A' || override.Country[1] > 'Z' || (override.Country != "XK" && countries.ByName(override.Country).Alpha2() != override.Country) || override.Country == "ZZ" {
			return File{}, fmt.Errorf("override %s has invalid country %q", override.CIDR, override.Country)
		}
		if previous, ok := seenOverrides[override.CIDR]; ok {
			if previous != override.Country {
				return File{}, fmt.Errorf("conflicting overrides for %s", override.CIDR)
			}
			continue
		}
		seenOverrides[override.CIDR] = override.Country
		uniqueOverrides = append(uniqueOverrides, override)
	}
	f.Overrides = uniqueOverrides
	seen := make(map[string]bool)
	for i := range f.Sources {
		s := &f.Sources[i]
		s.ID = strings.TrimSpace(s.ID)
		s.Name = strings.TrimSpace(s.Name)
		s.Format = strings.ToLower(strings.TrimSpace(s.Format))
		s.EvidenceClass = strings.ToLower(strings.TrimSpace(s.EvidenceClass))
		if s.EvidenceClass == "" {
			s.EvidenceClass = "geolocation"
		}
		if s.EvidenceClass != "geolocation" && s.EvidenceClass != "fallback" {
			return File{}, fmt.Errorf("source %q has unknown evidence class %q", s.ID, s.EvidenceClass)
		}
		s.IPEncoding = strings.ToLower(strings.TrimSpace(s.IPEncoding))
		if s.IPEncoding != "" && s.IPEncoding != "hex" && s.IPEncoding != "decimal" {
			return File{}, fmt.Errorf("source %q has unknown IP encoding %q", s.ID, s.IPEncoding)
		}
		if s.ID == "" || s.Name == "" || len(s.URLs) == 0 {
			return File{}, fmt.Errorf("source %d must have id, name, and at least one URL", i)
		}
		if seen[s.ID] {
			return File{}, fmt.Errorf("duplicate source id %q", s.ID)
		}
		seen[s.ID] = true
		if s.Role == "" {
			s.Role = "country"
		}
		if s.Enabled == nil {
			enabled := true
			s.Enabled = &enabled
		}
		if s.Role != "country" && s.Role != "filter" {
			return File{}, fmt.Errorf("source %q has unknown role %q", s.ID, s.Role)
		}
		if s.Role == "country" && s.VoteGroup == "" {
			s.VoteGroup = s.ID
		}
	}
	return f, nil
}

func (s Source) Evidence() model.EvidenceClass {
	if s.EvidenceClass == "fallback" {
		return model.Fallback
	}
	return model.Geolocation
}
