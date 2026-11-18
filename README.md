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

Last run: `2026-11-18T05:56:50Z`

Input rows: **40588187** · Published blocks: **2396035**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 564311 | 0 | 0 | 0 | 11143126 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 494164 | 0 | 0 | 15054413 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 568077 | 0 | 0 | 0 | 11896529 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 495575 | 0 | 0 | 34194675 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 407158 | 0 | 0 | 0 | 8870352 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 484474 | 0 | 1 | 15978289 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 406667 | 0 | 0 | 0 | 6050054 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 483591 | 0 | 1 | 18814161 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 426011 | 0 | 0 | 0 | 9279634 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 578022 | 0 | 1 | 17243699 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1043521 | 0 | 0 | 0 | 13736087 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 3113188 | 0 | 0 | 102955245 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1402325 | 3346342 | 0 | 0 | 234667156 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 804518 | 1405546 | 0 | 0 | 17687928 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177450 | 69894 | 0 | 0 | 3404107 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 295462 | 427677 | 0 | 97511 | 38070079 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 295462 | 427677 | 0 | 98011 | 17160887 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4331 | 1726 | 0 | 0 | 98602 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3895 | 0 | 0 | 0 | 83157 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 815576 | 365234 | 0 | 411831 | 18100104 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262665 | 71160 | 0 | 10 | 34634336 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4958198 | 0 | 0 | 0 | 66567473 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380280 | 0 | 0 | 11696034 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6222 | 3441 | 0 | 0 | 154063 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 578 | 0 | 0 | 36064 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 566994 | 494983 | 0 | 0 | 3396552 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 614779 | 3702257408 | 1781006 | 293421847207842436812546030166054273024 |
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
| AD | 218 | 309 | 192 | 43728565322 | 92.57% |
| AE | 2838 | 25537 | 2535 | 12770686497504 | 90.03% |
| AF | 220 | 875 | 528 | 121136965524 | 97.42% |
| AG | 127 | 177 | 186 | 40393307195 | 65.26% |
| AI | 44 | 45 | 103 | 5847643193 | 67.40% |
| AL | 592 | 1413 | 1248 | 2508745105305 | 95.10% |
| AM | 358 | 2323 | 521 | 1297972697295 | 94.48% |
| AO | 548 | 5126 | 369 | 147599974753 | 98.68% |
| AQ | 56 | 5 | 305 | 472131726 | 43.65% |
| AR | 3801 | 77014 | 2775 | 23626711480169 | 98.86% |
| AS | 41 | 29 | 310 | 9243193519 | 96.04% |
| AT | 4743 | 45210 | 4401 | 10505922021051 | 97.35% |
| AU | 13448 | 219822 | 71534 | 2309962499421648992 | 100.00% |
| AW | 50 | 449 | 101 | 11534578745 | 92.96% |
| AX | 93 | 132 | 628 | 69552490314 | 72.57% |
| AZ | 399 | 3070 | 347 | 1963968455297 | 98.36% |
| BA | 333 | 3265 | 743 | 833827497501 | 95.13% |
| BB | 194 | 507 | 110 | 69400590028 | 99.69% |
| BD | 2668 | 8193 | 2711 | 3577593947376 | 96.48% |
| BE | 8434 | 58034 | 4867 | 4398028309735309 | 66.61% |
| BF | 126 | 1388 | 146 | 91245528374 | 99.43% |
| BG | 2385 | 16712 | 1437 | 5129179028290 | 96.15% |
| BH | 336 | 5248 | 319 | 406108709711 | 97.65% |
| BI | 59 | 168 | 110 | 26969694760 | 92.79% |
| BJ | 85 | 622 | 129 | 56582575408 | 96.84% |
| BL | 23 | 10 | 87 | 1047851055 | 57.46% |
| BM | 320 | 501 | 109 | 43426869560 | 99.37% |
| BN | 111 | 879 | 118 | 26250616878 | 99.06% |
| BO | 324 | 4753 | 306 | 168302404200 | 98.07% |
| BQ | 38 | 103 | 71 | 17538741296 | 98.90% |
| BR | 11574 | 335592 | 42476 | 37427992387384 | 99.22% |
| BS | 140 | 634 | 166 | 14004569685 | 96.65% |
| BT | 125 | 226 | 159 | 60963021392 | 99.36% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 188 | 720 | 122 | 78628794671 | 97.43% |
| BY | 402 | 7406 | 432 | 700163598642 | 99.49% |
| BZ | 224 | 470 | 200 | 192713406281 | 89.02% |
| CA | 22147 | 285205 | 36256 | 6373148562922 | 95.85% |
| CC | 7 | 0 | 51 | 827256853 | 42.84% |
| CD | 339 | 940 | 202 | 155963863378 | 98.74% |
| CF | 54 | 40 | 98 | 1117335593 | 59.87% |
| CG | 92 | 585 | 118 | 36988538928 | 97.17% |
| CH | 6946 | 98597 | 57525 | 15878800934688683597 | 99.98% |
| CI | 348 | 6714 | 171 | 194921289016 | 98.93% |
| CK | 35 | 41 | 93 | 9234282543 | 96.05% |
| CL | 2078 | 41388 | 2386 | 1419535904325 | 96.63% |
| CM | 189 | 2439 | 162 | 67900154955 | 98.18% |
| CN | 6990 | 1347812 | 12886 | 294811339830803 | 99.94% |
| CO | 3184 | 70295 | 1945 | 1985927242641 | 98.00% |
| CR | 590 | 9167 | 372 | 461702268526 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 82 | 242 | 146 | 13715370031 | 97.22% |
| CW | 139 | 718 | 103 | 63125206060 | 92.43% |
| CX | 8 | 0 | 82 | 727380006 | 41.43% |
| CY | 784 | 4142 | 710 | 6008425358176 | 94.02% |
| CZ | 3900 | 33828 | 6436 | 13303842292142 | 97.50% |
| DE | 38363 | 550865 | 84667 | 115514968818734 | 96.03% |
| DJ | 65 | 247 | 122 | 9641325615 | 94.96% |
| DK | 3648 | 50104 | 2097 | 9637218039722 | 93.45% |
| DM | 76 | 71 | 118 | 44401882671 | 73.08% |
| DO | 747 | 8092 | 629 | 932803543395 | 98.92% |
| DZ | 233 | 18724 | 253 | 44176666805 | 95.48% |
| EC | 1092 | 11267 | 662 | 1162867668871 | 98.69% |
| EE | 1198 | 6499 | 1650 | 2725760640279 | 96.02% |
| EG | 800 | 94821 | 290 | 17726610989471 | 100.00% |
| EH | 4 | 0 | 56 | 457961497 | 46.78% |
| ER | 46 | 29 | 94 | 966149163 | 59.37% |
| ES | 11211 | 139916 | 6948 | 26203110470925 | 97.12% |
| ET | 95 | 1482 | 134 | 1117547839542 | 99.96% |
| FI | 4112 | 58763 | 34062 | 6895831435561 | 96.70% |
| FJ | 114 | 590 | 119 | 43379783468 | 99.48% |
| FK | 39 | 26 | 85 | 612233254 | 46.30% |
| FM | 57 | 49 | 107 | 8921740345 | 78.74% |
| FO | 46 | 181 | 88 | 108187546923 | 99.57% |
| FR | 27559 | 325704 | 21133 | 67548567167589 | 98.44% |
| GA | 90 | 1558 | 109 | 35779724584 | 94.29% |
| GB | 40213 | 498285 | 61239 | 105927302332687 | 93.85% |
| GD | 101 | 128 | 104 | 82705774122 | 96.88% |
| GE | 495 | 5120 | 359 | 1831825930362 | 95.38% |
| GF | 259 | 255 | 80 | 23010009128 | 97.39% |
| GG | 142 | 263 | 96 | 112129799470 | 81.66% |
| GH | 360 | 8925 | 265 | 1232412982881 | 99.85% |
| GI | 149 | 327 | 108 | 197924942891 | 90.89% |
| GL | 103 | 144 | 168 | 43621736570 | 99.27% |
| GM | 75 | 1103 | 110 | 39586922793 | 98.99% |
| GN | 77 | 168 | 118 | 48293807145 | 97.24% |
| GP | 224 | 801 | 105 | 54408922167 | 73.07% |
| GQ | 52 | 94 | 95 | 9454961707 | 96.02% |
| GR | 1539 | 23120 | 1004 | 2592595350287 | 96.51% |
| GS | 4 | 2 | 40 | 406904853 | 59.19% |
| GT | 671 | 2899 | 278 | 327769491056 | 99.41% |
| GU | 182 | 995 | 128 | 34757730619 | 99.56% |
| GW | 53 | 42 | 95 | 9320611883 | 96.97% |
| GY | 58 | 288 | 109 | 35023941675 | 96.09% |
| HK | 13936 | 95303 | 8028 | 5091409340241 | 85.96% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 844 | 1995 | 268 | 391924758128 | 94.76% |
| HR | 863 | 9689 | 431 | 1591451098324 | 96.38% |
| HT | 87 | 624 | 134 | 26468807231 | 93.34% |
| HU | 2028 | 23379 | 1320 | 3217143006101 | 98.33% |
| ID | 6904 | 81221 | 5262 | 3840379842671 | 98.07% |
| IE | 8196 | 72214 | 8722 | 4986551128456 | 88.67% |
| IL | 2506 | 37115 | 1543 | 4297859388483 | 97.41% |
| IM | 349 | 476 | 192 | 361489643069 | 96.14% |
| IN | 13409 | 197406 | 16536 | 29004691746985 | 97.79% |
| IO | 27 | 14 | 85 | 4755682346 | 95.94% |
| IQ | 679 | 3347 | 980 | 3415584355239 | 96.90% |
| IR | 1989 | 42739 | 2187 | 17060955147119 | 98.24% |
| IS | 344 | 3642 | 495 | 1562904222586 | 97.32% |
| IT | 9976 | 227814 | 9369 | 47054042337687 | 99.07% |
| JE | 232 | 640 | 118 | 219604759853 | 95.45% |
| JM | 253 | 1131 | 155 | 1122142713433 | 99.96% |
| JO | 449 | 2703 | 256 | 1121936404829 | 97.90% |
| JP | 12644 | 782648 | 15041 | 53263147200667 | 88.23% |
| KE | 1156 | 25127 | 669 | 692347316408 | 99.12% |
| KG | 238 | 1186 | 222 | 529225201260 | 99.38% |
| KH | 554 | 2011 | 300 | 430692688487 | 81.58% |
| KI | 79 | 36 | 137 | 13520203849 | 97.88% |
| KM | 67 | 53 | 84 | 9337298982 | 96.78% |
| KN | 133 | 135 | 123 | 10060018997 | 69.74% |
| KP | 9 | 11 | 418 | 499122479 | 55.30% |
| KR | 4036 | 455041 | 3669 | 22712175716173 | 99.87% |
| KW | 439 | 7585 | 245 | 877742068643 | 99.72% |
| KY | 129 | 449 | 108 | 17861638449 | 99.08% |
| KZ | 1387 | 13990 | 1214 | 3150641221168 | 86.82% |
| LA | 146 | 414 | 150 | 121260925230 | 88.17% |
| LB | 389 | 2384 | 293 | 3015926620999 | 97.70% |
| LC | 98 | 168 | 93 | 78392390948 | 97.53% |
| LI | 291 | 452 | 201 | 429947487270 | 93.65% |
| LK | 255 | 2325 | 265 | 82523791734 | 94.32% |
| LR | 83 | 440 | 108 | 27275920682 | 97.70% |
| LS | 68 | 507 | 94 | 26802794536 | 98.25% |
| LT | 1444 | 11936 | 971 | 4177006927824 | 84.82% |
| LU | 1560 | 6708 | 877 | 2222696006513 | 96.44% |
| LV | 1254 | 8164 | 781 | 3235507122396 | 65.64% |
| LY | 152 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 343 | 48136 | 268 | 1171100465830 | 99.78% |
| MC | 279 | 281 | 107 | 142166715948 | 99.88% |
| MD | 747 | 4967 | 974 | 5953066431819 | 64.09% |
| ME | 148 | 823 | 130 | 348120040249 | 99.48% |
| MF | 56 | 105 | 90 | 5443269678 | 92.31% |
| MG | 87 | 2309 | 113 | 22644159027 | 98.09% |
| MH | 85 | 94 | 81 | 4806418475 | 95.22% |
| MK | 255 | 2801 | 186 | 1138468598636 | 97.20% |
| ML | 83 | 388 | 89 | 22675267629 | 97.52% |
| MM | 460 | 1170 | 336 | 353155015036 | 97.86% |
| MN | 236 | 737 | 285 | 133921418860 | 99.23% |
| MO | 105 | 1500 | 716 | 39116546363 | 99.15% |
| MP | 88 | 16 | 77 | 4805746732 | 95.20% |
| MQ | 324 | 818 | 72 | 104815670307 | 89.69% |
| MR | 76 | 234 | 97 | 26517696554 | 98.88% |
| MS | 31 | 15 | 79 | 644674862 | 51.26% |
| MT | 320 | 2676 | 226 | 508545920822 | 92.95% |
| MU | 1980 | 23848 | 8205 | 4455171008065850 | 100.00% |
| MV | 119 | 403 | 131 | 18180677684 | 97.48% |
| MW | 124 | 2201 | 145 | 1242315518000 | 99.85% |
| MX | 4429 | 122531 | 2575 | 3651392584597 | 97.03% |
| MY | 2864 | 34665 | 1627 | 911924741199 | 96.08% |
| MZ | 195 | 1914 | 132 | 69652197436 | 97.38% |
| NA | 341 | 1857 | 185 | 29169103149 | 99.37% |
| NC | 77 | 571 | 109 | 47910416692 | 99.48% |
| NE | 64 | 218 | 91 | 22122717220 | 98.72% |
| NF | 37 | 3 | 86 | 357560365 | 59.06% |
| NG | 1100 | 12921 | 708 | 604382767232 | 98.31% |
| NI | 170 | 1736 | 123 | 95171114308 | 98.22% |
| NL | 24907 | 214618 | 486434 | 8407427832342060 | 99.94% |
| NO | 3355 | 63975 | 2945 | 10986138734383 | 97.26% |
| NP | 469 | 2393 | 342 | 430212202595 | 94.05% |
| NR | 42 | 53 | 90 | 8951231533 | 98.41% |
| NU | 27 | 11 | 75 | 4671337512 | 81.46% |
| NZ | 2287 | 28871 | 1785 | 1355928779281 | 94.98% |
| OM | 317 | 4418 | 203 | 595072376919 | 96.04% |
| PA | 561 | 7099 | 487 | 427213831475 | 93.34% |
| PE | 1307 | 14738 | 1081 | 1072415464979 | 96.36% |
| PF | 92 | 312 | 146 | 21875061852 | 99.30% |
| PG | 234 | 349 | 152 | 73449136459 | 87.10% |
| PH | 2271 | 26817 | 1276 | 1050699737349 | 97.67% |
| PK | 1463 | 23636 | 740 | 907295961334 | 97.95% |
| PL | 8784 | 83155 | 5404 | 28613661523533 | 96.74% |
| PM | 42 | 25 | 80 | 4671532332 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.41% |
| PR | 662 | 4574 | 333 | 230687482760 | 99.07% |
| PS | 303 | 3312 | 180 | 979424837967 | 93.84% |
| PT | 1903 | 27062 | 2061 | 3362638427337 | 92.83% |
| PW | 48 | 43 | 85 | 9085575213 | 97.53% |
| PY | 448 | 5094 | 273 | 451733299854 | 97.52% |
| QA | 501 | 5197 | 368 | 526445535199 | 96.23% |
| RE | 348 | 2008 | 99 | 95622438961 | 96.43% |
| RO | 3987 | 28319 | 2167 | 5572234368993 | 95.57% |
| RS | 1061 | 9071 | 1342 | 1887339793845 | 97.41% |
| RU | 13823 | 180914 | 16097 | 54670074224631 | 90.13% |
| RW | 190 | 1266 | 124 | 1147559424300 | 99.91% |
| SA | 1692 | 48872 | 925 | 2815443836387 | 96.57% |
| SB | 55 | 71 | 100 | 17944015142 | 90.06% |
| SC | 928 | 2629 | 3263 | 3305345753640 | 61.91% |
| SD | 116 | 7412 | 110 | 69736980787 | 97.36% |
| SE | 16269 | 126528 | 56557 | 30393369355447 | 79.90% |
| SG | 10966 | 110808 | 52744 | 133667403735388 | 98.94% |
| SH | 6 | 2 | 130 | 575140868 | 54.83% |
| SI | 1061 | 10331 | 621 | 2079447594939 | 95.35% |
| SJ | 7 | 2 | 42 | 374469650 | 60.68% |
| SK | 1230 | 11152 | 944 | 3009684354324 | 96.94% |
| SL | 164 | 359 | 113 | 26584806447 | 96.08% |
| SM | 72 | 171 | 97 | 197980646949 | 93.41% |
| SN | 119 | 1646 | 151 | 27161199685 | 98.08% |
| SO | 76 | 225 | 162 | 48249762112 | 99.11% |
| SR | 51 | 340 | 106 | 17813725489 | 98.45% |
| SS | 105 | 135 | 109 | 31098290221 | 98.66% |
| ST | 49 | 60 | 92 | 4773835818 | 96.05% |
| SV | 339 | 2787 | 210 | 211465054056 | 97.09% |
| SX | 47 | 128 | 82 | 30441468965 | 94.79% |
| SY | 297 | 5171 | 137 | 876931813422 | 99.64% |
| SZ | 77 | 243 | 106 | 34670981174 | 96.15% |
| TC | 44 | 56 | 85 | 1451586856 | 71.10% |
| TD | 94 | 134 | 106 | 13816296493 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 72 | 1403 | 114 | 26550813990 | 98.82% |
| TH | 2290 | 43582 | 1272 | 595753052759 | 95.52% |
| TJ | 159 | 402 | 168 | 641204728379 | 98.14% |
| TK | 26 | 15 | 115 | 4789891119 | 95.34% |
| TL | 84 | 121 | 108 | 43596832813 | 92.49% |
| TM | 63 | 126 | 105 | 69114983736 | 99.77% |
| TN | 148 | 30817 | 150 | 116780509759 | 97.52% |
| TO | 48 | 55 | 84 | 17523466281 | 99.25% |
| TR | 3617 | 68052 | 2897 | 22547089083772 | 93.80% |
| TT | 167 | 2226 | 146 | 1177257588298 | 70.56% |
| TV | 41 | 46 | 98 | 11112280102 | 82.95% |
| TW | 9306 | 160420 | 3163 | 11549442161469 | 98.68% |
| TZ | 427 | 4498 | 248 | 267366409332 | 98.77% |
| UA | 4904 | 34629 | 3342 | 11756587902042 | 93.16% |
| UG | 266 | 5774 | 182 | 132375003452 | 96.35% |
| UM | 5 | 2 | 84 | 355005483 | 58.92% |
| US | 135872 | 5869425 | 541228 | 4835808606723808 | 99.76% |
| UY | 246 | 10118 | 193 | 177623867459 | 99.80% |
| UZ | 546 | 1661 | 280 | 1083134232176 | 99.45% |
| VA | 36 | 51 | 104 | 103325379883 | 99.88% |
| VC | 84 | 103 | 91 | 69500515369 | 99.67% |
| VE | 990 | 30432 | 658 | 2279266779574 | 99.37% |
| VG | 408 | 1299 | 266 | 255457202226 | 89.93% |
| VI | 100 | 548 | 171 | 30633514052 | 94.72% |
| VN | 2267 | 65519 | 1798 | 1860462138390 | 99.21% |
| VU | 66 | 86 | 276 | 51951564592 | 80.96% |
| WF | 22 | 16 | 71 | 4788515881 | 95.30% |
| WS | 62 | 97 | 89 | 21837242666 | 92.81% |
| XK | 223 | 623 | 73 | 111806775287 | 60.60% |
| YE | 115 | 958 | 118 | 207096833081 | 99.81% |
| YT | 73 | 42 | 75 | 602988584 | 56.00% |
| ZA | 5161 | 117572 | 2695 | 41040392008520 | 78.31% |
| ZM | 348 | 6515 | 160 | 48484905298 | 96.08% |
| ZW | 305 | 938 | 195 | 65582730599 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
