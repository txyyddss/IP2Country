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

Last run: `2026-11-20T05:47:04Z`

Input rows: **38904781** · Published blocks: **2378255**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 566164 | 0 | 0 | 0 | 11182502 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 495491 | 0 | 0 | 15135175 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 568077 | 0 | 0 | 0 | 11896529 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 495575 | 0 | 0 | 34194675 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 408820 | 0 | 0 | 0 | 8902738 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 487459 | 0 | 1 | 16081122 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 407628 | 0 | 0 | 0 | 6063381 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 484683 | 0 | 1 | 18858459 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 427728 | 0 | 0 | 0 | 9312812 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 581007 | 0 | 1 | 17346532 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1076888 | 0 | 0 | 0 | 13975780 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2226789 | 0 | 0 | 53593542 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1435654 | 2461496 | 0 | 0 | 187006261 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 805504 | 1407104 | 0 | 0 | 17708280 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177469 | 69917 | 0 | 0 | 3405091 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 295462 | 427677 | 0 | 97511 | 38070079 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 295462 | 427677 | 0 | 98011 | 17160887 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4334 | 1728 | 0 | 0 | 98689 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3906 | 0 | 0 | 0 | 83367 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327930 | 0 | 0 | 0 | 537841 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 816074 | 365864 | 0 | 412392 | 18087722 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262736 | 71183 | 0 | 10 | 34644102 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4957269 | 0 | 0 | 0 | 66550081 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380534 | 0 | 0 | 11707152 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6227 | 3446 | 0 | 0 | 154234 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1750 | 578 | 0 | 0 | 36094 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 568847 | 496310 | 0 | 0 | 3402189 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 620794 | 3702257408 | 1757211 | 293421847207842436812546030166054273024 |
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
| AE | 2839 | 25549 | 2783 | 12770686251734 | 90.03% |
| AF | 220 | 875 | 479 | 121136965474 | 97.42% |
| AG | 123 | 177 | 186 | 40393307195 | 65.26% |
| AI | 44 | 45 | 100 | 5847643190 | 67.40% |
| AL | 576 | 1413 | 1253 | 2508745105310 | 95.10% |
| AM | 360 | 2324 | 501 | 1297972697275 | 94.48% |
| AO | 532 | 5125 | 507 | 147599979092 | 98.68% |
| AQ | 56 | 5 | 293 | 405022851 | 44.23% |
| AR | 3808 | 77019 | 2391 | 23631006441574 | 98.86% |
| AS | 41 | 29 | 348 | 9243193557 | 96.04% |
| AT | 4748 | 45214 | 4243 | 10540281511356 | 97.36% |
| AU | 13466 | 219839 | 71406 | 2309962490832161723 | 100.00% |
| AW | 50 | 449 | 96 | 11534578740 | 92.96% |
| AX | 84 | 132 | 643 | 69552506698 | 72.57% |
| AZ | 399 | 3070 | 333 | 1963968488044 | 98.36% |
| BA | 323 | 3264 | 743 | 833827497501 | 95.13% |
| BB | 194 | 507 | 109 | 69400590027 | 99.69% |
| BD | 2650 | 8191 | 2595 | 3577593943777 | 96.48% |
| BE | 8572 | 58037 | 4852 | 4398062668687413 | 66.61% |
| BF | 126 | 1388 | 145 | 91245528373 | 99.43% |
| BG | 2364 | 16711 | 1434 | 5129179011887 | 96.12% |
| BH | 336 | 5248 | 317 | 406108709709 | 97.65% |
| BI | 59 | 168 | 110 | 26969694760 | 92.79% |
| BJ | 85 | 622 | 128 | 56582575407 | 96.84% |
| BL | 23 | 10 | 87 | 1047851055 | 57.46% |
| BM | 320 | 501 | 107 | 43426918710 | 99.37% |
| BN | 111 | 879 | 116 | 26250616876 | 99.06% |
| BO | 324 | 4753 | 306 | 168302404200 | 98.07% |
| BQ | 38 | 103 | 70 | 17538741295 | 98.90% |
| BR | 11541 | 335619 | 36071 | 37445172215552 | 99.20% |
| BS | 140 | 634 | 161 | 14003324500 | 96.66% |
| BT | 125 | 226 | 157 | 60963021390 | 99.36% |
| BV | 1 | 1 | 3 | 17825792 | 34.31% |
| BW | 188 | 720 | 122 | 78628794671 | 97.50% |
| BY | 402 | 7406 | 444 | 700163598654 | 99.49% |
| BZ | 212 | 470 | 202 | 192713406283 | 89.02% |
| CA | 22369 | 285230 | 28626 | 6373433454617 | 95.85% |
| CC | 7 | 0 | 51 | 827256853 | 42.84% |
| CD | 340 | 940 | 202 | 155963863378 | 98.74% |
| CF | 54 | 40 | 96 | 1117335591 | 59.87% |
| CG | 92 | 585 | 116 | 36988538926 | 97.17% |
| CH | 6937 | 98602 | 57327 | 15878800940092925389 | 99.98% |
| CI | 347 | 6714 | 169 | 194921354549 | 98.93% |
| CK | 35 | 41 | 92 | 9234282542 | 96.05% |
| CL | 2080 | 41385 | 2058 | 1419535897172 | 96.66% |
| CM | 189 | 2439 | 162 | 67900154955 | 98.18% |
| CN | 6983 | 1347884 | 13369 | 294810402912326 | 99.94% |
| CO | 3188 | 70296 | 1924 | 1990222209907 | 97.96% |
| CR | 607 | 9167 | 371 | 461702268525 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 82 | 242 | 145 | 13715370030 | 97.22% |
| CW | 139 | 718 | 103 | 63125206060 | 92.42% |
| CX | 8 | 0 | 81 | 727429155 | 41.43% |
| CY | 755 | 4141 | 711 | 6008425292642 | 94.02% |
| CZ | 3902 | 33829 | 6271 | 13303825497935 | 97.51% |
| DE | 39381 | 550759 | 82203 | 115554347761644 | 96.02% |
| DJ | 65 | 247 | 121 | 9641325614 | 94.96% |
| DK | 3675 | 50107 | 2097 | 9637217875876 | 93.45% |
| DM | 74 | 71 | 116 | 44401817134 | 73.08% |
| DO | 741 | 8097 | 630 | 932803543396 | 98.92% |
| DZ | 234 | 18724 | 242 | 44176666794 | 95.48% |
| EC | 1093 | 11267 | 660 | 1162867668869 | 98.70% |
| EE | 1195 | 6491 | 1436 | 2725769061661 | 96.02% |
| EG | 801 | 94821 | 281 | 17726611054997 | 100.00% |
| EH | 4 | 0 | 54 | 457961495 | 46.78% |
| ER | 46 | 29 | 94 | 966149163 | 59.37% |
| ES | 11291 | 139918 | 6403 | 26202036688538 | 97.12% |
| ET | 96 | 1483 | 134 | 1117547839542 | 99.96% |
| FI | 4148 | 58807 | 33535 | 6895831386364 | 96.70% |
| FJ | 114 | 590 | 118 | 43379783467 | 99.48% |
| FK | 39 | 26 | 85 | 612233254 | 46.30% |
| FM | 57 | 49 | 107 | 8921740345 | 78.74% |
| FO | 46 | 181 | 88 | 108187546923 | 99.57% |
| FR | 28001 | 325640 | 18976 | 67546411684002 | 98.46% |
| GA | 90 | 1558 | 106 | 35779724581 | 94.29% |
| GB | 41022 | 498570 | 61110 | 105888629820356 | 93.88% |
| GD | 101 | 128 | 104 | 82705774122 | 96.88% |
| GE | 490 | 5120 | 364 | 1831825930367 | 95.38% |
| GF | 261 | 254 | 80 | 23010009128 | 97.39% |
| GG | 147 | 264 | 96 | 112129799470 | 81.66% |
| GH | 359 | 8925 | 266 | 1232412982882 | 99.85% |
| GI | 157 | 329 | 107 | 197924942890 | 90.89% |
| GL | 103 | 144 | 165 | 43621736567 | 99.27% |
| GM | 75 | 1103 | 110 | 39586922793 | 98.99% |
| GN | 77 | 168 | 117 | 48293807144 | 97.24% |
| GP | 210 | 802 | 105 | 54408922167 | 73.07% |
| GQ | 52 | 94 | 94 | 9454961706 | 96.02% |
| GR | 1534 | 23125 | 885 | 2592595808407 | 96.51% |
| GS | 4 | 2 | 40 | 406904853 | 59.18% |
| GT | 671 | 2899 | 281 | 327769491059 | 99.41% |
| GU | 182 | 995 | 128 | 34757665084 | 99.56% |
| GW | 53 | 42 | 95 | 9320611883 | 96.97% |
| GY | 58 | 288 | 110 | 35023941676 | 96.09% |
| HK | 13935 | 95194 | 8253 | 5095704599164 | 85.93% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 843 | 1994 | 266 | 391924627056 | 94.76% |
| HR | 857 | 9689 | 435 | 1591451098328 | 96.38% |
| HT | 87 | 624 | 134 | 26468807231 | 93.34% |
| HU | 2022 | 23379 | 1539 | 3217143104843 | 98.33% |
| ID | 6866 | 81184 | 5298 | 3844675006591 | 98.05% |
| IE | 8250 | 72212 | 8775 | 4986551049012 | 88.66% |
| IL | 2506 | 37117 | 1553 | 4297859322956 | 97.41% |
| IM | 355 | 476 | 191 | 361489643068 | 96.14% |
| IN | 13387 | 197411 | 14141 | 29021871636060 | 97.77% |
| IO | 28 | 14 | 83 | 4755682344 | 95.94% |
| IQ | 679 | 3346 | 845 | 3415584355102 | 96.90% |
| IR | 1985 | 42739 | 2200 | 17060954983276 | 98.20% |
| IS | 339 | 3642 | 486 | 1562904222577 | 97.32% |
| IT | 10182 | 227816 | 8606 | 47055385555467 | 99.07% |
| JE | 233 | 640 | 117 | 219604759852 | 95.45% |
| JM | 253 | 1131 | 155 | 1122142713433 | 99.96% |
| JO | 447 | 2703 | 256 | 1121936404829 | 97.90% |
| JP | 12675 | 782626 | 14080 | 53263147183184 | 88.23% |
| KE | 1153 | 25127 | 768 | 694493751415 | 98.92% |
| KG | 240 | 1186 | 212 | 529225201250 | 99.38% |
| KH | 554 | 2011 | 299 | 430675911271 | 81.58% |
| KI | 79 | 36 | 137 | 13520203849 | 97.88% |
| KM | 67 | 53 | 82 | 9337298980 | 96.78% |
| KN | 105 | 102 | 122 | 12207453493 | 64.51% |
| KP | 9 | 11 | 421 | 499401003 | 55.27% |
| KR | 4035 | 455025 | 3541 | 22712175714948 | 99.87% |
| KW | 439 | 7585 | 223 | 877742068164 | 99.72% |
| KY | 128 | 447 | 108 | 17861638449 | 99.08% |
| KZ | 1390 | 13991 | 1248 | 3150641221202 | 86.77% |
| LA | 148 | 450 | 150 | 121260925230 | 88.17% |
| LB | 391 | 2384 | 286 | 3015926620992 | 97.70% |
| LC | 102 | 168 | 93 | 78392390948 | 97.53% |
| LI | 290 | 451 | 208 | 429951419430 | 93.65% |
| LK | 252 | 2325 | 261 | 82523791730 | 94.32% |
| LR | 82 | 439 | 107 | 27275920681 | 97.70% |
| LS | 68 | 507 | 93 | 26802794535 | 98.25% |
| LT | 1446 | 11947 | 970 | 4177006862102 | 84.82% |
| LU | 1575 | 6713 | 877 | 2222694958210 | 96.44% |
| LV | 1248 | 8163 | 790 | 3235507122405 | 65.64% |
| LY | 152 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 345 | 48138 | 261 | 1171100465823 | 99.78% |
| MC | 279 | 281 | 107 | 142166715948 | 99.88% |
| MD | 746 | 4969 | 978 | 5953066431823 | 64.00% |
| ME | 148 | 823 | 130 | 348120040249 | 99.48% |
| MF | 56 | 105 | 89 | 5443269677 | 92.30% |
| MG | 87 | 2309 | 113 | 22644159027 | 98.09% |
| MH | 85 | 94 | 80 | 4806418474 | 95.22% |
| MK | 255 | 2800 | 185 | 1138468647766 | 97.20% |
| ML | 83 | 388 | 88 | 22675267628 | 97.52% |
| MM | 459 | 1170 | 336 | 353155015036 | 97.86% |
| MN | 236 | 737 | 275 | 133921418850 | 99.23% |
| MO | 105 | 1500 | 752 | 39116611935 | 99.15% |
| MP | 88 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 320 | 827 | 72 | 104815670307 | 89.69% |
| MR | 76 | 234 | 97 | 26517696554 | 98.88% |
| MS | 32 | 15 | 79 | 644674607 | 51.26% |
| MT | 320 | 2676 | 231 | 509619662650 | 92.82% |
| MU | 1960 | 23850 | 8213 | 4455168861630846 | 100.00% |
| MV | 119 | 403 | 131 | 18180677684 | 97.48% |
| MW | 124 | 2201 | 145 | 1242315518000 | 99.85% |
| MX | 4463 | 122531 | 2554 | 3651392584725 | 96.99% |
| MY | 2872 | 34677 | 1637 | 911924807256 | 96.08% |
| MZ | 195 | 1914 | 132 | 69652197436 | 97.38% |
| NA | 305 | 1858 | 185 | 29169103149 | 98.29% |
| NC | 77 | 571 | 108 | 47910416691 | 99.48% |
| NE | 64 | 218 | 91 | 22122717220 | 98.72% |
| NF | 37 | 3 | 85 | 357560364 | 59.06% |
| NG | 1100 | 12921 | 708 | 604382767232 | 98.31% |
| NI | 170 | 1736 | 123 | 95171114308 | 98.22% |
| NL | 24706 | 214617 | 485971 | 8407283937282518 | 99.94% |
| NO | 3369 | 63969 | 2430 | 10985932147271 | 97.26% |
| NP | 469 | 2393 | 341 | 430212202594 | 94.05% |
| NR | 42 | 53 | 90 | 8951231533 | 98.41% |
| NU | 27 | 11 | 74 | 4671337511 | 81.46% |
| NZ | 2278 | 28860 | 1777 | 1360223697410 | 94.94% |
| OM | 321 | 4418 | 204 | 595072376920 | 96.04% |
| PA | 560 | 7098 | 463 | 427213619507 | 93.34% |
| PE | 1298 | 14742 | 1071 | 1076710432260 | 96.28% |
| PF | 92 | 312 | 144 | 21875061850 | 99.30% |
| PG | 233 | 349 | 152 | 73449136459 | 87.10% |
| PH | 2271 | 26817 | 1280 | 1050699737353 | 97.67% |
| PK | 1466 | 23635 | 742 | 907295961318 | 97.95% |
| PL | 8881 | 83152 | 5230 | 28613661354363 | 96.71% |
| PM | 42 | 25 | 79 | 4671532331 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.40% |
| PR | 666 | 4575 | 338 | 230687499658 | 99.07% |
| PS | 301 | 3306 | 176 | 979424837963 | 93.84% |
| PT | 1933 | 27062 | 1850 | 3362638389205 | 92.83% |
| PW | 48 | 43 | 84 | 9085575212 | 97.53% |
| PY | 448 | 5092 | 273 | 451733299854 | 97.52% |
| QA | 517 | 5197 | 360 | 526445535178 | 96.23% |
| RE | 358 | 2034 | 95 | 95622373422 | 96.43% |
| RO | 4043 | 28320 | 2169 | 5572234401767 | 95.57% |
| RS | 1057 | 9069 | 1318 | 1887339792337 | 97.41% |
| RU | 13875 | 180904 | 16442 | 54717318701286 | 90.05% |
| RW | 173 | 1266 | 122 | 1147559424298 | 99.91% |
| SA | 1691 | 48872 | 929 | 2815443836391 | 96.57% |
| SB | 55 | 71 | 100 | 17944015142 | 90.06% |
| SC | 818 | 2533 | 3264 | 3288165884457 | 62.04% |
| SD | 116 | 7412 | 109 | 69736980786 | 97.36% |
| SE | 17617 | 126542 | 54999 | 30393369285004 | 79.94% |
| SG | 10890 | 110847 | 52737 | 133701763213159 | 98.92% |
| SH | 6 | 2 | 130 | 575140868 | 54.83% |
| SI | 1037 | 10331 | 630 | 2079455786938 | 95.35% |
| SJ | 7 | 2 | 42 | 374469650 | 60.67% |
| SK | 1227 | 11153 | 949 | 3009701131543 | 96.94% |
| SL | 164 | 359 | 112 | 26584806446 | 96.08% |
| SM | 72 | 172 | 97 | 197980646949 | 93.41% |
| SN | 119 | 1646 | 150 | 27161199684 | 98.08% |
| SO | 76 | 225 | 162 | 48249762112 | 99.11% |
| SR | 51 | 340 | 106 | 17813725489 | 98.45% |
| SS | 105 | 135 | 108 | 31098290220 | 98.66% |
| ST | 49 | 60 | 92 | 4773835818 | 96.05% |
| SV | 339 | 2787 | 209 | 211465054055 | 97.09% |
| SX | 64 | 128 | 81 | 30441468964 | 94.79% |
| SY | 297 | 5171 | 136 | 876931813421 | 99.64% |
| SZ | 77 | 243 | 104 | 34670981172 | 96.15% |
| TC | 44 | 56 | 84 | 1451586855 | 71.10% |
| TD | 95 | 134 | 105 | 13816296492 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 72 | 1403 | 114 | 26550813990 | 98.82% |
| TH | 2284 | 43544 | 1259 | 595753102899 | 95.52% |
| TJ | 160 | 402 | 166 | 641204728377 | 98.14% |
| TK | 26 | 15 | 114 | 4789891118 | 95.34% |
| TL | 84 | 121 | 107 | 43596832812 | 92.49% |
| TM | 63 | 126 | 104 | 69114983735 | 99.77% |
| TN | 148 | 30817 | 146 | 116780509755 | 97.52% |
| TO | 48 | 55 | 83 | 17523466280 | 99.25% |
| TR | 3702 | 68053 | 2902 | 22529908985238 | 93.84% |
| TT | 167 | 2226 | 146 | 1177257588298 | 70.56% |
| TV | 41 | 46 | 98 | 11112280102 | 82.95% |
| TW | 9292 | 160416 | 3197 | 11549442360174 | 98.68% |
| TZ | 429 | 4498 | 248 | 267366409332 | 98.77% |
| UA | 4944 | 34595 | 3306 | 11756586983913 | 93.16% |
| UG | 266 | 5774 | 180 | 132375003450 | 96.35% |
| UM | 5 | 2 | 85 | 355005484 | 58.92% |
| US | 137851 | 5870349 | 544738 | 4835778497642762 | 99.76% |
| UY | 246 | 10118 | 191 | 177623867457 | 99.80% |
| UZ | 548 | 1661 | 272 | 1083134183018 | 99.45% |
| VA | 36 | 51 | 102 | 103325379881 | 99.88% |
| VC | 81 | 103 | 90 | 69500515368 | 99.67% |
| VE | 993 | 30442 | 652 | 2279266779568 | 99.37% |
| VG | 380 | 301 | 260 | 255456170034 | 89.93% |
| VI | 101 | 548 | 171 | 30615360583 | 94.73% |
| VN | 2264 | 65517 | 1809 | 1860462187552 | 99.21% |
| VU | 66 | 86 | 274 | 51951564590 | 80.96% |
| WF | 22 | 16 | 71 | 4788515881 | 95.30% |
| WS | 62 | 97 | 89 | 21837242666 | 92.81% |
| XK | 216 | 623 | 72 | 111806775286 | 60.60% |
| YE | 115 | 958 | 118 | 207096833081 | 99.81% |
| YT | 81 | 54 | 74 | 602988583 | 56.00% |
| ZA | 5153 | 117569 | 2604 | 41040391988334 | 78.31% |
| ZM | 336 | 6514 | 160 | 48484905298 | 96.08% |
| ZW | 275 | 934 | 306 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
