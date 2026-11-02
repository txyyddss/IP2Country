# IP2Country GeoIP Generator

A Go program that combines public IP-to-country feeds into compact country databases and country-specific CSV files. It produces Xray/Mihomo V2Fly protobuf GeoIP, current sing-box rule sets, and a MaxMind DB country database.

The input inventory, parser settings, enabled state, evidence classes, vote groups, source order, and manual overrides live in [`config/sources.json`](config/sources.json). Feeds are enabled unless explicitly disabled in the config. The IANA feed supplies country-tagged allocation prefixes; built-in special-use ranges are excluded. Public Cloudflare and hosting ranges remain eligible.

## Build locally

Requires Go 1.26 or newer and the pinned sing-box rule-set compiler. Install the compiler with:

```powershell
go install github.com/sagernet/sing-box/cmd/sing-box@v1.14.0
```

Then run:

```powershell
go run ./cmd/geoipgen -config config/sources.json -work-dir .cache/geoip -output-dir dist
```

The first run downloads source data. Later runs keep one last-known-good copy per source under the work directory. A failed source uses its cache when available; a source with no usable download or cache is skipped and reported. The program refuses to generate a release when no country data can be used.

For an offline rebuild from cached feeds, add `-offline` to the command.

## Output files

- `geoip.dat` — V2Fly protobuf country database for Xray and Mihomo DAT mode.
- `country.mmdb` — country-schema MaxMind DB for Mihomo and general MMDB readers.
- `sing-box/geoip-<country>.json` and `.srs` — per-country rule sets for current sing-box.
- `countries/<country>.csv` — final country blocks, confidence, and vote counts.
- `source-csv/<source>/<country>.csv` — normalized per-source country blocks.
- ZIP archives, `run-summary.json`, `manifest.json`, and `SHA256SUMS` accompany release builds.

## Country selection and manual overrides

Add manual rules to the top-level `overrides` array in `config/sources.json`:

```json
"overrides": [
  { "cidr": "119.28.225.0/24", "country": "HK", "reason": "Confirmed Hong Kong range" }
]
```

Rules accept canonical IPv4 or IPv6 CIDRs and ISO country codes (including `XK`; `UK` normalizes to `GB`). Invalid networks, host bits, IPv4-mapped IPv6 networks, unknown countries, and conflicting assignments for an identical CIDR fail before downloads begin. Identical same-country rules are deduplicated; nested rules use the longest matching prefix. Manual rules override automatic assignments and fill uncovered public space, while special-use exclusions still apply. At least one usable upstream country source remains required. Remove a rule to restore automatic selection.

The bundled overrides assign `119.28.225.0/24`, `47.243.107.0/24`, `8.217.174.0/24`, and `47.242.137.0/24` to HK.

Voting follows exact source boundaries, including splits smaller than IPv4 /24 and IPv6 /64. Source `evidence_class` is `geolocation` or `fallback`; omitted values default to `geolocation` for existing configurations. The bundled configuration treats DB-IP, GeoLite2, IPinfo, IP2Location, IPLocate, SypexGeo, and IPAPI as geolocation evidence. Allocation/ASN feeds, mixed-source aggregations, and derived routing databases are fallback evidence. At each address, only geolocation evidence votes when available; fallback evidence votes when no geolocation evidence covers that address. Lower-tier sources cannot break a higher-tier tie.

Each provider group casts at most one vote. Within a group, the stronger evidence class wins, followed by the most-specific prefix, source order, and input-row order. Across groups, the country with most votes wins; ties follow configuration order, then country code. Mirrors and known derivatives share groups: the configured China-operator feed with GeoLite2; Sapics server/user-country, Geo-Whois, and HostByBelle together. V2Fly and Meta/Loyalsoldier feeds are absent from the current inventory; if added, configure them as fallback evidence in the DB-IP and GeoLite2 groups respectively. These conservative family groups prevent repeated upstream data from adding votes. They do not imply that every derived row is identical to its upstream provider.

MMDB location-country fields take precedence over `registered_country`. Registration-only records are fallback evidence even when the source normally supplies geolocation. No-country records, such as metadata-only MMDB entries and regional tags like `EU`, are reported separately from malformed records. Country ISO-code columns take precedence over country names. Integer address feeds can declare `ip_encoding` as `hex` or `decimal`; tdulcet TSV feeds use hexadecimal endpoints.

Final country CSVs append `assignment_method` (`manual`, `geolocation`, or `fallback`) after the existing columns. Confidence is agreement among participating provider groups in the selected evidence class, averaged by address space when compatible winning ranges merge. Vote totals count vote-address pairs. Manual rows have blank confidence and zero provider votes and are excluded from confidence averages. Method boundaries and individual override boundaries remain distinct. `run-summary.json` and `manifest.json` include override definitions, their effective coverage after nested rules and exclusions, and exact IPv4/IPv6 address counts by evidence class. Confidence is not a measured probability of geographic correctness.

Source provenance: [V2Fly configuration](https://github.com/v2fly/geoip/blob/main/config.json), [Sapics datasets](https://github.com/sapics/ip-location-db#databases-referenced-in-this-project), [MaxMind country fields](https://dev.maxmind.com/geoip/docs/databases/city-and-country/country-binary/).

Range processing sorts exact boundaries on disk without expanding addresses, /24s, or /64s. It handles broad IPv6 networks with bounded event-sort memory. Preprocessing is O(n log n); final sorted-prefix lookups are O(log n). README /24 and /64 coverage totals are summed exactly and rounded once to whole units; JSON totals retain fractional units.

## Automated updates

The GeoIP workflow runs at 00:00 UTC each day and supports `workflow_dispatch`. It publishes one dated GitHub release per UTC day, reuses that release on reruns, and updates this README's generated statistics section. New releases stay drafts until all eight assets have uploaded with matching sizes and SHA-256 hashes. Only after verified publication does cleanup delete older published `geoip-YYYYMMDD` releases, their assets, and their Git tags. The current release, newer dates, drafts, prereleases, and unrelated releases/tags are preserved. Failed upload or verification skips cleanup; cleanup errors fail the workflow and can be retried. Older orphaned GeoIP tags left by interrupted cleanup are removed on a later successful run. Same-day uploads retain the existing asset replacement behavior. GitHub Actions caches source files between runs. The quality workflow runs formatting, Go tests, mocked release-publication tests, vet, and build checks on pushes and pull requests.

<!-- GEOIP-STATISTICS:START -->
## Generated data status

Last run: `2026-11-02T04:25:43Z`

Input rows: **39289606** · Published blocks: **2391588**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 563424 | 0 | 0 | 0 | 11130321 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 498242 | 0 | 0 | 15023010 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 565442 | 0 | 0 | 0 | 11841456 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 498343 | 0 | 0 | 34385667 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 402523 | 0 | 0 | 0 | 8783869 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 473542 | 0 | 1 | 15581270 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 402555 | 0 | 0 | 0 | 5999811 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 473591 | 0 | 1 | 18415755 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 419618 | 0 | 0 | 0 | 9164474 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 565450 | 0 | 1 | 16796047 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 987857 | 0 | 0 | 0 | 13047911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2808358 | 0 | 0 | 89677782 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1346923 | 3039074 | 0 | 0 | 215117937 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 389960 | 0 | 0 | 0 | 6095165 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1556723 | 0 | 0 | 24472437 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 788145 | 1371823 | 0 | 0 | 17287160 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177221 | 69787 | 0 | 0 | 3398263 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 289320 | 488523 | 0 | 97881 | 40781675 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 289320 | 488523 | 0 | 98386 | 18433259 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4347 | 1685 | 0 | 0 | 98224 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3887 | 0 | 0 | 0 | 82479 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 811989 | 429296 | 0 | 411579 | 18597237 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262431 | 71052 | 0 | 10 | 34598640 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802828 | 0 | 0 | 0 | 11122714 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6174 | 3479 | 0 | 0 | 153949 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 566210 | 499044 | 0 | 0 | 3401659 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 599115 | 3702257408 | 1792223 | 293421847207842436812546030166054273024 |
| fallback | 0 | 0 | 246 | 42535295865116098988655467213201866752 |
| manual | 4 | 1024 | 0 | 0 |

### Manual overrides

| Network | Country | Applied IPv4 addresses | Applied IPv6 addresses | Reason |
|---|---|---:|---:|---|
| 119.28.225.0/24 | HK | 256 | 0 | Manually confirmed Hong Kong range |
| 47.243.107.0/24 | HK | 256 | 0 | Manually confirmed Hong Kong range |
| 8.217.174.0/24 | HK | 256 | 0 | Manually confirmed Hong Kong range |
| 47.242.137.0/24 | HK | 256 | 0 | Manually confirmed Hong Kong range |

Confidence measures agreement among participating provider groups, not measured geolocation accuracy. Manual assignments are excluded from confidence averages.

### Final country coverage

| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |
|---|---:|---:|---:|---:|---:|
| AD | 213 | 308 | 181 | 43728694343 | 92.57% |
| AE | 2830 | 25495 | 2549 | 12830481150660 | 89.87% |
| AF | 219 | 875 | 519 | 121136965514 | 97.42% |
| AG | 126 | 177 | 184 | 40393304124 | 65.26% |
| AI | 44 | 45 | 101 | 5847641145 | 67.40% |
| AL | 597 | 1417 | 1244 | 2500693679258 | 95.22% |
| AM | 353 | 2322 | 518 | 1259317991630 | 95.16% |
| AO | 605 | 5144 | 298 | 147599966562 | 98.68% |
| AQ | 56 | 5 | 310 | 472128653 | 43.66% |
| AR | 3731 | 76975 | 2854 | 23617874075848 | 98.83% |
| AS | 38 | 28 | 310 | 9243191473 | 96.04% |
| AT | 4755 | 45185 | 4448 | 10471310325164 | 97.55% |
| AU | 13317 | 220281 | 71704 | 2309971764744661824 | 99.94% |
| AW | 49 | 449 | 97 | 11534575672 | 92.96% |
| AX | 92 | 131 | 605 | 69552372555 | 72.57% |
| AZ | 399 | 3069 | 351 | 1964102673028 | 97.85% |
| BA | 339 | 3265 | 742 | 833827495454 | 95.13% |
| BB | 193 | 507 | 109 | 69400586958 | 99.69% |
| BD | 2659 | 8190 | 2713 | 3577593750770 | 96.45% |
| BE | 8455 | 57880 | 4779 | 4389004620388994 | 66.65% |
| BF | 124 | 1383 | 144 | 91245525303 | 99.43% |
| BG | 2390 | 16715 | 1457 | 5137769092930 | 96.03% |
| BH | 329 | 5245 | 319 | 406108578641 | 97.65% |
| BI | 57 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 85 | 622 | 127 | 56582572337 | 96.84% |
| BL | 41 | 11 | 85 | 1047717936 | 57.46% |
| BM | 324 | 499 | 108 | 43426867513 | 99.37% |
| BN | 110 | 879 | 119 | 26250616879 | 99.06% |
| BO | 330 | 4751 | 307 | 168302404201 | 98.07% |
| BQ | 33 | 103 | 69 | 17538738225 | 98.90% |
| BR | 11574 | 335176 | 42535 | 37419670690685 | 99.25% |
| BS | 137 | 632 | 163 | 14003518038 | 96.66% |
| BT | 123 | 226 | 164 | 60978750033 | 99.35% |
| BV | 1 | 1 | 3 | 1048577 | 50.00% |
| BW | 198 | 718 | 123 | 78628794672 | 97.50% |
| BY | 397 | 7401 | 429 | 700163287349 | 99.59% |
| BZ | 209 | 470 | 198 | 192713403210 | 89.02% |
| CA | 19643 | 285488 | 30953 | 6351225300248 | 95.58% |
| CC | 6 | 0 | 51 | 827254807 | 42.84% |
| CD | 317 | 940 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 118 | 36988473393 | 97.17% |
| CH | 6875 | 98183 | 57399 | 15878800909851441318 | 99.98% |
| CI | 353 | 6844 | 169 | 194921285945 | 98.93% |
| CK | 34 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2068 | 41384 | 2405 | 1415241805342 | 96.61% |
| CM | 188 | 2438 | 160 | 67900151884 | 98.18% |
| CN | 6938 | 1347088 | 12746 | 294804070741524 | 99.94% |
| CO | 3168 | 70315 | 1940 | 1981632206739 | 97.96% |
| CR | 599 | 9164 | 368 | 461701217133 | 97.36% |
| CU | 96 | 1023 | 129 | 26439696697 | 99.17% |
| CV | 82 | 241 | 143 | 13715433519 | 97.22% |
| CW | 139 | 718 | 101 | 63125202989 | 92.43% |
| CX | 8 | 0 | 84 | 727574567 | 41.43% |
| CY | 703 | 4159 | 690 | 5995536325475 | 93.97% |
| CZ | 3866 | 33840 | 6531 | 13303848535214 | 97.55% |
| DE | 37292 | 550828 | 83740 | 115396871400189 | 96.06% |
| DJ | 70 | 245 | 119 | 9775540270 | 94.20% |
| DK | 3690 | 50115 | 2078 | 9667281781671 | 93.43% |
| DM | 74 | 71 | 116 | 44401879600 | 73.08% |
| DO | 741 | 8067 | 626 | 928508605795 | 98.99% |
| DZ | 220 | 18723 | 258 | 44176730299 | 95.48% |
| EC | 1124 | 11258 | 658 | 1158571649929 | 98.76% |
| EE | 1186 | 6509 | 1840 | 2725759275268 | 96.02% |
| EG | 793 | 94817 | 371 | 17730906038686 | 99.99% |
| EH | 4 | 0 | 54 | 457959449 | 46.78% |
| ER | 44 | 29 | 90 | 966146090 | 59.37% |
| ES | 11300 | 139982 | 7004 | 26203112099265 | 97.14% |
| ET | 92 | 1478 | 133 | 1117547837495 | 99.96% |
| FI | 4032 | 58642 | 34149 | 6865768733199 | 96.74% |
| FJ | 112 | 590 | 117 | 43379780397 | 99.48% |
| FK | 38 | 25 | 84 | 612231207 | 46.31% |
| FM | 56 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 86 | 108187543852 | 99.57% |
| FR | 27350 | 325706 | 21669 | 67344343078386 | 98.52% |
| GA | 91 | 1559 | 107 | 35779721513 | 94.29% |
| GB | 39958 | 498488 | 60562 | 105840868389513 | 94.12% |
| GD | 98 | 128 | 102 | 82705771051 | 96.88% |
| GE | 487 | 5114 | 347 | 1831825782905 | 95.38% |
| GF | 263 | 258 | 80 | 23010009128 | 97.39% |
| GG | 133 | 262 | 94 | 112129796399 | 81.66% |
| GH | 372 | 8926 | 264 | 1232412718691 | 99.85% |
| GI | 145 | 326 | 106 | 197924939820 | 90.89% |
| GL | 103 | 144 | 170 | 43621736572 | 99.27% |
| GM | 76 | 1104 | 109 | 39586920746 | 98.99% |
| GN | 74 | 167 | 116 | 48293804074 | 97.24% |
| GP | 233 | 805 | 103 | 54408919096 | 73.06% |
| GQ | 53 | 95 | 91 | 9454958634 | 96.02% |
| GR | 1521 | 23126 | 1010 | 2592090977053 | 96.52% |
| GS | 4 | 2 | 42 | 406970390 | 59.18% |
| GT | 666 | 2886 | 278 | 327902657139 | 99.39% |
| GU | 155 | 996 | 132 | 34892079420 | 99.33% |
| GW | 51 | 42 | 94 | 9320609836 | 96.97% |
| GY | 57 | 288 | 108 | 35023939628 | 96.09% |
| HK | 13834 | 95150 | 8104 | 5121442631072 | 85.95% |
| HM | 1 | 2 | 5 | 17895424 | 48.96% |
| HN | 842 | 1992 | 265 | 391924755060 | 94.77% |
| HR | 839 | 9689 | 434 | 1591450979599 | 96.38% |
| HT | 83 | 624 | 133 | 26468805184 | 93.34% |
| HU | 1991 | 23365 | 1316 | 3217142834614 | 98.33% |
| ID | 6900 | 81303 | 5149 | 3830715838577 | 98.04% |
| IE | 8331 | 72291 | 8964 | 4921926519290 | 88.98% |
| IL | 2462 | 39220 | 1583 | 4297869076273 | 97.44% |
| IM | 344 | 476 | 191 | 361489574464 | 96.14% |
| IN | 13422 | 197350 | 16920 | 29266944730165 | 97.36% |
| IO | 25 | 14 | 83 | 4755679275 | 95.94% |
| IQ | 674 | 3350 | 966 | 3415584355225 | 96.90% |
| IR | 1970 | 42730 | 2209 | 16876338350988 | 98.64% |
| IS | 339 | 3639 | 500 | 1605317133183 | 96.10% |
| IT | 9927 | 227747 | 9590 | 47046803517852 | 99.08% |
| JE | 223 | 642 | 116 | 219604756782 | 95.45% |
| JM | 251 | 1131 | 151 | 1122142710360 | 99.96% |
| JO | 440 | 2702 | 253 | 1121936402780 | 97.90% |
| JP | 12546 | 782429 | 14903 | 52725584180682 | 88.45% |
| KE | 1159 | 25133 | 666 | 688320719031 | 99.32% |
| KG | 236 | 1184 | 220 | 529225198189 | 99.38% |
| KH | 554 | 2015 | 300 | 430692688487 | 81.58% |
| KI | 78 | 36 | 137 | 13520201803 | 97.75% |
| KM | 66 | 52 | 84 | 9337298982 | 96.78% |
| KN | 101 | 133 | 103 | 5765065014 | 76.99% |
| KP | 9 | 11 | 416 | 500449589 | 55.27% |
| KR | 4006 | 455095 | 3807 | 22720777643568 | 99.73% |
| KW | 421 | 7579 | 218 | 877741003201 | 99.72% |
| KY | 129 | 438 | 107 | 17863732530 | 99.08% |
| KZ | 1369 | 13971 | 1221 | 3150641399606 | 86.84% |
| LA | 146 | 444 | 151 | 121260925231 | 88.17% |
| LB | 392 | 2388 | 293 | 3015926620999 | 98.08% |
| LC | 98 | 167 | 91 | 78392387877 | 97.53% |
| LI | 375 | 452 | 201 | 429947419690 | 93.40% |
| LK | 244 | 2325 | 265 | 82540568948 | 94.31% |
| LR | 82 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1475 | 12010 | 970 | 4168685438927 | 84.93% |
| LU | 1543 | 6703 | 863 | 2222679357040 | 96.04% |
| LV | 1240 | 8154 | 780 | 3240070978784 | 65.66% |
| LY | 151 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 340 | 48132 | 264 | 1171100463780 | 99.78% |
| MC | 283 | 284 | 108 | 142166649391 | 99.88% |
| MD | 762 | 4975 | 976 | 5953066380629 | 64.06% |
| ME | 154 | 828 | 129 | 348120038202 | 99.48% |
| MF | 56 | 103 | 90 | 5377009711 | 92.70% |
| MG | 85 | 2307 | 115 | 22644159029 | 98.09% |
| MH | 73 | 90 | 81 | 4806418475 | 95.24% |
| MK | 254 | 2800 | 186 | 1142897781613 | 97.13% |
| ML | 81 | 388 | 90 | 22675267630 | 97.52% |
| MM | 459 | 1169 | 337 | 353155015037 | 97.86% |
| MN | 236 | 738 | 284 | 133921418859 | 99.23% |
| MO | 106 | 1500 | 709 | 39116173881 | 99.15% |
| MP | 69 | 15 | 78 | 4805746733 | 95.20% |
| MQ | 318 | 814 | 70 | 104815667236 | 89.69% |
| MR | 73 | 233 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 77 | 644671791 | 51.26% |
| MT | 324 | 2683 | 226 | 501308648249 | 93.64% |
| MU | 2035 | 23892 | 8188 | 4455187919499580 | 100.00% |
| MV | 120 | 403 | 135 | 18180595767 | 97.48% |
| MW | 123 | 2200 | 146 | 1242315518001 | 99.85% |
| MX | 4404 | 122502 | 2492 | 3642801328889 | 97.00% |
| MY | 2763 | 34650 | 1633 | 912008152137 | 96.07% |
| MZ | 195 | 1912 | 133 | 69652197437 | 97.38% |
| NA | 413 | 1854 | 186 | 29169103150 | 99.37% |
| NC | 75 | 571 | 108 | 47910545715 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 35 | 3 | 85 | 357558318 | 59.06% |
| NG | 1108 | 12922 | 874 | 604651366526 | 98.29% |
| NI | 168 | 1731 | 122 | 95171111238 | 98.22% |
| NL | 23505 | 214475 | 498419 | 8408267123317733 | 99.94% |
| NO | 3254 | 63881 | 3045 | 10968956647248 | 97.31% |
| NP | 469 | 2390 | 343 | 430195425381 | 94.05% |
| NR | 40 | 52 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2322 | 28858 | 1785 | 1359417250130 | 94.87% |
| OM | 302 | 4416 | 203 | 595072376919 | 96.04% |
| PA | 538 | 7090 | 466 | 429360149813 | 92.79% |
| PE | 1247 | 14659 | 1078 | 1072415514128 | 96.36% |
| PF | 91 | 312 | 144 | 21875058781 | 99.30% |
| PG | 257 | 347 | 153 | 73449136460 | 88.27% |
| PH | 2239 | 26824 | 1250 | 1042106851843 | 97.89% |
| PK | 1460 | 23632 | 747 | 907295961341 | 97.95% |
| PL | 8638 | 83038 | 5360 | 28587705921861 | 96.77% |
| PM | 42 | 25 | 80 | 4671531309 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 644 | 4576 | 337 | 230687463561 | 99.07% |
| PS | 308 | 3308 | 178 | 979424835919 | 93.84% |
| PT | 1921 | 27054 | 2036 | 3328529136919 | 93.01% |
| PW | 46 | 43 | 87 | 9085575215 | 97.53% |
| PY | 433 | 5086 | 270 | 451733561976 | 97.52% |
| QA | 502 | 5193 | 367 | 526445600733 | 96.23% |
| RE | 350 | 1994 | 100 | 95622438962 | 96.44% |
| RO | 3950 | 28292 | 2565 | 5569630562863 | 95.61% |
| RS | 1071 | 9073 | 1361 | 1852984526516 | 98.28% |
| RU | 13701 | 180655 | 15753 | 54549023383455 | 90.18% |
| RW | 486 | 1265 | 125 | 1147559424301 | 99.91% |
| SA | 1656 | 46667 | 920 | 2815427171816 | 96.60% |
| SB | 54 | 71 | 87 | 17675643175 | 90.67% |
| SC | 940 | 2675 | 3234 | 3322843882029 | 62.00% |
| SD | 114 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16257 | 126459 | 61456 | 30397158376708 | 79.89% |
| SG | 10983 | 111634 | 52549 | 133650159873103 | 98.93% |
| SH | 6 | 2 | 130 | 574482436 | 54.36% |
| SI | 1066 | 10331 | 662 | 2113802876348 | 95.29% |
| SJ | 7 | 2 | 40 | 374466579 | 60.68% |
| SK | 1214 | 11161 | 928 | 3009684245777 | 97.07% |
| SL | 125 | 358 | 111 | 26719021103 | 95.80% |
| SM | 73 | 173 | 98 | 202275613222 | 92.49% |
| SN | 119 | 1645 | 149 | 27161196614 | 98.08% |
| SO | 75 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 341 | 107 | 17813725490 | 98.45% |
| SS | 100 | 135 | 109 | 31098290221 | 98.66% |
| ST | 47 | 60 | 90 | 4773832747 | 96.05% |
| SV | 333 | 2776 | 207 | 211465050984 | 97.09% |
| SX | 44 | 129 | 80 | 30441400359 | 94.79% |
| SY | 296 | 5171 | 138 | 876931813423 | 98.86% |
| SZ | 76 | 242 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 85 | 1586850089 | 68.46% |
| TD | 93 | 133 | 104 | 13816293422 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2301 | 43596 | 1264 | 595720415568 | 95.57% |
| TJ | 157 | 401 | 168 | 641204726333 | 98.14% |
| TK | 25 | 15 | 114 | 4789889072 | 95.34% |
| TL | 87 | 125 | 109 | 43596832814 | 94.46% |
| TM | 63 | 126 | 103 | 69249132857 | 99.65% |
| TN | 148 | 30817 | 148 | 116780507711 | 97.52% |
| TO | 48 | 56 | 85 | 17523466282 | 99.25% |
| TR | 3733 | 68050 | 2906 | 22555794868122 | 93.52% |
| TT | 167 | 2226 | 145 | 1177257585228 | 70.56% |
| TV | 39 | 46 | 98 | 11246495783 | 82.44% |
| TW | 2588 | 147547 | 3110 | 11552662605885 | 98.66% |
| TZ | 425 | 4497 | 248 | 267366409332 | 98.77% |
| UA | 4872 | 34792 | 3348 | 11855375688690 | 92.81% |
| UG | 270 | 5773 | 184 | 132375003454 | 96.35% |
| UM | 5 | 2 | 83 | 355002413 | 58.92% |
| US | 133491 | 5884453 | 540217 | 4835780354492361 | 99.76% |
| UY | 245 | 10117 | 193 | 177623867459 | 99.80% |
| UZ | 548 | 1662 | 278 | 1083134229105 | 99.45% |
| VA | 36 | 51 | 106 | 103325444397 | 99.88% |
| VC | 82 | 103 | 89 | 69500512298 | 99.67% |
| VE | 975 | 30283 | 658 | 2279265730999 | 99.32% |
| VG | 323 | 272 | 420 | 251178960947 | 90.61% |
| VI | 89 | 547 | 159 | 30642686022 | 94.69% |
| VN | 2257 | 65541 | 1666 | 1890259178513 | 98.40% |
| VU | 65 | 86 | 282 | 51968273201 | 79.57% |
| WF | 21 | 16 | 85 | 9083414570 | 69.15% |
| WS | 60 | 97 | 90 | 21837242667 | 92.81% |
| XK | 192 | 624 | 76 | 111807037430 | 60.60% |
| YE | 113 | 958 | 119 | 207096833082 | 99.81% |
| YT | 52 | 46 | 76 | 602988585 | 56.00% |
| ZA | 5183 | 117560 | 3403 | 41032072713506 | 78.31% |
| ZM | 348 | 6508 | 159 | 48500502867 | 96.06% |
| ZW | 261 | 934 | 197 | 65582730601 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
