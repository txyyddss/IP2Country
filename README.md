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

Last run: `2026-10-26T03:16:15Z`

Input rows: **38963901** · Published blocks: **2347927**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 597286 | 0 | 0 | 0 | 11032548 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 785668 | 0 | 0 | 20177287 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 597286 | 0 | 0 | 0 | 7580021 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 785668 | 0 | 0 | 24060852 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562055 | 0 | 0 | 0 | 11089776 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 499124 | 0 | 0 | 14990890 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564038 | 0 | 0 | 158 | 11815050 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 499207 | 0 | 88 | 34451179 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 399791 | 0 | 0 | 0 | 8712547 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 472523 | 0 | 1 | 15557062 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 399726 | 0 | 0 | 0 | 5950606 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 472301 | 0 | 1 | 18367179 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 416904 | 0 | 0 | 0 | 9092935 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 564429 | 0 | 1 | 16771719 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 721939 | 0 | 0 | 2143 | 9618635 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 289086 | 0 | 3995 | 8477476 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 322116 | 0 | 0 | 0 | 7041865 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 87881 | 0 | 0 | 3816211 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 978654 | 0 | 0 | 0 | 12966588 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2708038 | 0 | 0 | 85202442 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1337977 | 2937794 | 0 | 0 | 208656999 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 786527 | 1368454 | 0 | 0 | 17247264 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177163 | 69733 | 0 | 0 | 3395659 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 290366 | 492397 | 0 | 97495 | 41029325 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 290366 | 492397 | 0 | 98065 | 18562569 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4328 | 1676 | 0 | 0 | 97770 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3880 | 0 | 0 | 0 | 82329 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327613 | 0 | 0 | 0 | 537556 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 811519 | 427393 | 0 | 411163 | 18451739 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262329 | 70996 | 0 | 10 | 34582134 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802825 | 0 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6525 | 3488 | 0 | 0 | 159289 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 574 | 0 | 0 | 36004 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 564963 | 499997 | 0 | 0 | 3403332 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 601473 | 3702257408 | 1746204 | 293421847207842436812546030166054273024 |
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
| AD | 212 | 307 | 188 | 43781127495 | 91.72% |
| AE | 2857 | 25471 | 2563 | 12826428474176 | 89.83% |
| AF | 232 | 878 | 520 | 121053017998 | 97.45% |
| AG | 120 | 177 | 184 | 40375548219 | 65.28% |
| AI | 44 | 46 | 102 | 5830802746 | 67.48% |
| AL | 569 | 1421 | 1262 | 2500709067418 | 95.22% |
| AM | 333 | 2306 | 527 | 1259266681303 | 94.71% |
| AO | 555 | 5126 | 297 | 143305134687 | 98.64% |
| AQ | 63 | 6 | 296 | 469396616 | 43.60% |
| AR | 3742 | 76924 | 2915 | 23609250788245 | 98.82% |
| AS | 39 | 28 | 332 | 9241098697 | 96.06% |
| AT | 4691 | 45121 | 4637 | 10476626752782 | 97.51% |
| AU | 13006 | 219003 | 8211 | 2305920540272086282 | 100.00% |
| AW | 49 | 449 | 98 | 11516819767 | 93.04% |
| AX | 101 | 132 | 608 | 69601594443 | 74.62% |
| AZ | 398 | 3068 | 346 | 1934087189375 | 99.30% |
| BA | 340 | 3264 | 744 | 803811946268 | 94.68% |
| BB | 190 | 508 | 105 | 69382748975 | 99.71% |
| BD | 2635 | 8174 | 2722 | 3577608893952 | 96.42% |
| BE | 8522 | 58086 | 73323 | 8440506885909643 | 66.66% |
| BF | 125 | 1382 | 144 | 91261127222 | 99.42% |
| BG | 2367 | 16720 | 1441 | 5172162259008 | 95.72% |
| BH | 320 | 5245 | 255 | 406107403345 | 97.65% |
| BI | 58 | 167 | 109 | 26902590248 | 92.88% |
| BJ | 83 | 620 | 113 | 56564881968 | 96.86% |
| BL | 42 | 10 | 84 | 1029962031 | 57.76% |
| BM | 316 | 498 | 115 | 43678792248 | 99.00% |
| BN | 109 | 878 | 120 | 26232795439 | 95.82% |
| BO | 336 | 4758 | 312 | 168267936621 | 97.23% |
| BQ | 36 | 103 | 74 | 17538087218 | 98.90% |
| BR | 11490 | 335972 | 43002 | 37432203406920 | 99.15% |
| BS | 135 | 632 | 165 | 13972409177 | 96.80% |
| BT | 124 | 226 | 165 | 60945265233 | 99.38% |
| BV | 4 | 1 | 3 | 1048577 | 50.00% |
| BW | 182 | 717 | 125 | 78595310128 | 97.52% |
| BY | 408 | 7402 | 429 | 700179084599 | 99.59% |
| BZ | 210 | 475 | 198 | 192696613700 | 89.03% |
| CA | 19388 | 285275 | 31139 | 6346741955196 | 95.55% |
| CC | 6 | 0 | 51 | 842922262 | 43.18% |
| CD | 313 | 932 | 203 | 155930247762 | 98.74% |
| CF | 49 | 40 | 98 | 1050231082 | 60.02% |
| CG | 89 | 585 | 119 | 36970586417 | 97.19% |
| CH | 6926 | 98843 | 57198 | 15878800879350653365 | 99.98% |
| CI | 348 | 6837 | 168 | 194903530039 | 98.93% |
| CK | 33 | 41 | 85 | 9217637679 | 96.15% |
| CL | 2023 | 41392 | 2394 | 1415225294385 | 96.62% |
| CM | 191 | 2438 | 161 | 67915819084 | 98.17% |
| CN | 7029 | 1347939 | 13082 | 294817155704526 | 99.94% |
| CO | 3141 | 70219 | 1950 | 1977320178832 | 97.73% |
| CR | 597 | 9171 | 396 | 461687196789 | 97.36% |
| CU | 99 | 1022 | 132 | 26422788669 | 99.21% |
| CV | 81 | 240 | 139 | 13663992110 | 97.38% |
| CW | 140 | 720 | 100 | 62973163821 | 92.55% |
| CX | 7 | 0 | 77 | 726513958 | 41.65% |
| CY | 688 | 4146 | 694 | 5999830445154 | 93.92% |
| CZ | 3770 | 33870 | 6768 | 13290450376572 | 97.58% |
| DE | 38030 | 550997 | 81448 | 115252641966494 | 96.05% |
| DJ | 70 | 243 | 117 | 9808902189 | 94.09% |
| DK | 3713 | 50113 | 2087 | 9667298825636 | 93.46% |
| DM | 74 | 71 | 115 | 44384058159 | 73.09% |
| DO | 739 | 8066 | 652 | 928490784380 | 98.99% |
| DZ | 226 | 18725 | 259 | 44143311290 | 93.89% |
| EC | 1122 | 11253 | 670 | 1158522081388 | 98.69% |
| EE | 1174 | 6515 | 1726 | 2730065773796 | 95.97% |
| EG | 799 | 94820 | 375 | 17730956358301 | 99.99% |
| EH | 4 | 0 | 55 | 457177368 | 46.78% |
| ER | 46 | 29 | 93 | 998656297 | 59.40% |
| ES | 11206 | 139842 | 6924 | 26207440919678 | 97.13% |
| ET | 95 | 1478 | 132 | 1117513173302 | 99.97% |
| FI | 4135 | 58643 | 28930 | 6857211863567 | 96.61% |
| FJ | 110 | 594 | 119 | 43363073068 | 99.51% |
| FK | 38 | 26 | 81 | 544143654 | 44.62% |
| FM | 56 | 49 | 106 | 8953067835 | 88.23% |
| FO | 48 | 181 | 91 | 108270320172 | 99.54% |
| FR | 28098 | 325600 | 20410 | 67346430898069 | 98.51% |
| GA | 89 | 1559 | 102 | 35744991527 | 94.32% |
| GB | 40703 | 498865 | 57928 | 106033847050863 | 94.08% |
| GD | 101 | 128 | 101 | 82688014890 | 96.90% |
| GE | 488 | 5108 | 345 | 1797448223095 | 95.68% |
| GF | 235 | 251 | 82 | 23008899368 | 97.40% |
| GG | 137 | 263 | 96 | 146471778863 | 81.57% |
| GH | 373 | 8926 | 262 | 1232395749217 | 99.79% |
| GI | 148 | 327 | 113 | 198008895787 | 91.23% |
| GL | 108 | 145 | 171 | 43603915133 | 99.30% |
| GM | 76 | 1102 | 111 | 39552387627 | 99.04% |
| GN | 75 | 172 | 114 | 48242362665 | 97.28% |
| GP | 231 | 795 | 102 | 54391032119 | 73.08% |
| GQ | 56 | 95 | 92 | 9487272233 | 95.90% |
| GR | 1535 | 23127 | 997 | 2592006046748 | 96.46% |
| GS | 4 | 2 | 41 | 389148949 | 59.18% |
| GT | 670 | 2887 | 286 | 327885949817 | 99.39% |
| GU | 158 | 990 | 125 | 34908795451 | 99.32% |
| GW | 50 | 42 | 95 | 9319500075 | 96.98% |
| GY | 57 | 288 | 110 | 35006118190 | 99.18% |
| HK | 13598 | 94959 | 7961 | 5142575611862 | 85.83% |
| HM | 1 | 2 | 7 | 17965056 | 49.38% |
| HN | 831 | 1992 | 264 | 391906802561 | 94.76% |
| HR | 864 | 9687 | 436 | 1591501315598 | 96.38% |
| HT | 89 | 624 | 134 | 26450983744 | 93.37% |
| HU | 1939 | 23373 | 1340 | 3217193401768 | 98.25% |
| ID | 7053 | 85395 | 5085 | 3792028451197 | 98.45% |
| IE | 8318 | 72149 | 8170 | 4921994182406 | 89.13% |
| IL | 2451 | 39288 | 1533 | 4302329931957 | 97.44% |
| IM | 345 | 474 | 194 | 361538861887 | 96.14% |
| IN | 13418 | 197209 | 17967 | 28987603821051 | 97.91% |
| IO | 33 | 15 | 82 | 4737857834 | 96.08% |
| IQ | 681 | 3342 | 973 | 3415550739613 | 96.90% |
| IR | 1965 | 42725 | 2225 | 16859158416501 | 98.53% |
| IS | 339 | 3642 | 504 | 1605333849215 | 96.23% |
| IT | 10024 | 227832 | 9208 | 46978083385751 | 99.09% |
| JE | 223 | 641 | 121 | 219655158317 | 95.45% |
| JM | 254 | 1132 | 158 | 1122124888925 | 99.96% |
| JO | 445 | 2705 | 253 | 1121935489373 | 97.90% |
| JP | 12617 | 782300 | 14587 | 52720738655756 | 88.45% |
| KE | 1097 | 25118 | 646 | 671057099191 | 98.83% |
| KG | 234 | 1179 | 221 | 529224088428 | 99.25% |
| KH | 559 | 2011 | 306 | 434970882921 | 81.10% |
| KI | 78 | 35 | 132 | 13519288651 | 92.47% |
| KM | 61 | 51 | 87 | 9353097509 | 96.74% |
| KN | 116 | 164 | 103 | 5747112504 | 77.10% |
| KP | 9 | 11 | 416 | 499597873 | 55.62% |
| KR | 3966 | 455234 | 3754 | 22750793135952 | 99.65% |
| KW | 418 | 7567 | 213 | 876851815103 | 99.77% |
| KY | 126 | 439 | 107 | 17845780019 | 99.15% |
| KZ | 1326 | 13961 | 1207 | 3150624081455 | 86.62% |
| LA | 150 | 452 | 154 | 121244152367 | 88.18% |
| LB | 391 | 2388 | 291 | 3015892022341 | 98.08% |
| LC | 89 | 167 | 89 | 78374500900 | 97.54% |
| LI | 327 | 451 | 205 | 429998738729 | 93.40% |
| LK | 295 | 2327 | 268 | 82539459192 | 92.58% |
| LR | 79 | 437 | 105 | 27241253419 | 97.76% |
| LS | 71 | 507 | 92 | 26751353129 | 98.32% |
| LT | 1470 | 12218 | 959 | 4368449755342 | 84.92% |
| LU | 1549 | 6705 | 869 | 2188410848879 | 96.22% |
| LV | 1238 | 8164 | 773 | 3235809242589 | 65.68% |
| LY | 151 | 1923 | 109 | 99581329724 | 89.62% |
| MA | 346 | 48129 | 270 | 1171116196775 | 99.78% |
| MC | 281 | 284 | 114 | 142217116462 | 99.87% |
| MD | 772 | 4958 | 976 | 5931591679572 | 64.14% |
| ME | 148 | 823 | 136 | 348170439737 | 99.48% |
| MF | 57 | 103 | 88 | 5359122734 | 92.88% |
| MG | 88 | 2308 | 118 | 22676603701 | 98.05% |
| MH | 71 | 90 | 81 | 4789645610 | 95.44% |
| MK | 250 | 2798 | 190 | 1142897917037 | 96.63% |
| ML | 85 | 388 | 89 | 22640603437 | 97.58% |
| MM | 461 | 1166 | 340 | 353154101885 | 97.56% |
| MN | 242 | 739 | 286 | 133921321329 | 99.24% |
| MO | 107 | 1500 | 702 | 39132300347 | 99.13% |
| MP | 76 | 18 | 76 | 4771148076 | 95.56% |
| MQ | 308 | 819 | 70 | 104798763299 | 89.70% |
| MR | 70 | 232 | 95 | 26483030314 | 98.94% |
| MS | 30 | 15 | 77 | 626850351 | 51.57% |
| MT | 320 | 2678 | 229 | 501360950327 | 93.67% |
| MU | 2032 | 23886 | 8150 | 4455217983291963 | 100.00% |
| MV | 117 | 401 | 140 | 18197508407 | 97.47% |
| MW | 124 | 2200 | 146 | 1242314408240 | 99.85% |
| MX | 4392 | 122274 | 2756 | 3629849010520 | 96.72% |
| MY | 2862 | 34672 | 1473 | 911705346881 | 95.78% |
| MZ | 194 | 1911 | 136 | 69652267324 | 97.38% |
| NA | 358 | 1856 | 186 | 29151347245 | 99.39% |
| NC | 74 | 571 | 108 | 47875881524 | 99.52% |
| NE | 61 | 217 | 93 | 22121672997 | 98.71% |
| NF | 33 | 2 | 79 | 358742318 | 59.47% |
| NG | 1112 | 12938 | 882 | 604634397057 | 98.29% |
| NI | 176 | 1731 | 123 | 95154338118 | 98.22% |
| NL | 23882 | 215107 | 501309 | 8408915274000200 | 99.94% |
| NO | 3219 | 63858 | 2847 | 10943233165389 | 97.37% |
| NP | 472 | 2389 | 347 | 434489348453 | 93.37% |
| NR | 40 | 53 | 87 | 8917678381 | 98.61% |
| NU | 27 | 11 | 76 | 4671404329 | 96.82% |
| NZ | 2293 | 29285 | 1780 | 1359949316620 | 95.35% |
| OM | 304 | 4418 | 205 | 595054621015 | 96.04% |
| PA | 537 | 7101 | 473 | 429343311418 | 92.60% |
| PE | 1294 | 14883 | 1099 | 1072377720866 | 96.46% |
| PF | 88 | 312 | 142 | 21857106267 | 99.35% |
| PG | 263 | 349 | 155 | 73449140811 | 91.78% |
| PH | 2250 | 26769 | 1309 | 1046351426103 | 97.62% |
| PK | 1428 | 23635 | 748 | 907362091513 | 97.65% |
| PL | 8640 | 83053 | 5345 | 28443875031891 | 97.03% |
| PM | 44 | 25 | 78 | 4670356012 | 96.79% |
| PN | 5 | 2 | 41 | 355531030 | 60.85% |
| PR | 665 | 4576 | 346 | 230653880976 | 99.07% |
| PS | 308 | 3298 | 177 | 979424774735 | 93.84% |
| PT | 1901 | 27023 | 2067 | 3328580079057 | 92.98% |
| PW | 48 | 43 | 82 | 9084268845 | 97.54% |
| PY | 444 | 5087 | 274 | 451683365755 | 97.34% |
| QA | 489 | 5169 | 372 | 526342861020 | 96.21% |
| RE | 358 | 1994 | 103 | 95638171954 | 96.44% |
| RO | 4051 | 28313 | 2517 | 5569648810805 | 95.60% |
| RS | 1048 | 9071 | 1400 | 1853001231015 | 98.28% |
| RU | 13565 | 180645 | 15574 | 54467518821273 | 90.30% |
| RW | 199 | 1264 | 125 | 1147541537324 | 99.91% |
| SA | 1653 | 46656 | 924 | 2815426340300 | 96.13% |
| SB | 54 | 71 | 87 | 17657887270 | 94.77% |
| SC | 1070 | 2832 | 3351 | 3336783065884 | 61.78% |
| SD | 115 | 7412 | 110 | 69686719027 | 97.39% |
| SE | 16101 | 126269 | 57194 | 30392290210248 | 79.88% |
| SG | 10776 | 107653 | 52167 | 151211297336070 | 97.12% |
| SH | 6 | 2 | 132 | 590805251 | 54.52% |
| SI | 1083 | 10325 | 661 | 2113819497145 | 95.56% |
| SJ | 7 | 2 | 42 | 390265106 | 60.69% |
| SK | 1221 | 11164 | 843 | 2980508016140 | 97.23% |
| SL | 126 | 356 | 110 | 26701134126 | 95.84% |
| SM | 71 | 172 | 101 | 202294426405 | 92.49% |
| SN | 121 | 1644 | 152 | 27193706821 | 98.05% |
| SO | 74 | 226 | 162 | 48199434816 | 99.16% |
| SR | 51 | 341 | 109 | 17812615730 | 98.46% |
| SS | 101 | 135 | 109 | 31080468780 | 98.68% |
| ST | 47 | 60 | 91 | 4772854058 | 96.06% |
| SV | 344 | 2774 | 212 | 211430563810 | 97.10% |
| SX | 44 | 129 | 78 | 30423513382 | 94.82% |
| SY | 295 | 5167 | 141 | 842622275888 | 99.62% |
| SZ | 76 | 242 | 105 | 30358061367 | 99.16% |
| TC | 45 | 56 | 82 | 1569815081 | 68.78% |
| TD | 91 | 132 | 105 | 13816232237 | 97.15% |
| TF | 2 | 0 | 37 | 305525010 | 59.88% |
| TG | 71 | 1402 | 111 | 26516146726 | 98.87% |
| TH | 2295 | 43566 | 1282 | 595602585964 | 95.58% |
| TJ | 148 | 392 | 171 | 636925623101 | 98.12% |
| TK | 24 | 15 | 110 | 4789041456 | 95.39% |
| TL | 88 | 126 | 104 | 43595854126 | 96.44% |
| TM | 62 | 126 | 106 | 69265914425 | 99.64% |
| TN | 150 | 30817 | 148 | 116746957630 | 97.54% |
| TO | 47 | 55 | 87 | 17540182314 | 99.20% |
| TR | 3740 | 68022 | 2877 | 22512559396958 | 93.53% |
| TT | 167 | 2226 | 151 | 1177239829329 | 70.56% |
| TV | 41 | 46 | 91 | 11247679782 | 88.42% |
| TW | 2551 | 147478 | 3078 | 11501070048589 | 98.79% |
| TZ | 357 | 4488 | 251 | 267349570933 | 98.78% |
| UA | 4977 | 34793 | 3321 | 11790964416670 | 93.24% |
| UG | 271 | 5773 | 185 | 132373893694 | 95.54% |
| UM | 5 | 2 | 73 | 322828587 | 60.31% |
| US | 134547 | 5881701 | 501980 | 4835635624411196 | 99.76% |
| UY | 248 | 10380 | 192 | 177573474372 | 99.81% |
| UZ | 556 | 1659 | 277 | 1083166673772 | 99.45% |
| VA | 35 | 51 | 107 | 103209056811 | 99.95% |
| VC | 82 | 104 | 89 | 69482625322 | 99.68% |
| VE | 978 | 30273 | 722 | 2253462508287 | 99.23% |
| VG | 344 | 556 | 407 | 251162925362 | 87.77% |
| VI | 85 | 532 | 184 | 30583835973 | 92.45% |
| VN | 2264 | 65537 | 1666 | 1860176708883 | 99.18% |
| VU | 65 | 86 | 282 | 51950517296 | 79.82% |
| WF | 22 | 16 | 69 | 4770756905 | 95.50% |
| WS | 60 | 97 | 92 | 21852058154 | 92.79% |
| XK | 189 | 623 | 73 | 111805005814 | 60.60% |
| YE | 116 | 958 | 122 | 207080387643 | 99.82% |
| YT | 42 | 45 | 78 | 635433256 | 56.23% |
| ZA | 5227 | 117984 | 3456 | 23435221689936 | 99.55% |
| ZM | 357 | 6506 | 157 | 44204360274 | 95.69% |
| ZW | 257 | 931 | 200 | 65564843623 | 98.33% |

<!-- GEOIP-STATISTICS:END -->
