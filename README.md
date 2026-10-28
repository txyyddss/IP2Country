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

Last run: `2026-10-28T04:03:20Z`

Input rows: **38998502** · Published blocks: **2346972**

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
| Server-Country via sapics IPv4 | fresh | 400333 | 0 | 0 | 0 | 8727443 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 473080 | 0 | 1 | 15569792 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 400118 | 0 | 0 | 0 | 5957054 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 472705 | 0 | 1 | 18390087 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 417458 | 0 | 0 | 0 | 9108112 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 564988 | 0 | 1 | 16784569 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 722354 | 0 | 0 | 2131 | 9625457 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 289024 | 0 | 3994 | 8486655 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 322551 | 0 | 0 | 0 | 7052454 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 88064 | 0 | 0 | 3825073 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 984656 | 0 | 0 | 0 | 13037850 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2801616 | 0 | 0 | 89713386 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1344090 | 3030875 | 0 | 0 | 214611571 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 787331 | 1369927 | 0 | 0 | 17265480 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177189 | 69746 | 0 | 0 | 3396313 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 275042 | 422535 | 0 | 97303 | 36916021 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 275042 | 422535 | 0 | 97812 | 16647933 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4322 | 1680 | 0 | 0 | 97740 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3882 | 0 | 0 | 0 | 82389 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327613 | 0 | 0 | 0 | 537556 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 811454 | 427307 | 0 | 410930 | 18454568 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262363 | 71007 | 0 | 10 | 34586847 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802825 | 0 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6523 | 3497 | 0 | 0 | 159433 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 564963 | 499997 | 0 | 0 | 3403332 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 601251 | 3702257408 | 1745471 | 293421847207842436812546030166054273024 |
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
| AD | 213 | 307 | 188 | 43781127495 | 91.72% |
| AE | 2880 | 25481 | 2569 | 12826160038724 | 89.83% |
| AF | 232 | 878 | 520 | 121053017998 | 97.45% |
| AG | 119 | 177 | 184 | 40375548219 | 65.28% |
| AI | 44 | 46 | 102 | 5830802746 | 67.48% |
| AL | 588 | 1422 | 1262 | 2500709067418 | 95.22% |
| AM | 337 | 2307 | 524 | 1259266681300 | 94.71% |
| AO | 597 | 5132 | 300 | 147600101985 | 97.71% |
| AQ | 55 | 4 | 302 | 469331087 | 43.60% |
| AR | 3742 | 76933 | 2903 | 23613545820660 | 98.82% |
| AS | 39 | 28 | 334 | 9241098699 | 96.06% |
| AT | 4693 | 45108 | 4674 | 10476642428288 | 97.51% |
| AU | 12995 | 219010 | 8113 | 2305920540271643886 | 100.00% |
| AW | 49 | 449 | 98 | 11516819767 | 93.04% |
| AX | 92 | 132 | 608 | 69601594443 | 74.62% |
| AZ | 399 | 3069 | 351 | 1934087189380 | 99.30% |
| BA | 340 | 3264 | 745 | 838171684636 | 93.53% |
| BB | 190 | 508 | 106 | 69382748976 | 99.71% |
| BD | 2641 | 8176 | 2716 | 3577608893923 | 96.42% |
| BE | 8554 | 58086 | 73318 | 8440498296040578 | 66.66% |
| BF | 124 | 1382 | 144 | 91261127222 | 99.42% |
| BG | 2378 | 16720 | 1442 | 5172162259008 | 95.72% |
| BH | 320 | 5245 | 322 | 406107403346 | 97.65% |
| BI | 58 | 167 | 109 | 26902590248 | 92.88% |
| BJ | 83 | 620 | 113 | 56564881968 | 96.86% |
| BL | 42 | 10 | 84 | 1029962031 | 57.76% |
| BM | 321 | 498 | 115 | 43678792248 | 99.00% |
| BN | 108 | 878 | 120 | 26232795439 | 95.82% |
| BO | 337 | 4759 | 311 | 168267936620 | 97.23% |
| BQ | 36 | 103 | 74 | 17538087218 | 98.90% |
| BR | 11512 | 335972 | 42727 | 37432216120718 | 99.15% |
| BS | 135 | 632 | 165 | 13972409177 | 96.80% |
| BT | 124 | 226 | 165 | 60945265233 | 99.38% |
| BV | 1 | 1 | 3 | 1048577 | 50.00% |
| BW | 196 | 717 | 124 | 78595310127 | 97.52% |
| BY | 409 | 7402 | 428 | 700179084598 | 99.59% |
| BZ | 210 | 475 | 198 | 192696613700 | 89.03% |
| CA | 19489 | 285311 | 30865 | 6346775370651 | 95.56% |
| CC | 6 | 0 | 51 | 842922262 | 43.18% |
| CD | 315 | 932 | 203 | 155930247762 | 98.74% |
| CF | 49 | 40 | 98 | 1050231082 | 60.02% |
| CG | 89 | 585 | 119 | 36970586417 | 97.19% |
| CH | 6883 | 98331 | 57149 | 15878800857897356844 | 99.98% |
| CI | 355 | 6842 | 169 | 194903530040 | 98.93% |
| CK | 33 | 41 | 85 | 9217637679 | 96.15% |
| CL | 2021 | 41392 | 2395 | 1415225277995 | 96.62% |
| CM | 191 | 2438 | 160 | 67915819083 | 98.17% |
| CN | 6993 | 1347948 | 13044 | 294817155704508 | 99.94% |
| CO | 3137 | 70219 | 1949 | 1977320195215 | 97.69% |
| CR | 597 | 9171 | 396 | 461687196789 | 97.36% |
| CU | 99 | 1022 | 131 | 26422788668 | 99.21% |
| CV | 81 | 240 | 139 | 13663992110 | 97.38% |
| CW | 140 | 720 | 101 | 63107381549 | 92.44% |
| CX | 7 | 0 | 77 | 726513958 | 41.65% |
| CY | 681 | 4145 | 695 | 5999830445155 | 93.92% |
| CZ | 3771 | 33864 | 6784 | 13290450343286 | 97.58% |
| DE | 37899 | 551144 | 82824 | 115252087431030 | 96.06% |
| DJ | 71 | 243 | 117 | 9808902189 | 94.09% |
| DK | 3708 | 50112 | 2089 | 9667298891174 | 93.46% |
| DM | 74 | 71 | 116 | 44384058160 | 73.09% |
| DO | 739 | 8066 | 653 | 928490784381 | 98.99% |
| DZ | 226 | 18725 | 261 | 44143311292 | 93.89% |
| EC | 1123 | 11254 | 670 | 1158522081388 | 98.69% |
| EE | 1186 | 6515 | 1827 | 2730064729510 | 95.97% |
| EG | 800 | 94820 | 376 | 17730956292767 | 99.99% |
| EH | 4 | 0 | 55 | 457177368 | 46.78% |
| ER | 46 | 29 | 93 | 998656297 | 59.40% |
| ES | 11205 | 139849 | 6969 | 26207440854206 | 97.13% |
| ET | 95 | 1478 | 132 | 1117513173302 | 99.97% |
| FI | 4108 | 58619 | 29390 | 6857212024336 | 96.71% |
| FJ | 110 | 594 | 119 | 43363073068 | 99.51% |
| FK | 38 | 26 | 81 | 544143654 | 44.62% |
| FM | 56 | 49 | 104 | 8953067833 | 88.23% |
| FO | 48 | 181 | 90 | 108270320171 | 99.54% |
| FR | 28101 | 325640 | 20921 | 67346431678991 | 98.51% |
| GA | 89 | 1559 | 103 | 35744991528 | 94.32% |
| GB | 40415 | 499377 | 58390 | 106037062865120 | 94.08% |
| GD | 101 | 128 | 101 | 82688014890 | 96.90% |
| GE | 483 | 5108 | 347 | 1797448223097 | 95.68% |
| GF | 280 | 258 | 82 | 23008899368 | 97.40% |
| GG | 146 | 261 | 95 | 146471778862 | 77.66% |
| GH | 373 | 8926 | 262 | 1232395749218 | 99.79% |
| GI | 148 | 327 | 112 | 198008895786 | 91.23% |
| GL | 107 | 145 | 168 | 43603849595 | 99.30% |
| GM | 76 | 1102 | 109 | 39552387625 | 99.04% |
| GN | 75 | 172 | 114 | 48242362665 | 97.28% |
| GP | 206 | 788 | 102 | 54391032119 | 73.08% |
| GQ | 56 | 95 | 92 | 9487272233 | 95.90% |
| GR | 1543 | 23126 | 999 | 2592056378396 | 96.46% |
| GS | 4 | 2 | 41 | 389148949 | 59.18% |
| GT | 665 | 2887 | 287 | 327885949818 | 99.39% |
| GU | 161 | 991 | 125 | 34908795451 | 99.32% |
| GW | 50 | 42 | 95 | 9319500075 | 96.98% |
| GY | 57 | 288 | 109 | 35006118189 | 99.18% |
| HK | 13528 | 94986 | 7928 | 5142576185822 | 85.94% |
| HM | 1 | 2 | 9 | 18096128 | 49.27% |
| HN | 833 | 1992 | 264 | 391906802561 | 94.77% |
| HR | 849 | 9687 | 436 | 1591501315598 | 96.38% |
| HT | 89 | 624 | 134 | 26450983744 | 93.37% |
| HU | 1956 | 23372 | 1352 | 3217193385652 | 98.25% |
| ID | 7052 | 85349 | 5089 | 3792028778877 | 98.45% |
| IE | 8310 | 72145 | 8167 | 4921994182403 | 89.13% |
| IL | 2462 | 39267 | 1547 | 4302329801414 | 97.44% |
| IM | 349 | 473 | 194 | 361538861887 | 96.14% |
| IN | 13433 | 197204 | 16866 | 28991899058050 | 97.78% |
| IO | 33 | 15 | 82 | 4737857834 | 96.08% |
| IQ | 681 | 3342 | 976 | 3415550739616 | 96.90% |
| IR | 1964 | 42727 | 2215 | 16859158400141 | 98.53% |
| IS | 339 | 3642 | 526 | 1605333898367 | 96.23% |
| IT | 10061 | 227831 | 9526 | 46978083451391 | 99.09% |
| JE | 224 | 643 | 121 | 219655158317 | 95.45% |
| JM | 254 | 1132 | 158 | 1122124888925 | 99.96% |
| JO | 445 | 2705 | 253 | 1121935489373 | 97.90% |
| JP | 12609 | 782284 | 14527 | 52720737312601 | 88.45% |
| KE | 1165 | 25125 | 651 | 679647033784 | 98.42% |
| KG | 234 | 1179 | 221 | 529224088428 | 99.25% |
| KH | 559 | 2011 | 306 | 434970882921 | 81.10% |
| KI | 77 | 35 | 133 | 13519288652 | 92.47% |
| KM | 61 | 51 | 87 | 9353097509 | 96.74% |
| KN | 103 | 132 | 103 | 5747112504 | 77.10% |
| KP | 9 | 11 | 419 | 499466806 | 55.62% |
| KR | 3973 | 455234 | 3756 | 22750793135931 | 99.65% |
| KW | 418 | 7567 | 214 | 876851815104 | 99.77% |
| KY | 131 | 439 | 107 | 17845780019 | 99.15% |
| KZ | 1320 | 13962 | 1211 | 3150624081714 | 86.62% |
| LA | 149 | 451 | 154 | 121244152367 | 88.18% |
| LB | 390 | 2387 | 291 | 3015892022341 | 97.85% |
| LC | 89 | 167 | 88 | 78374435364 | 97.54% |
| LI | 377 | 453 | 205 | 429998738729 | 93.40% |
| LK | 267 | 2327 | 268 | 82539459192 | 92.58% |
| LR | 79 | 438 | 105 | 27241253419 | 97.76% |
| LS | 71 | 507 | 92 | 26751353129 | 98.32% |
| LT | 1479 | 12221 | 963 | 4368449820878 | 84.92% |
| LU | 1550 | 6706 | 869 | 2188410848879 | 96.22% |
| LV | 1241 | 8162 | 774 | 3235809242589 | 65.68% |
| LY | 151 | 1923 | 109 | 99581329724 | 89.62% |
| MA | 347 | 48129 | 268 | 1171116196773 | 99.78% |
| MC | 281 | 284 | 114 | 142217116462 | 99.87% |
| MD | 766 | 4957 | 975 | 5931591679571 | 64.24% |
| ME | 148 | 823 | 136 | 348170439737 | 99.48% |
| MF | 57 | 103 | 88 | 5359122734 | 92.88% |
| MG | 88 | 2308 | 117 | 22676603700 | 98.05% |
| MH | 71 | 90 | 81 | 4789645610 | 95.44% |
| MK | 250 | 2798 | 190 | 1142897917037 | 96.63% |
| ML | 85 | 388 | 89 | 22640603437 | 97.58% |
| MM | 461 | 1167 | 341 | 353154101886 | 97.56% |
| MN | 242 | 739 | 286 | 133921304944 | 99.24% |
| MO | 107 | 1500 | 699 | 39132300344 | 99.13% |
| MP | 70 | 17 | 76 | 4771148076 | 95.56% |
| MQ | 299 | 784 | 70 | 104798763299 | 89.70% |
| MR | 70 | 232 | 95 | 26483030314 | 98.94% |
| MS | 30 | 15 | 77 | 626850351 | 51.57% |
| MT | 319 | 2677 | 229 | 501360950327 | 93.67% |
| MU | 2051 | 23846 | 8181 | 4455196508455483 | 100.00% |
| MV | 118 | 401 | 140 | 18197508407 | 97.47% |
| MW | 124 | 2200 | 146 | 1242314408240 | 99.85% |
| MX | 4439 | 122273 | 2524 | 3638438943822 | 96.65% |
| MY | 2818 | 34668 | 1607 | 916000314432 | 95.64% |
| MZ | 196 | 1911 | 136 | 69652267324 | 97.38% |
| NA | 408 | 1854 | 186 | 29151347245 | 99.39% |
| NC | 74 | 571 | 108 | 47875881524 | 99.52% |
| NE | 61 | 217 | 93 | 22121672997 | 98.71% |
| NF | 33 | 2 | 78 | 358742317 | 59.47% |
| NG | 1108 | 12936 | 876 | 604634331520 | 98.29% |
| NI | 176 | 1731 | 123 | 95154338118 | 98.22% |
| NL | 23855 | 215047 | 501060 | 8408807916921943 | 99.94% |
| NO | 3213 | 63860 | 2878 | 10943233017936 | 97.37% |
| NP | 472 | 2389 | 347 | 434489348453 | 93.37% |
| NR | 40 | 53 | 87 | 8917678381 | 98.61% |
| NU | 27 | 11 | 75 | 4671404328 | 96.82% |
| NZ | 2295 | 29286 | 1781 | 1359948268046 | 95.35% |
| OM | 304 | 4418 | 205 | 595054621015 | 96.04% |
| PA | 536 | 7093 | 472 | 429343245882 | 92.60% |
| PE | 1292 | 14886 | 1095 | 1072377638944 | 96.46% |
| PF | 91 | 312 | 143 | 21857106268 | 99.35% |
| PG | 263 | 349 | 155 | 73449140811 | 91.78% |
| PH | 2256 | 26812 | 1253 | 1046351426308 | 97.62% |
| PK | 1435 | 23639 | 752 | 907362157052 | 97.65% |
| PL | 8652 | 83062 | 5334 | 28443791540308 | 97.01% |
| PM | 44 | 25 | 78 | 4670356012 | 96.79% |
| PN | 6 | 3 | 41 | 355531030 | 60.85% |
| PR | 657 | 4576 | 347 | 230653913744 | 99.07% |
| PS | 308 | 3298 | 177 | 979424774735 | 93.84% |
| PT | 1899 | 27023 | 2043 | 3328580013587 | 92.98% |
| PW | 48 | 43 | 82 | 9084268845 | 97.54% |
| PY | 444 | 5087 | 274 | 451683365755 | 97.34% |
| QA | 483 | 5168 | 372 | 526342861020 | 96.21% |
| RE | 348 | 1981 | 102 | 95638171953 | 96.44% |
| RO | 3989 | 28311 | 2517 | 5569649007412 | 95.60% |
| RS | 1062 | 9073 | 1358 | 1853001247923 | 98.28% |
| RU | 13592 | 180642 | 15750 | 54501825966248 | 90.27% |
| RW | 459 | 1263 | 125 | 1147541537324 | 99.91% |
| SA | 1656 | 46657 | 924 | 2815426323943 | 96.13% |
| SB | 54 | 71 | 87 | 17657887270 | 94.77% |
| SC | 1040 | 2825 | 3234 | 3319656411932 | 61.84% |
| SD | 115 | 7412 | 110 | 69686719027 | 97.39% |
| SE | 16150 | 126289 | 57090 | 30392290058460 | 79.87% |
| SG | 10771 | 107646 | 52115 | 151211297324417 | 97.13% |
| SH | 6 | 2 | 132 | 590805251 | 54.52% |
| SI | 1043 | 10325 | 661 | 2113819497145 | 95.56% |
| SJ | 7 | 2 | 42 | 390265106 | 60.69% |
| SK | 1222 | 11162 | 922 | 2980508016140 | 97.23% |
| SL | 124 | 357 | 110 | 26701134126 | 95.84% |
| SM | 71 | 172 | 101 | 202294426405 | 92.49% |
| SN | 121 | 1644 | 152 | 27193706821 | 98.05% |
| SO | 74 | 226 | 162 | 48199434816 | 99.16% |
| SR | 51 | 341 | 109 | 17812615730 | 98.46% |
| SS | 101 | 135 | 109 | 31080468780 | 98.68% |
| ST | 47 | 60 | 91 | 4772854058 | 96.06% |
| SV | 339 | 2775 | 212 | 211430563810 | 97.10% |
| SX | 44 | 129 | 78 | 30423513382 | 94.82% |
| SY | 295 | 5167 | 141 | 842622275888 | 99.62% |
| SZ | 76 | 242 | 105 | 30358061367 | 99.16% |
| TC | 45 | 56 | 81 | 1569815080 | 68.78% |
| TD | 91 | 132 | 105 | 13816232237 | 97.15% |
| TF | 2 | 0 | 37 | 305525010 | 59.88% |
| TG | 71 | 1402 | 111 | 26516146726 | 98.87% |
| TH | 2299 | 43569 | 1264 | 595602586202 | 95.58% |
| TJ | 148 | 392 | 172 | 641220590397 | 97.91% |
| TK | 24 | 15 | 109 | 4789041455 | 95.39% |
| TL | 88 | 126 | 105 | 43595919662 | 96.44% |
| TM | 63 | 126 | 106 | 69265914425 | 99.64% |
| TN | 150 | 30817 | 148 | 116746957630 | 97.54% |
| TO | 47 | 55 | 86 | 17540182313 | 99.20% |
| TR | 3704 | 68023 | 2893 | 22495379527838 | 93.51% |
| TT | 167 | 2226 | 151 | 1177239829329 | 70.56% |
| TV | 41 | 46 | 91 | 11247679782 | 88.42% |
| TW | 2555 | 147477 | 3086 | 11501070180688 | 98.79% |
| TZ | 404 | 4496 | 251 | 267349570933 | 98.78% |
| UA | 4889 | 34788 | 3293 | 11795260432321 | 93.12% |
| UG | 271 | 5773 | 185 | 132373893694 | 95.54% |
| UM | 5 | 2 | 73 | 322763052 | 60.31% |
| US | 134318 | 5881641 | 500172 | 4835697320913318 | 99.76% |
| UY | 248 | 10380 | 193 | 177573474373 | 99.81% |
| UZ | 554 | 1659 | 281 | 1083166673776 | 99.45% |
| VA | 35 | 51 | 108 | 103343274539 | 99.87% |
| VC | 82 | 104 | 88 | 69482625321 | 99.68% |
| VE | 979 | 30273 | 720 | 2262052442875 | 99.11% |
| VG | 345 | 556 | 410 | 251162925362 | 87.77% |
| VI | 85 | 532 | 184 | 30583835973 | 92.45% |
| VN | 2264 | 65536 | 1672 | 1860176618778 | 99.18% |
| VU | 65 | 86 | 282 | 51950517296 | 79.82% |
| WF | 22 | 16 | 69 | 4770756905 | 95.50% |
| WS | 60 | 97 | 92 | 21852058154 | 92.79% |
| XK | 198 | 621 | 73 | 111805005814 | 60.60% |
| YE | 116 | 958 | 121 | 207080387642 | 99.82% |
| YT | 42 | 36 | 79 | 635433257 | 56.23% |
| ZA | 5232 | 118002 | 3461 | 23439516608036 | 99.54% |
| ZM | 344 | 6509 | 158 | 48499327570 | 93.12% |
| ZW | 261 | 933 | 201 | 65564843624 | 98.33% |

<!-- GEOIP-STATISTICS:END -->
