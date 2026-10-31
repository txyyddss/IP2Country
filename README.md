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

Last run: `2026-10-31T04:05:46Z`

Input rows: **38970332** · Published blocks: **2348879**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 597286 | 0 | 0 | 0 | 11032548 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 785668 | 0 | 0 | 20177287 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 597286 | 0 | 0 | 0 | 7580021 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 785668 | 0 | 0 | 24060852 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562236 | 0 | 0 | 0 | 11103523 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 498000 | 0 | 0 | 15010491 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564227 | 0 | 0 | 0 | 11815981 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 498101 | 0 | 0 | 34368969 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 401945 | 0 | 0 | 0 | 8766214 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 473565 | 0 | 1 | 15581532 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 400847 | 0 | 0 | 0 | 5968721 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 473392 | 0 | 1 | 18410718 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 419040 | 0 | 0 | 0 | 9146819 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 565473 | 0 | 1 | 16796309 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 722358 | 0 | 0 | 2115 | 9630001 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 289612 | 0 | 3991 | 8503842 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 322312 | 0 | 0 | 0 | 7046816 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 88221 | 0 | 0 | 3831719 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 987418 | 0 | 0 | 0 | 13060185 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2798765 | 0 | 0 | 89559654 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1346386 | 3029581 | 0 | 0 | 214519552 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 787733 | 1371164 | 0 | 0 | 17278592 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177222 | 69779 | 0 | 0 | 3397795 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 290142 | 422735 | 0 | 97886 | 37598030 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 290142 | 422735 | 0 | 98409 | 16939450 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4353 | 1682 | 0 | 0 | 98265 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3887 | 0 | 0 | 0 | 82509 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327613 | 0 | 0 | 0 | 537556 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 810793 | 363741 | 0 | 411612 | 18119989 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262422 | 71038 | 0 | 10 | 34596249 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802828 | 0 | 0 | 0 | 11122714 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6174 | 3475 | 0 | 0 | 153879 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 564994 | 498815 | 0 | 0 | 3395840 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 601765 | 3702257408 | 1746864 | 293421847207842436812546030166054273024 |
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
| AD | 213 | 307 | 188 | 43781127495 | 91.71% |
| AE | 2876 | 25467 | 2571 | 12830455071556 | 89.82% |
| AF | 231 | 878 | 520 | 121053017998 | 97.45% |
| AG | 125 | 177 | 184 | 40375548219 | 65.28% |
| AI | 44 | 46 | 102 | 5830802746 | 67.48% |
| AL | 582 | 1421 | 1261 | 2500742032026 | 95.22% |
| AM | 337 | 2307 | 524 | 1259266681300 | 94.71% |
| AO | 551 | 5130 | 300 | 147600101985 | 97.95% |
| AQ | 55 | 4 | 302 | 469314702 | 43.60% |
| AR | 3740 | 76933 | 2982 | 23617840804340 | 98.82% |
| AS | 39 | 28 | 334 | 9241098699 | 96.06% |
| AT | 4704 | 45102 | 4675 | 10476642428289 | 97.52% |
| AU | 12998 | 219021 | 8132 | 2305920544566508628 | 100.00% |
| AW | 49 | 449 | 98 | 11516819767 | 93.04% |
| AX | 92 | 132 | 608 | 69601594443 | 74.62% |
| AZ | 398 | 3068 | 351 | 1934087189380 | 99.30% |
| BA | 340 | 3264 | 745 | 838171684636 | 93.87% |
| BB | 190 | 508 | 106 | 69382748976 | 99.71% |
| BD | 2644 | 8181 | 2718 | 3577609024995 | 96.42% |
| BE | 8553 | 58091 | 73145 | 8440481116171391 | 66.66% |
| BF | 123 | 1382 | 144 | 91261127222 | 99.42% |
| BG | 2383 | 16718 | 1443 | 5172162259009 | 95.72% |
| BH | 329 | 5247 | 322 | 406107403346 | 97.65% |
| BI | 58 | 167 | 109 | 26902590248 | 92.88% |
| BJ | 83 | 620 | 113 | 56564881968 | 96.86% |
| BL | 42 | 10 | 84 | 1029962031 | 57.76% |
| BM | 322 | 498 | 115 | 43678792248 | 98.99% |
| BN | 108 | 878 | 120 | 26232795439 | 95.81% |
| BO | 336 | 4759 | 311 | 168267936620 | 97.23% |
| BQ | 33 | 103 | 74 | 17538087218 | 98.90% |
| BR | 11512 | 335965 | 42761 | 37445101022619 | 99.15% |
| BS | 135 | 632 | 165 | 13972458326 | 96.77% |
| BT | 124 | 226 | 165 | 60945265233 | 99.37% |
| BV | 1 | 1 | 3 | 1048577 | 50.00% |
| BW | 196 | 717 | 124 | 78595310127 | 97.52% |
| BY | 410 | 7402 | 428 | 700179084598 | 99.59% |
| BZ | 210 | 475 | 198 | 192696613700 | 89.03% |
| CA | 19561 | 285336 | 30870 | 6342480051419 | 95.58% |
| CC | 6 | 0 | 51 | 842922262 | 43.18% |
| CD | 337 | 932 | 203 | 155930247762 | 98.74% |
| CF | 49 | 40 | 98 | 1050231082 | 60.02% |
| CG | 89 | 585 | 119 | 36970586417 | 97.19% |
| CH | 6878 | 98322 | 57250 | 15878800840776846378 | 99.98% |
| CI | 353 | 6841 | 169 | 194903530040 | 98.93% |
| CK | 33 | 41 | 85 | 9217637679 | 96.15% |
| CL | 2022 | 41394 | 2404 | 1415225343531 | 96.62% |
| CM | 190 | 2438 | 160 | 67915819083 | 98.17% |
| CN | 6987 | 1347931 | 12959 | 294821433894587 | 99.94% |
| CO | 3137 | 70220 | 1951 | 1977320277135 | 97.69% |
| CR | 597 | 9171 | 396 | 461687196789 | 97.36% |
| CU | 99 | 1022 | 131 | 26422788668 | 99.21% |
| CV | 81 | 240 | 139 | 13663992110 | 97.36% |
| CW | 140 | 720 | 101 | 63107381549 | 92.44% |
| CX | 7 | 0 | 77 | 726513958 | 41.65% |
| CY | 678 | 4146 | 695 | 5999830445155 | 93.92% |
| CZ | 3821 | 33862 | 6788 | 13294744393078 | 97.57% |
| DE | 38115 | 551063 | 82941 | 115252238887473 | 96.05% |
| DJ | 71 | 243 | 117 | 9808902189 | 94.09% |
| DK | 3707 | 50107 | 2089 | 9667298891174 | 93.46% |
| DM | 75 | 71 | 116 | 44384058160 | 73.09% |
| DO | 736 | 8065 | 653 | 928490784381 | 98.99% |
| DZ | 225 | 18725 | 261 | 44143311292 | 93.89% |
| EC | 1123 | 11258 | 670 | 1158522081388 | 98.69% |
| EE | 1184 | 6521 | 1830 | 2730064748561 | 95.97% |
| EG | 798 | 94818 | 376 | 17730956292767 | 99.99% |
| EH | 4 | 0 | 55 | 457177368 | 46.78% |
| ER | 46 | 29 | 93 | 998656297 | 59.40% |
| ES | 11206 | 139848 | 6997 | 26207440919740 | 97.13% |
| ET | 94 | 1478 | 132 | 1117513173302 | 99.97% |
| FI | 4080 | 58615 | 29596 | 6857212048624 | 96.71% |
| FJ | 110 | 594 | 119 | 43363073068 | 99.50% |
| FK | 38 | 26 | 81 | 544143654 | 44.62% |
| FM | 56 | 49 | 104 | 8953067833 | 88.23% |
| FO | 48 | 181 | 90 | 108270320171 | 99.54% |
| FR | 28062 | 325581 | 20964 | 67346431760932 | 98.51% |
| GA | 89 | 1559 | 103 | 35744991528 | 94.32% |
| GB | 40318 | 498819 | 58393 | 106033009177065 | 94.06% |
| GD | 101 | 128 | 101 | 82688014890 | 96.90% |
| GE | 483 | 5108 | 348 | 1831807961465 | 94.82% |
| GF | 276 | 258 | 81 | 23008899367 | 97.40% |
| GG | 136 | 264 | 95 | 146471778862 | 78.14% |
| GH | 372 | 8926 | 262 | 1232395749218 | 99.79% |
| GI | 149 | 327 | 113 | 198008895787 | 91.23% |
| GL | 108 | 145 | 168 | 43603849595 | 99.30% |
| GM | 76 | 1102 | 109 | 39552387625 | 99.03% |
| GN | 75 | 172 | 114 | 48242362665 | 97.28% |
| GP | 216 | 796 | 102 | 54391032119 | 73.08% |
| GQ | 56 | 95 | 92 | 9487272233 | 95.90% |
| GR | 1544 | 23128 | 999 | 2592056378396 | 96.52% |
| GS | 4 | 2 | 41 | 389148949 | 59.18% |
| GT | 665 | 2887 | 287 | 327885949818 | 99.39% |
| GU | 161 | 993 | 125 | 34908795451 | 99.31% |
| GW | 50 | 42 | 95 | 9319500075 | 96.98% |
| GY | 57 | 288 | 109 | 35006118189 | 99.17% |
| HK | 13580 | 95000 | 7969 | 5142576054756 | 85.94% |
| HM | 1 | 2 | 9 | 18096128 | 49.27% |
| HN | 831 | 1992 | 264 | 391906802561 | 94.77% |
| HR | 853 | 9687 | 436 | 1591501315598 | 96.38% |
| HT | 89 | 624 | 134 | 26450983744 | 93.37% |
| HU | 1956 | 23373 | 1352 | 3217193385652 | 98.30% |
| ID | 7038 | 85374 | 5094 | 3796324008316 | 98.42% |
| IE | 8328 | 72144 | 8168 | 4921993133844 | 89.14% |
| IL | 2474 | 39254 | 1636 | 4302329804852 | 97.44% |
| IM | 349 | 473 | 194 | 361538861887 | 96.14% |
| IN | 13438 | 197215 | 16863 | 28996194459615 | 97.78% |
| IO | 33 | 15 | 82 | 4737857834 | 96.08% |
| IQ | 680 | 3342 | 976 | 3415550739616 | 96.90% |
| IR | 1965 | 42723 | 2215 | 16859158400141 | 98.53% |
| IS | 339 | 3642 | 526 | 1605333898367 | 96.23% |
| IT | 10064 | 227800 | 9524 | 47012443244673 | 99.05% |
| JE | 228 | 640 | 121 | 219655158317 | 95.45% |
| JM | 254 | 1132 | 158 | 1122124888925 | 99.96% |
| JO | 444 | 2705 | 253 | 1121935489373 | 97.90% |
| JP | 12632 | 782314 | 14539 | 52720721710813 | 88.45% |
| KE | 1167 | 25130 | 653 | 688236968376 | 98.18% |
| KG | 232 | 1176 | 221 | 529224088428 | 99.25% |
| KH | 563 | 2008 | 306 | 434970882921 | 81.10% |
| KI | 77 | 35 | 133 | 13519288652 | 92.47% |
| KM | 61 | 51 | 87 | 9353097509 | 96.71% |
| KN | 100 | 132 | 103 | 5747112504 | 77.10% |
| KP | 9 | 11 | 420 | 499532342 | 55.61% |
| KR | 3972 | 455243 | 3746 | 22750793103154 | 99.65% |
| KW | 422 | 7573 | 214 | 876851815104 | 99.77% |
| KY | 131 | 439 | 107 | 17845780019 | 99.14% |
| KZ | 1321 | 13960 | 1211 | 3150624081714 | 86.62% |
| LA | 149 | 451 | 154 | 121244152367 | 88.18% |
| LB | 390 | 2386 | 291 | 3015892022341 | 97.85% |
| LC | 94 | 167 | 88 | 78374435364 | 97.54% |
| LI | 377 | 453 | 205 | 429998738729 | 93.40% |
| LK | 255 | 2328 | 268 | 82539459192 | 93.01% |
| LR | 79 | 438 | 105 | 27241253419 | 97.75% |
| LS | 71 | 507 | 92 | 26751353129 | 98.32% |
| LT | 1474 | 12219 | 959 | 4368449755342 | 84.96% |
| LU | 1551 | 6707 | 869 | 2188408751727 | 96.29% |
| LV | 1240 | 8163 | 774 | 3235809242589 | 65.68% |
| LY | 150 | 1923 | 109 | 99581329724 | 89.62% |
| MA | 346 | 48129 | 268 | 1171116196773 | 99.77% |
| MC | 282 | 284 | 114 | 142217116462 | 99.87% |
| MD | 764 | 4956 | 974 | 5931591614035 | 64.05% |
| ME | 148 | 823 | 136 | 348170439737 | 99.47% |
| MF | 57 | 103 | 88 | 5359122734 | 92.88% |
| MG | 87 | 2308 | 117 | 22676603700 | 98.04% |
| MH | 71 | 90 | 81 | 4789645610 | 95.44% |
| MK | 250 | 2798 | 190 | 1142897917037 | 96.63% |
| ML | 84 | 388 | 89 | 22640603437 | 97.58% |
| MM | 460 | 1169 | 341 | 353154101886 | 97.56% |
| MN | 242 | 739 | 290 | 133921341295 | 99.24% |
| MO | 107 | 1500 | 699 | 39132300344 | 99.13% |
| MP | 68 | 15 | 76 | 4771148076 | 95.56% |
| MQ | 324 | 808 | 70 | 104798763299 | 89.70% |
| MR | 70 | 232 | 95 | 26483030314 | 98.93% |
| MS | 30 | 15 | 77 | 626850351 | 51.57% |
| MT | 320 | 2678 | 229 | 501360950327 | 93.67% |
| MU | 2078 | 23808 | 8189 | 4455187918520891 | 100.00% |
| MV | 117 | 401 | 140 | 18197508407 | 97.46% |
| MW | 124 | 2200 | 146 | 1242314408240 | 99.85% |
| MX | 4419 | 122261 | 2525 | 3638438943822 | 96.68% |
| MY | 2822 | 34667 | 1603 | 916000314432 | 95.68% |
| MZ | 195 | 1911 | 136 | 69652267324 | 97.38% |
| NA | 360 | 1856 | 186 | 29151347245 | 99.39% |
| NC | 74 | 571 | 107 | 47875947058 | 99.52% |
| NE | 61 | 217 | 93 | 22121672997 | 98.71% |
| NF | 33 | 2 | 78 | 358742317 | 59.47% |
| NG | 1109 | 12937 | 876 | 604634331520 | 98.29% |
| NI | 176 | 1731 | 123 | 95154338118 | 98.22% |
| NL | 23822 | 215021 | 501538 | 8408624035788193 | 99.94% |
| NO | 3224 | 63861 | 2878 | 10943233017936 | 97.37% |
| NP | 469 | 2387 | 347 | 434489348453 | 93.37% |
| NR | 40 | 53 | 87 | 8917678381 | 98.61% |
| NU | 27 | 11 | 75 | 4671404328 | 96.82% |
| NZ | 2293 | 29283 | 1785 | 1359948423759 | 95.35% |
| OM | 303 | 4416 | 205 | 595054621015 | 96.04% |
| PA | 536 | 7091 | 472 | 429343245882 | 92.60% |
| PE | 1286 | 14740 | 1096 | 1072377704480 | 96.46% |
| PF | 91 | 312 | 143 | 21857106268 | 99.35% |
| PG | 261 | 349 | 155 | 73449140811 | 91.78% |
| PH | 2258 | 26804 | 1253 | 1046351426308 | 97.62% |
| PK | 1436 | 23643 | 753 | 911657124348 | 97.54% |
| PL | 8681 | 83053 | 5380 | 28478151213111 | 96.94% |
| PM | 44 | 25 | 78 | 4670356012 | 96.79% |
| PN | 6 | 3 | 41 | 355531030 | 60.85% |
| PR | 649 | 4576 | 346 | 230653913488 | 99.07% |
| PS | 307 | 3296 | 177 | 979424774735 | 93.84% |
| PT | 1911 | 27033 | 2048 | 3328579653142 | 93.03% |
| PW | 48 | 43 | 82 | 9084268845 | 97.54% |
| PY | 444 | 5087 | 274 | 451683365755 | 97.34% |
| QA | 484 | 5171 | 372 | 526342861020 | 96.21% |
| RE | 362 | 1993 | 102 | 95638171953 | 96.44% |
| RO | 4003 | 28322 | 2557 | 5569681775412 | 95.62% |
| RS | 1064 | 9072 | 1358 | 1853001247923 | 98.28% |
| RU | 13617 | 180660 | 15760 | 54510399090858 | 90.26% |
| RW | 515 | 1263 | 125 | 1147541537324 | 99.91% |
| SA | 1658 | 46658 | 924 | 2815426323943 | 96.13% |
| SB | 54 | 71 | 87 | 17657887270 | 94.77% |
| SC | 1061 | 2858 | 3234 | 3319656411932 | 61.84% |
| SD | 114 | 7412 | 110 | 69686719027 | 97.39% |
| SE | 16175 | 126295 | 57049 | 30392307865200 | 79.87% |
| SG | 10765 | 107637 | 52349 | 151211293326730 | 97.13% |
| SH | 6 | 2 | 132 | 590805251 | 54.52% |
| SI | 1044 | 10325 | 662 | 2113819513529 | 95.56% |
| SJ | 7 | 2 | 42 | 390265106 | 60.69% |
| SK | 1223 | 11162 | 929 | 3018357415436 | 96.61% |
| SL | 124 | 357 | 110 | 26701134126 | 95.83% |
| SM | 71 | 172 | 101 | 202294426405 | 92.49% |
| SN | 119 | 1643 | 152 | 27193706821 | 98.04% |
| SO | 74 | 226 | 162 | 48199434816 | 99.15% |
| SR | 51 | 341 | 109 | 17812615730 | 98.45% |
| SS | 101 | 135 | 109 | 31080468780 | 98.68% |
| ST | 47 | 60 | 91 | 4772854058 | 96.06% |
| SV | 339 | 2775 | 212 | 211430563810 | 97.10% |
| SX | 44 | 129 | 78 | 30423513382 | 94.82% |
| SY | 295 | 5167 | 141 | 842622275888 | 99.62% |
| SZ | 76 | 242 | 105 | 30358061367 | 99.16% |
| TC | 45 | 56 | 81 | 1569815080 | 68.78% |
| TD | 91 | 132 | 105 | 13816232237 | 97.13% |
| TF | 2 | 0 | 37 | 305525010 | 59.88% |
| TG | 71 | 1402 | 111 | 26516146726 | 98.87% |
| TH | 2298 | 43576 | 1264 | 595602586202 | 95.58% |
| TJ | 148 | 392 | 172 | 641220590397 | 97.97% |
| TK | 24 | 15 | 109 | 4789041455 | 95.39% |
| TL | 88 | 126 | 105 | 43595919662 | 96.44% |
| TM | 63 | 126 | 106 | 69265914425 | 99.64% |
| TN | 148 | 30816 | 148 | 116746957630 | 97.54% |
| TO | 46 | 55 | 86 | 17540182313 | 99.20% |
| TR | 3730 | 68021 | 2886 | 22521149397149 | 93.45% |
| TT | 167 | 2226 | 151 | 1177239829329 | 70.56% |
| TV | 41 | 46 | 91 | 11247679782 | 88.42% |
| TW | 2555 | 147481 | 3095 | 11501070311759 | 98.79% |
| TZ | 409 | 4499 | 251 | 267349570933 | 98.78% |
| UA | 4906 | 34798 | 3310 | 11795260437745 | 93.13% |
| UG | 270 | 5773 | 185 | 132373893694 | 95.54% |
| UM | 5 | 2 | 73 | 322763052 | 60.31% |
| US | 134377 | 5882408 | 500212 | 4835692532713665 | 99.76% |
| UY | 247 | 10376 | 193 | 177573474373 | 99.81% |
| UZ | 556 | 1662 | 281 | 1083166673776 | 99.45% |
| VA | 35 | 51 | 108 | 103343274539 | 99.87% |
| VC | 87 | 104 | 88 | 69482625321 | 99.68% |
| VE | 987 | 30317 | 722 | 2270642377467 | 99.05% |
| VG | 349 | 556 | 410 | 251162925362 | 87.77% |
| VI | 83 | 532 | 185 | 30600613189 | 92.43% |
| VN | 2264 | 65544 | 1674 | 1860176684315 | 99.18% |
| VU | 65 | 86 | 282 | 51950517296 | 79.82% |
| WF | 22 | 16 | 69 | 4770756905 | 95.50% |
| WS | 60 | 97 | 92 | 21852058154 | 92.78% |
| XK | 192 | 622 | 74 | 111805071350 | 60.60% |
| YE | 115 | 958 | 121 | 207080387642 | 99.82% |
| YT | 42 | 38 | 79 | 635433257 | 56.23% |
| ZA | 5260 | 117945 | 3464 | 23439516624418 | 99.54% |
| ZM | 347 | 6509 | 158 | 48499327570 | 93.85% |
| ZW | 261 | 933 | 201 | 65564843624 | 98.33% |

<!-- GEOIP-STATISTICS:END -->
