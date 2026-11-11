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

Last run: `2026-11-11T04:45:28Z`

Input rows: **39718508** · Published blocks: **2411283**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562781 | 0 | 0 | 0 | 11107464 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 494891 | 0 | 0 | 15044030 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564711 | 0 | 0 | 0 | 11825981 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 494974 | 0 | 0 | 34153206 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 403303 | 0 | 0 | 0 | 8799288 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 476404 | 0 | 1 | 15623319 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 403061 | 0 | 0 | 0 | 6006202 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 476594 | 0 | 1 | 18474198 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 420647 | 0 | 0 | 0 | 9184330 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 568474 | 0 | 1 | 16843574 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1035624 | 0 | 0 | 0 | 13644516 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2953826 | 0 | 0 | 96435228 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1394650 | 3186718 | 0 | 0 | 225498699 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 389960 | 0 | 0 | 0 | 6095165 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1556723 | 0 | 0 | 24472437 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 799745 | 1396371 | 0 | 0 | 17576344 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177288 | 69834 | 0 | 0 | 3401089 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 291905 | 488730 | 0 | 97927 | 40911070 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 291905 | 488730 | 0 | 98442 | 18490598 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4334 | 1725 | 0 | 0 | 98632 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3879 | 0 | 0 | 0 | 82497 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 812933 | 427543 | 0 | 412098 | 18498275 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262568 | 71101 | 0 | 10 | 34618068 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6204 | 3422 | 0 | 0 | 153395 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 565492 | 495713 | 0 | 0 | 3392927 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 612331 | 3702257408 | 1798702 | 293421847207842436812546030166054273024 |
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
| AD | 214 | 305 | 183 | 43728628810 | 92.57% |
| AE | 2843 | 25518 | 2579 | 12835044520657 | 89.85% |
| AF | 219 | 875 | 526 | 121136965522 | 97.42% |
| AG | 107 | 168 | 183 | 40393304123 | 65.26% |
| AI | 44 | 45 | 101 | 5847641145 | 67.40% |
| AL | 580 | 1415 | 1244 | 2500156283802 | 95.22% |
| AM | 351 | 2325 | 519 | 1263612958926 | 95.01% |
| AO | 561 | 5140 | 297 | 147599966561 | 98.68% |
| AQ | 56 | 5 | 299 | 472128652 | 43.65% |
| AR | 3783 | 77006 | 2828 | 23622416457951 | 98.82% |
| AS | 41 | 29 | 309 | 9243191472 | 96.04% |
| AT | 4754 | 45184 | 4467 | 10471298196840 | 97.54% |
| AU | 13342 | 220213 | 71455 | 2309962504777081945 | 100.00% |
| AW | 49 | 449 | 98 | 11534575673 | 92.96% |
| AX | 88 | 132 | 626 | 69552487243 | 72.57% |
| AZ | 399 | 3070 | 341 | 1963968586370 | 98.31% |
| BA | 334 | 3264 | 742 | 833827495454 | 95.13% |
| BB | 193 | 507 | 109 | 69400586958 | 99.69% |
| BD | 2659 | 8183 | 2714 | 3577593881841 | 96.48% |
| BE | 8495 | 57835 | 4824 | 4398024048089734 | 66.61% |
| BF | 122 | 1386 | 144 | 91245525303 | 99.43% |
| BG | 2411 | 16718 | 1435 | 5137769043776 | 96.07% |
| BH | 330 | 5246 | 320 | 406108644177 | 97.65% |
| BI | 57 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 86 | 623 | 127 | 56582572337 | 96.84% |
| BL | 41 | 11 | 85 | 1047717936 | 57.46% |
| BM | 326 | 498 | 107 | 43426867512 | 99.37% |
| BN | 110 | 879 | 119 | 26250616879 | 99.06% |
| BO | 326 | 4751 | 307 | 168302404201 | 98.07% |
| BQ | 38 | 103 | 69 | 17538738225 | 98.90% |
| BR | 11502 | 336045 | 42491 | 37419670339773 | 99.23% |
| BS | 139 | 633 | 163 | 14003518038 | 96.66% |
| BT | 125 | 226 | 164 | 60978750033 | 99.35% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 205 | 718 | 123 | 78628794672 | 97.43% |
| BY | 389 | 7400 | 426 | 700163287346 | 99.59% |
| BZ | 219 | 469 | 198 | 192713403210 | 89.02% |
| CA | 22083 | 285487 | 34350 | 6351235604222 | 95.81% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 317 | 939 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 119 | 36988538929 | 97.17% |
| CH | 6886 | 98183 | 57431 | 15878800938989027928 | 99.98% |
| CI | 348 | 6712 | 170 | 194921285946 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2067 | 41384 | 2383 | 1419535904446 | 96.52% |
| CM | 189 | 2439 | 160 | 67900151884 | 98.18% |
| CN | 6937 | 1347604 | 12689 | 294809138755020 | 99.94% |
| CO | 3176 | 70318 | 1946 | 1981632206735 | 98.03% |
| CR | 600 | 9164 | 368 | 461701217133 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 82 | 241 | 144 | 13715367983 | 97.22% |
| CW | 139 | 718 | 101 | 63125202989 | 92.43% |
| CX | 8 | 0 | 83 | 727509031 | 41.43% |
| CY | 699 | 4161 | 706 | 5999831227235 | 94.03% |
| CZ | 3892 | 33852 | 6441 | 13269500134560 | 97.63% |
| DE | 38004 | 550790 | 84522 | 115481271411885 | 96.04% |
| DJ | 73 | 246 | 119 | 9641322543 | 94.96% |
| DK | 3646 | 50106 | 2080 | 9667282682796 | 93.43% |
| DM | 74 | 71 | 116 | 44401879600 | 73.08% |
| DO | 761 | 8074 | 625 | 928508573027 | 98.99% |
| DZ | 232 | 18723 | 256 | 44176730297 | 95.48% |
| EC | 1088 | 11258 | 657 | 1158571649927 | 98.76% |
| EE | 1201 | 6498 | 2236 | 2725776303447 | 96.02% |
| EG | 793 | 94816 | 371 | 17726611071391 | 100.00% |
| EH | 4 | 0 | 55 | 457959450 | 46.78% |
| ER | 44 | 29 | 91 | 966146091 | 59.37% |
| ES | 11292 | 139920 | 7003 | 26203111969510 | 97.15% |
| ET | 92 | 1478 | 133 | 1117547837495 | 99.96% |
| FI | 4044 | 58672 | 33569 | 6900128486150 | 96.65% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 38 | 25 | 83 | 612231206 | 46.31% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 86 | 108187543852 | 99.57% |
| FR | 27356 | 325665 | 21093 | 67447895007558 | 98.48% |
| GA | 91 | 1559 | 107 | 35779721513 | 94.29% |
| GB | 40409 | 498479 | 61273 | 105871457070219 | 94.11% |
| GD | 98 | 128 | 103 | 82705836587 | 96.88% |
| GE | 488 | 5117 | 355 | 1831824816249 | 95.38% |
| GF | 313 | 257 | 80 | 23010009128 | 97.39% |
| GG | 143 | 261 | 93 | 112129796398 | 81.66% |
| GH | 372 | 8927 | 264 | 1232412718691 | 99.85% |
| GI | 151 | 329 | 105 | 197924939819 | 90.89% |
| GL | 103 | 144 | 169 | 43621736571 | 99.27% |
| GM | 75 | 1103 | 109 | 39586920746 | 98.99% |
| GN | 76 | 167 | 116 | 48293804074 | 97.24% |
| GP | 236 | 808 | 102 | 54408919095 | 73.06% |
| GQ | 53 | 95 | 92 | 9454958635 | 96.02% |
| GR | 1523 | 23125 | 998 | 2592058471182 | 96.52% |
| GS | 4 | 2 | 40 | 406904853 | 59.19% |
| GT | 667 | 2886 | 276 | 327768439410 | 99.41% |
| GU | 190 | 995 | 129 | 34757730620 | 99.56% |
| GW | 53 | 42 | 94 | 9320609836 | 96.97% |
| GY | 57 | 288 | 108 | 35023939628 | 96.09% |
| HK | 13742 | 95151 | 8043 | 5095541165391 | 85.87% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 846 | 1991 | 266 | 391924755057 | 94.76% |
| HR | 837 | 9689 | 431 | 1591451110613 | 96.38% |
| HT | 83 | 624 | 133 | 26468805184 | 93.34% |
| HU | 1994 | 23375 | 1301 | 3217142965699 | 98.33% |
| ID | 6908 | 81280 | 5292 | 3843600756852 | 98.03% |
| IE | 8298 | 72340 | 8953 | 4956485318838 | 88.87% |
| IL | 2475 | 37163 | 1559 | 4293572876907 | 97.46% |
| IM | 349 | 476 | 189 | 361489639997 | 96.14% |
| IN | 13404 | 197347 | 16789 | 29004690387120 | 97.79% |
| IO | 27 | 14 | 83 | 4755679275 | 95.94% |
| IQ | 673 | 3347 | 981 | 3415584420776 | 96.90% |
| IR | 1971 | 42731 | 2210 | 16962237647755 | 98.49% |
| IS | 338 | 3638 | 492 | 1596995666811 | 96.34% |
| IT | 9974 | 227758 | 9576 | 47051084569843 | 99.09% |
| JE | 223 | 643 | 115 | 219604756781 | 95.45% |
| JM | 251 | 1131 | 152 | 1122142710361 | 99.96% |
| JO | 444 | 2702 | 253 | 1121936402780 | 97.90% |
| JP | 12532 | 782498 | 14836 | 53257924209820 | 88.23% |
| KE | 1074 | 25123 | 662 | 688052414647 | 99.34% |
| KG | 236 | 1184 | 220 | 529225198189 | 99.38% |
| KH | 559 | 2015 | 300 | 430692688487 | 81.58% |
| KI | 79 | 36 | 136 | 13520201802 | 97.75% |
| KM | 66 | 52 | 85 | 9337298983 | 96.78% |
| KN | 100 | 133 | 103 | 5765065014 | 76.99% |
| KP | 9 | 11 | 421 | 500236592 | 55.29% |
| KR | 4031 | 455045 | 3660 | 22712187953995 | 99.76% |
| KW | 425 | 7581 | 217 | 877741019583 | 99.72% |
| KY | 124 | 438 | 105 | 17861635377 | 99.08% |
| KZ | 1403 | 13976 | 1211 | 3150641399341 | 86.82% |
| LA | 148 | 450 | 151 | 121260925231 | 88.17% |
| LB | 392 | 2388 | 293 | 3015926620999 | 98.08% |
| LC | 96 | 168 | 90 | 78392322341 | 97.53% |
| LI | 297 | 452 | 201 | 429947419690 | 93.40% |
| LK | 251 | 2325 | 265 | 82540585332 | 94.31% |
| LR | 82 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1451 | 12015 | 964 | 4177275226000 | 84.79% |
| LU | 1550 | 6703 | 881 | 2222679373433 | 96.39% |
| LV | 1252 | 8162 | 776 | 3235775257820 | 65.66% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 340 | 48132 | 266 | 1171100463782 | 99.78% |
| MC | 284 | 284 | 108 | 142166649391 | 99.88% |
| MD | 760 | 4973 | 973 | 5953066380626 | 63.94% |
| ME | 153 | 828 | 129 | 348120038202 | 99.48% |
| MF | 56 | 103 | 88 | 5443266607 | 92.31% |
| MG | 85 | 2307 | 114 | 22644159028 | 98.09% |
| MH | 84 | 94 | 81 | 4806418475 | 95.24% |
| MK | 253 | 2800 | 186 | 1142897781613 | 97.05% |
| ML | 83 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 236 | 737 | 286 | 133921418861 | 99.23% |
| MO | 105 | 1500 | 712 | 39116173884 | 99.15% |
| MP | 94 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 312 | 812 | 70 | 104815667236 | 89.69% |
| MR | 75 | 233 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 76 | 644671790 | 51.26% |
| MT | 324 | 2681 | 230 | 509630147384 | 92.89% |
| MU | 1958 | 23828 | 8192 | 4455188187869500 | 100.00% |
| MV | 120 | 403 | 135 | 18180595767 | 97.48% |
| MW | 123 | 2200 | 146 | 1242315518001 | 99.85% |
| MX | 4411 | 122451 | 2521 | 3642801423150 | 96.96% |
| MY | 2835 | 34574 | 1627 | 911924937803 | 96.08% |
| MZ | 197 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 329 | 1857 | 186 | 29169103150 | 99.37% |
| NC | 77 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1100 | 12922 | 710 | 604382963839 | 98.31% |
| NI | 168 | 1731 | 119 | 95171111235 | 98.22% |
| NL | 24876 | 214437 | 496796 | 8407906457325069 | 99.94% |
| NO | 3295 | 63893 | 3018 | 10986135402288 | 97.26% |
| NP | 467 | 2389 | 343 | 430195425381 | 94.05% |
| NR | 41 | 52 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2376 | 28872 | 1788 | 1359417233746 | 94.87% |
| OM | 314 | 4416 | 202 | 595072376918 | 96.04% |
| PA | 536 | 7089 | 469 | 429360248116 | 92.96% |
| PE | 1251 | 14660 | 1077 | 1072415464975 | 96.36% |
| PF | 93 | 312 | 143 | 21875058780 | 99.30% |
| PG | 233 | 347 | 153 | 73449136460 | 87.10% |
| PH | 2258 | 26805 | 1271 | 1046404637187 | 97.68% |
| PK | 1446 | 23637 | 742 | 907296026871 | 97.95% |
| PL | 8669 | 83072 | 5359 | 28613407393103 | 96.79% |
| PM | 42 | 25 | 80 | 4671531309 | 96.85% |
| PN | 6 | 3 | 42 | 355461142 | 60.41% |
| PR | 642 | 4572 | 336 | 230687497097 | 99.07% |
| PS | 304 | 3313 | 178 | 979424835919 | 93.84% |
| PT | 1912 | 27057 | 1987 | 3328261700824 | 93.01% |
| PW | 47 | 43 | 87 | 9085575215 | 97.53% |
| PY | 434 | 5086 | 270 | 451733561976 | 97.52% |
| QA | 497 | 5195 | 366 | 526445600732 | 96.23% |
| RE | 356 | 2019 | 100 | 95622438962 | 96.44% |
| RO | 3928 | 28288 | 2512 | 5569029794279 | 95.61% |
| RS | 1080 | 9077 | 1350 | 1887339595192 | 97.41% |
| RU | 13729 | 180696 | 15988 | 54600553848656 | 90.18% |
| RW | 199 | 1265 | 125 | 1147559424301 | 99.91% |
| SA | 1660 | 48715 | 920 | 2815427188195 | 96.58% |
| SB | 55 | 71 | 99 | 17944013095 | 90.06% |
| SC | 874 | 2562 | 3192 | 3361213964844 | 61.73% |
| SD | 115 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16286 | 126468 | 61361 | 30393102610933 | 79.92% |
| SG | 11029 | 111039 | 52670 | 133650226995551 | 98.93% |
| SH | 6 | 2 | 65 | 574416917 | 54.37% |
| SI | 1068 | 10335 | 621 | 2079455836088 | 95.52% |
| SJ | 7 | 2 | 40 | 374466579 | 60.68% |
| SK | 1220 | 11164 | 932 | 3009684262164 | 97.04% |
| SL | 157 | 358 | 111 | 26584803376 | 96.08% |
| SM | 73 | 173 | 98 | 202275613222 | 92.49% |
| SN | 128 | 1645 | 149 | 27161196614 | 98.08% |
| SO | 75 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 341 | 107 | 17813725490 | 98.45% |
| SS | 100 | 135 | 110 | 31098290222 | 98.66% |
| ST | 49 | 60 | 90 | 4773832747 | 96.05% |
| SV | 338 | 2776 | 208 | 211465050985 | 97.09% |
| SX | 48 | 128 | 80 | 30441400359 | 94.79% |
| SY | 296 | 5171 | 138 | 876931813423 | 99.64% |
| SZ | 76 | 242 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 82 | 1451583784 | 71.10% |
| TD | 93 | 133 | 104 | 13816293422 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2293 | 43585 | 1270 | 595719563862 | 95.52% |
| TJ | 158 | 402 | 167 | 641204726332 | 98.14% |
| TK | 25 | 15 | 114 | 4789889072 | 95.34% |
| TL | 87 | 125 | 109 | 43596832814 | 92.49% |
| TM | 62 | 126 | 102 | 69114915129 | 99.77% |
| TN | 148 | 30817 | 148 | 116780507711 | 97.52% |
| TO | 49 | 56 | 85 | 17523466282 | 99.25% |
| TR | 3590 | 68037 | 2906 | 22512863163256 | 93.82% |
| TT | 167 | 2226 | 144 | 1177257585227 | 70.56% |
| TV | 40 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9267 | 160470 | 3113 | 11552662460702 | 98.66% |
| TZ | 396 | 4498 | 249 | 267366409333 | 98.77% |
| UA | 4890 | 34657 | 3326 | 11756620406506 | 93.12% |
| UG | 270 | 5773 | 183 | 132375003453 | 96.35% |
| UM | 5 | 2 | 81 | 355002411 | 58.92% |
| US | 135281 | 5871089 | 545104 | 4835783628012796 | 99.76% |
| UY | 245 | 10116 | 193 | 177623867459 | 99.80% |
| UZ | 548 | 1660 | 277 | 1083134229104 | 99.45% |
| VA | 36 | 51 | 107 | 103325509933 | 99.88% |
| VC | 84 | 103 | 89 | 69500512298 | 99.67% |
| VE | 969 | 30270 | 658 | 2279265730999 | 99.33% |
| VG | 327 | 514 | 401 | 251178944570 | 90.61% |
| VI | 89 | 547 | 157 | 30643800133 | 94.69% |
| VN | 2256 | 65523 | 1664 | 1860194006032 | 99.18% |
| VU | 67 | 87 | 274 | 51951561521 | 80.96% |
| WF | 22 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 197 | 623 | 73 | 111806840822 | 60.60% |
| YE | 113 | 958 | 119 | 207096833082 | 99.81% |
| YT | 52 | 58 | 75 | 602988584 | 56.00% |
| ZA | 5221 | 117599 | 3391 | 41031802197322 | 78.31% |
| ZM | 376 | 6515 | 161 | 48484905299 | 96.08% |
| ZW | 291 | 938 | 196 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
