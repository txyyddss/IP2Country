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

Last run: `2026-11-23T05:48:35Z`

Input rows: **38879244** · Published blocks: **2376746**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 569628 | 0 | 0 | 0 | 11251768 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 496255 | 0 | 0 | 15185935 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 571548 | 0 | 0 | 0 | 11969494 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 496335 | 0 | 0 | 34247115 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 409570 | 0 | 0 | 0 | 8927870 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 488066 | 0 | 1 | 16102162 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 409493 | 0 | 0 | 0 | 6094362 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 487635 | 0 | 1 | 18978243 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 428478 | 0 | 0 | 0 | 9337944 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 581614 | 0 | 1 | 17367572 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1090447 | 0 | 0 | 0 | 14010804 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2250105 | 0 | 0 | 53560284 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1449201 | 2485348 | 0 | 0 | 188832889 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 805879 | 1408667 | 0 | 0 | 17723784 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177493 | 69942 | 0 | 0 | 3406189 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 302856 | 322344 | 0 | 9 | 29165424 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 302856 | 322344 | 0 | 9 | 13359298 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4341 | 1719 | 0 | 0 | 98632 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3901 | 0 | 0 | 0 | 83187 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327930 | 0 | 0 | 0 | 537841 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 819103 | 436464 | 0 | 412664 | 18639841 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262780 | 71208 | 0 | 10 | 34651309 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4957269 | 0 | 0 | 0 | 66550081 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380534 | 0 | 0 | 11707152 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 803015 | 0 | 0 | 0 | 11123810 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6224 | 3447 | 0 | 0 | 154220 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1750 | 578 | 0 | 0 | 36094 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 572321 | 497070 | 0 | 0 | 3411761 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 618892 | 3702257408 | 1757604 | 293421847207842436812546030166054273024 |
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
| AD | 217 | 308 | 193 | 43728565323 | 92.57% |
| AE | 2844 | 25495 | 2793 | 12770686710498 | 90.03% |
| AF | 224 | 875 | 478 | 121136965473 | 97.42% |
| AG | 135 | 177 | 186 | 40393307195 | 65.26% |
| AI | 45 | 45 | 100 | 5847643190 | 67.40% |
| AL | 577 | 1412 | 1254 | 2508745105311 | 95.10% |
| AM | 364 | 2324 | 497 | 1293644175546 | 94.63% |
| AO | 530 | 5124 | 507 | 147599979092 | 98.68% |
| AQ | 60 | 5 | 291 | 405022849 | 44.23% |
| AR | 3859 | 77017 | 2382 | 23630993858663 | 98.86% |
| AS | 42 | 29 | 348 | 9243193557 | 96.04% |
| AT | 4751 | 45206 | 4244 | 10540281511357 | 97.36% |
| AU | 13603 | 219872 | 71435 | 2309962473652358059 | 100.00% |
| AW | 51 | 449 | 96 | 11534578740 | 92.96% |
| AX | 131 | 132 | 643 | 69552506698 | 72.57% |
| AZ | 403 | 3070 | 333 | 1963968488044 | 98.36% |
| BA | 326 | 3264 | 743 | 833827497501 | 95.13% |
| BB | 195 | 507 | 109 | 69400590027 | 99.69% |
| BD | 2653 | 8195 | 2596 | 3581888911073 | 96.45% |
| BE | 8542 | 58038 | 4860 | 4398097011648573 | 66.61% |
| BF | 126 | 1387 | 145 | 91245528373 | 99.43% |
| BG | 2354 | 16717 | 1436 | 5129179011889 | 96.12% |
| BH | 340 | 5248 | 317 | 406108709709 | 97.65% |
| BI | 60 | 168 | 109 | 26952917544 | 92.82% |
| BJ | 86 | 622 | 128 | 56582575407 | 96.84% |
| BL | 24 | 10 | 87 | 1047851055 | 57.46% |
| BM | 317 | 501 | 107 | 43426918710 | 99.37% |
| BN | 113 | 879 | 118 | 26250616878 | 99.06% |
| BO | 327 | 4753 | 306 | 168302404200 | 98.07% |
| BQ | 39 | 103 | 71 | 21833708591 | 91.24% |
| BR | 11787 | 335636 | 36072 | 37440877117185 | 99.21% |
| BS | 141 | 634 | 161 | 14003324500 | 96.66% |
| BT | 126 | 226 | 157 | 60963021390 | 99.36% |
| BV | 1 | 1 | 3 | 17825792 | 34.31% |
| BW | 189 | 723 | 123 | 82923761967 | 96.34% |
| BY | 397 | 7406 | 444 | 700163598654 | 99.49% |
| BZ | 214 | 469 | 201 | 192579188555 | 89.06% |
| CA | 20001 | 284786 | 28500 | 6442137202714 | 95.46% |
| CC | 7 | 0 | 51 | 827256853 | 42.84% |
| CD | 348 | 940 | 202 | 155963863378 | 98.74% |
| CF | 55 | 40 | 96 | 1117335591 | 59.87% |
| CG | 93 | 585 | 116 | 36988538926 | 97.17% |
| CH | 6990 | 98601 | 57334 | 15878800912996803020 | 99.98% |
| CI | 356 | 6712 | 167 | 194887800117 | 98.94% |
| CK | 36 | 41 | 92 | 9234282542 | 96.05% |
| CL | 2091 | 41393 | 2058 | 1419535897172 | 96.66% |
| CM | 190 | 2438 | 162 | 67900154955 | 98.18% |
| CN | 6958 | 1347894 | 13390 | 294844762585155 | 99.94% |
| CO | 3192 | 70302 | 1926 | 1994517111667 | 97.92% |
| CR | 593 | 9170 | 363 | 461706462829 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 82 | 241 | 145 | 13715370030 | 97.22% |
| CW | 152 | 718 | 385 | 67420173339 | 90.78% |
| CX | 7 | 0 | 81 | 727429155 | 41.43% |
| CY | 673 | 4130 | 712 | 6042768253795 | 93.92% |
| CZ | 3952 | 33837 | 6270 | 13303791943504 | 97.50% |
| DE | 38701 | 550754 | 81938 | 115519852982953 | 96.03% |
| DJ | 66 | 247 | 121 | 9641325614 | 94.96% |
| DK | 3665 | 50104 | 2097 | 9637217892261 | 93.39% |
| DM | 77 | 71 | 116 | 44401817134 | 73.08% |
| DO | 746 | 8104 | 631 | 937098510692 | 98.77% |
| DZ | 239 | 18724 | 244 | 47397892266 | 92.39% |
| EC | 1085 | 11270 | 659 | 1162834114437 | 98.71% |
| EE | 1203 | 6488 | 1431 | 2725760804126 | 96.02% |
| EG | 812 | 94821 | 279 | 17726577500565 | 100.00% |
| EH | 4 | 0 | 54 | 457961495 | 46.78% |
| ER | 47 | 29 | 95 | 1100366891 | 57.01% |
| ES | 11471 | 139915 | 6405 | 26202036688538 | 97.12% |
| ET | 97 | 1483 | 134 | 1117547839542 | 99.96% |
| FI | 4161 | 58807 | 33611 | 6895814625336 | 96.70% |
| FJ | 115 | 590 | 118 | 43379783467 | 99.48% |
| FK | 40 | 26 | 83 | 578678822 | 46.09% |
| FM | 59 | 49 | 107 | 8921740345 | 78.74% |
| FO | 47 | 181 | 88 | 108187546923 | 99.57% |
| FR | 28298 | 325721 | 18873 | 67555001633953 | 98.45% |
| GA | 91 | 1558 | 106 | 35779724581 | 94.29% |
| GB | 41596 | 498638 | 61713 | 105918776526792 | 93.87% |
| GD | 102 | 128 | 104 | 82705774122 | 96.88% |
| GE | 494 | 5120 | 364 | 1831825930367 | 95.38% |
| GF | 279 | 253 | 79 | 23010009127 | 97.39% |
| GG | 139 | 263 | 96 | 112129799470 | 81.66% |
| GH | 361 | 8925 | 264 | 1232362655330 | 99.85% |
| GI | 157 | 329 | 107 | 197924942890 | 90.89% |
| GL | 104 | 144 | 165 | 43621736567 | 99.27% |
| GM | 75 | 1102 | 110 | 39586922793 | 98.99% |
| GN | 78 | 168 | 117 | 48293807144 | 97.24% |
| GP | 223 | 810 | 105 | 54408922167 | 73.07% |
| GQ | 53 | 94 | 94 | 9454961706 | 96.02% |
| GR | 1553 | 23125 | 885 | 2592595808407 | 96.51% |
| GS | 4 | 2 | 40 | 406904853 | 59.18% |
| GT | 676 | 2899 | 281 | 327769491059 | 99.41% |
| GU | 188 | 995 | 191 | 34757665085 | 99.56% |
| GW | 54 | 42 | 95 | 9320611883 | 96.97% |
| GY | 58 | 290 | 110 | 35023941676 | 96.09% |
| HK | 14015 | 95099 | 8190 | 5099982658171 | 85.81% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 847 | 1995 | 267 | 396219594352 | 94.38% |
| HR | 880 | 9689 | 435 | 1591451098328 | 96.38% |
| HT | 88 | 624 | 134 | 26468807231 | 93.34% |
| HU | 2086 | 23380 | 1568 | 3217143235911 | 98.33% |
| ID | 6915 | 81213 | 5304 | 3827495465088 | 98.11% |
| IE | 8330 | 72212 | 8775 | 4986551114547 | 88.66% |
| IL | 2504 | 37086 | 1567 | 4297791689803 | 97.42% |
| IM | 368 | 476 | 191 | 361489643068 | 96.14% |
| IN | 13496 | 197405 | 14147 | 28991807061602 | 97.81% |
| IO | 29 | 14 | 83 | 4755682344 | 95.94% |
| IQ | 694 | 3348 | 845 | 3415584355102 | 96.90% |
| IR | 1980 | 42742 | 2200 | 17060954983276 | 98.16% |
| IS | 342 | 3642 | 485 | 1562904206193 | 97.32% |
| IT | 10389 | 227815 | 8607 | 47055385555981 | 99.06% |
| JE | 231 | 641 | 117 | 219604759852 | 95.45% |
| JM | 255 | 1131 | 155 | 1122142713433 | 99.96% |
| JO | 452 | 2702 | 256 | 1121936404829 | 97.90% |
| JP | 12872 | 782559 | 14055 | 53263147379795 | 88.22% |
| KE | 1177 | 25131 | 773 | 698755164342 | 98.77% |
| KG | 241 | 1186 | 212 | 529225201250 | 99.38% |
| KH | 555 | 2011 | 299 | 430675911271 | 81.58% |
| KI | 80 | 36 | 137 | 13520203849 | 97.88% |
| KM | 68 | 52 | 81 | 9320521764 | 96.86% |
| KN | 117 | 103 | 123 | 14354937141 | 66.82% |
| KP | 9 | 11 | 421 | 499401003 | 55.27% |
| KR | 4087 | 455043 | 3542 | 22712175780484 | 99.87% |
| KW | 442 | 7581 | 222 | 877742051780 | 99.72% |
| KY | 127 | 447 | 108 | 17861638449 | 99.08% |
| KZ | 1382 | 13991 | 1249 | 3150641221203 | 86.75% |
| LA | 147 | 414 | 150 | 121260925230 | 88.17% |
| LB | 393 | 2384 | 285 | 3016027284288 | 97.70% |
| LC | 103 | 168 | 93 | 78392390948 | 97.53% |
| LI | 285 | 451 | 209 | 429951435814 | 93.12% |
| LK | 260 | 2325 | 261 | 82523791730 | 94.32% |
| LR | 85 | 440 | 107 | 27275920681 | 97.70% |
| LS | 69 | 507 | 93 | 26802794535 | 98.25% |
| LT | 1442 | 11952 | 970 | 4177006845719 | 84.82% |
| LU | 1578 | 6712 | 878 | 2222695023746 | 96.44% |
| LV | 1252 | 8166 | 789 | 3235473567974 | 65.64% |
| LY | 153 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 352 | 48138 | 261 | 1171100465823 | 99.78% |
| MC | 280 | 281 | 107 | 142166715948 | 99.88% |
| MD | 753 | 4974 | 977 | 5953032877392 | 64.00% |
| ME | 150 | 823 | 130 | 348120040249 | 99.48% |
| MF | 57 | 105 | 90 | 5443335213 | 92.30% |
| MG | 87 | 2308 | 112 | 22627381811 | 98.13% |
| MH | 87 | 94 | 80 | 4806418474 | 95.22% |
| MK | 257 | 2800 | 186 | 1138602865494 | 97.19% |
| ML | 84 | 388 | 87 | 22658490412 | 97.55% |
| MM | 462 | 1170 | 336 | 353155015036 | 97.86% |
| MN | 239 | 737 | 275 | 133921418850 | 99.23% |
| MO | 105 | 1500 | 752 | 39116611935 | 99.15% |
| MP | 82 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 309 | 807 | 72 | 104815670307 | 89.69% |
| MR | 77 | 234 | 97 | 26517696554 | 98.88% |
| MS | 33 | 15 | 79 | 644674607 | 51.26% |
| MT | 323 | 2676 | 232 | 508545920828 | 92.95% |
| MU | 1963 | 23839 | 8221 | 4455160271696191 | 100.00% |
| MV | 120 | 403 | 131 | 18180677684 | 97.48% |
| MW | 125 | 2200 | 145 | 1242315518000 | 99.85% |
| MX | 4509 | 122537 | 2554 | 3651392584726 | 96.99% |
| MY | 2818 | 34665 | 1637 | 911924741721 | 96.08% |
| MZ | 196 | 1914 | 132 | 69652197436 | 97.38% |
| NA | 309 | 1860 | 184 | 29152325933 | 98.31% |
| NC | 79 | 571 | 108 | 47910416691 | 99.48% |
| NE | 64 | 218 | 90 | 22105940004 | 98.76% |
| NF | 37 | 3 | 85 | 357560364 | 59.06% |
| NG | 1120 | 12922 | 710 | 604382767234 | 98.31% |
| NI | 172 | 1736 | 123 | 95171114308 | 98.22% |
| NL | 23846 | 214568 | 485826 | 8407039694358638 | 99.94% |
| NO | 3419 | 63982 | 2430 | 10985932147271 | 97.26% |
| NP | 471 | 2393 | 341 | 430212202594 | 94.05% |
| NR | 43 | 53 | 90 | 8951231533 | 98.41% |
| NU | 28 | 11 | 73 | 4654560295 | 81.58% |
| NZ | 2297 | 28862 | 1793 | 1360223762945 | 94.94% |
| OM | 323 | 4418 | 203 | 595055599704 | 96.04% |
| PA | 557 | 7096 | 463 | 427213619507 | 93.34% |
| PE | 1288 | 14776 | 1072 | 1081034759684 | 96.11% |
| PF | 94 | 312 | 144 | 21875061850 | 99.30% |
| PG | 235 | 349 | 152 | 73449136459 | 87.10% |
| PH | 2290 | 26818 | 1281 | 1050699737354 | 97.67% |
| PK | 1481 | 23641 | 743 | 907296026854 | 97.95% |
| PL | 8927 | 83149 | 5234 | 28613661485438 | 96.71% |
| PM | 43 | 25 | 79 | 4671532331 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.40% |
| PR | 667 | 4575 | 337 | 230687466890 | 99.07% |
| PS | 301 | 3298 | 176 | 979424837963 | 93.84% |
| PT | 1962 | 27062 | 1846 | 3362638323669 | 92.83% |
| PW | 49 | 43 | 84 | 9085575212 | 97.53% |
| PY | 454 | 5092 | 275 | 451733299856 | 97.52% |
| QA | 547 | 5197 | 360 | 560788496330 | 94.93% |
| RE | 337 | 1996 | 95 | 95622373422 | 96.43% |
| RO | 4102 | 28302 | 2155 | 5572233353193 | 95.57% |
| RS | 1068 | 9069 | 1316 | 1887339792339 | 97.41% |
| RU | 13888 | 180965 | 16476 | 54751678374163 | 90.02% |
| RW | 174 | 1266 | 122 | 1147559424298 | 99.91% |
| SA | 1702 | 48872 | 929 | 2815443836391 | 96.57% |
| SB | 56 | 71 | 100 | 17944015142 | 90.06% |
| SC | 839 | 2562 | 3264 | 3288165884457 | 62.04% |
| SD | 118 | 7412 | 109 | 69736980786 | 97.36% |
| SE | 17745 | 126532 | 54808 | 30393335731335 | 79.94% |
| SG | 10767 | 110843 | 52788 | 133701762951034 | 98.92% |
| SH | 6 | 2 | 130 | 575140868 | 54.83% |
| SI | 1045 | 10331 | 630 | 2079455786938 | 95.35% |
| SJ | 7 | 2 | 42 | 374469650 | 60.67% |
| SK | 1232 | 11154 | 949 | 3009701131543 | 96.94% |
| SL | 165 | 359 | 112 | 26584806446 | 96.08% |
| SM | 73 | 172 | 97 | 197980646949 | 93.41% |
| SN | 120 | 1646 | 150 | 27161199684 | 98.08% |
| SO | 77 | 225 | 160 | 48216207680 | 99.14% |
| SR | 53 | 340 | 104 | 17780171057 | 98.54% |
| SS | 106 | 135 | 108 | 31098290220 | 98.66% |
| ST | 50 | 60 | 92 | 4773835818 | 96.05% |
| SV | 336 | 2785 | 209 | 211465054055 | 97.09% |
| SX | 49 | 144 | 82 | 30575686692 | 94.55% |
| SY | 297 | 5171 | 138 | 945651290157 | 97.85% |
| SZ | 78 | 243 | 104 | 34670981172 | 96.15% |
| TC | 45 | 56 | 84 | 1451586855 | 71.10% |
| TD | 95 | 133 | 105 | 13816296492 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 73 | 1403 | 112 | 26517259558 | 98.88% |
| TH | 2283 | 43573 | 1259 | 595753037379 | 95.52% |
| TJ | 162 | 403 | 166 | 641204728377 | 98.14% |
| TK | 27 | 15 | 114 | 4789891118 | 95.34% |
| TL | 85 | 121 | 107 | 43596832812 | 92.49% |
| TM | 64 | 126 | 102 | 69081429303 | 99.79% |
| TN | 151 | 30817 | 144 | 116746955323 | 97.54% |
| TO | 48 | 55 | 83 | 17523466280 | 99.25% |
| TR | 3723 | 68054 | 2904 | 22529892273560 | 93.74% |
| TT | 167 | 2226 | 146 | 1177257588298 | 70.56% |
| TV | 42 | 46 | 98 | 11112280102 | 82.95% |
| TW | 9312 | 160423 | 3195 | 11549442359917 | 98.68% |
| TZ | 428 | 4495 | 247 | 267349632116 | 98.78% |
| UA | 4934 | 34577 | 3307 | 11756586983914 | 93.14% |
| UG | 266 | 5773 | 180 | 132375003450 | 96.35% |
| UM | 5 | 2 | 85 | 355005484 | 58.92% |
| US | 136747 | 5870762 | 544822 | 4835771223656473 | 99.76% |
| UY | 247 | 10119 | 192 | 177623932993 | 99.80% |
| UZ | 553 | 1661 | 272 | 1083134183018 | 99.45% |
| VA | 36 | 51 | 102 | 103325379881 | 99.88% |
| VC | 82 | 103 | 90 | 69500515368 | 99.67% |
| VE | 987 | 30445 | 652 | 2279266779568 | 99.37% |
| VG | 357 | 311 | 260 | 255456170034 | 89.93% |
| VI | 103 | 548 | 171 | 30615360583 | 94.73% |
| VN | 2288 | 65517 | 1810 | 1860462253088 | 99.17% |
| VU | 67 | 86 | 274 | 51951564590 | 80.96% |
| WF | 23 | 16 | 71 | 4788515881 | 95.30% |
| WS | 63 | 97 | 89 | 21837242666 | 92.81% |
| XK | 222 | 624 | 72 | 111806775286 | 60.60% |
| YE | 118 | 958 | 118 | 207096833081 | 99.81% |
| YT | 73 | 42 | 74 | 602988583 | 56.00% |
| ZA | 5189 | 117587 | 2603 | 41040391988334 | 78.31% |
| ZM | 382 | 6515 | 160 | 48484905298 | 96.08% |
| ZW | 271 | 938 | 306 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
