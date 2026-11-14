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

Last run: `2026-11-14T04:59:18Z`

Input rows: **40348864** · Published blocks: **2411289**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562899 | 0 | 0 | 0 | 11115565 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 494493 | 0 | 0 | 15052423 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564812 | 0 | 0 | 0 | 11828088 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 494577 | 0 | 0 | 34125813 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 400931 | 0 | 0 | 0 | 8722013 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 477299 | 0 | 1 | 15640547 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 403682 | 0 | 0 | 0 | 6015129 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 476731 | 0 | 1 | 18481995 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 419772 | 0 | 0 | 0 | 9131509 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 570847 | 0 | 1 | 16905957 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1036887 | 0 | 0 | 0 | 13678705 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2951044 | 0 | 0 | 96186690 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1395917 | 3183860 | 0 | 0 | 225452022 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 389960 | 0 | 0 | 0 | 6095165 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1556723 | 0 | 0 | 24472437 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 800896 | 1397574 | 0 | 0 | 17595176 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177315 | 69864 | 0 | 0 | 3402211 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 293586 | 489746 | 0 | 97819 | 41028264 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 293586 | 489746 | 0 | 98325 | 18542930 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4337 | 1722 | 0 | 0 | 98623 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3884 | 0 | 0 | 0 | 82599 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 814393 | 427056 | 0 | 412045 | 18496076 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262639 | 71130 | 0 | 10 | 34628512 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4958198 | 0 | 0 | 0 | 66567473 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380280 | 0 | 0 | 11696034 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6211 | 3428 | 0 | 0 | 153634 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 578 | 0 | 0 | 36064 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 565582 | 495307 | 0 | 0 | 3394049 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 613566 | 3702257408 | 1797473 | 293421847207842436812546030166054273024 |
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
| AD | 217 | 306 | 190 | 43728563274 | 92.57% |
| AE | 2840 | 25528 | 2578 | 12835044389585 | 89.85% |
| AF | 219 | 875 | 525 | 121136965521 | 97.42% |
| AG | 120 | 178 | 183 | 40393304123 | 65.26% |
| AI | 44 | 45 | 102 | 5847641146 | 67.40% |
| AL | 576 | 1416 | 1243 | 2500156808089 | 95.22% |
| AM | 350 | 2322 | 519 | 1263612958926 | 95.01% |
| AO | 567 | 5141 | 297 | 147599966561 | 98.68% |
| AQ | 56 | 5 | 285 | 472046733 | 43.65% |
| AR | 3798 | 77014 | 2830 | 23626711425248 | 98.82% |
| AS | 41 | 29 | 309 | 9243191472 | 96.04% |
| AT | 4750 | 45207 | 4439 | 10471298213492 | 97.53% |
| AU | 13322 | 220195 | 71495 | 2309962503213119562 | 100.00% |
| AW | 50 | 449 | 98 | 11534575673 | 92.96% |
| AX | 86 | 132 | 626 | 69552487243 | 72.57% |
| AZ | 402 | 3072 | 340 | 1963968520834 | 98.36% |
| BA | 334 | 3266 | 742 | 833827495454 | 95.13% |
| BB | 194 | 507 | 109 | 69400586958 | 99.69% |
| BD | 2658 | 8181 | 2713 | 3577593881840 | 96.48% |
| BE | 8508 | 57826 | 4846 | 4398028328492681 | 66.61% |
| BF | 125 | 1387 | 144 | 91245525303 | 99.43% |
| BG | 2363 | 16713 | 1435 | 5137769027393 | 96.07% |
| BH | 334 | 5246 | 319 | 406108709711 | 97.65% |
| BI | 59 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 86 | 623 | 127 | 56582572337 | 96.84% |
| BL | 24 | 10 | 86 | 1047849008 | 57.46% |
| BM | 328 | 498 | 108 | 43426867513 | 99.37% |
| BN | 111 | 880 | 119 | 26250616879 | 99.06% |
| BO | 327 | 4752 | 307 | 168302404201 | 98.07% |
| BQ | 38 | 103 | 69 | 17538738225 | 98.90% |
| BR | 11475 | 336087 | 42501 | 37419671490253 | 99.23% |
| BS | 141 | 634 | 164 | 14004566614 | 96.65% |
| BT | 125 | 226 | 160 | 60963021393 | 99.36% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 192 | 719 | 123 | 78628794672 | 97.43% |
| BY | 390 | 7401 | 426 | 700163287346 | 99.49% |
| BZ | 220 | 470 | 197 | 192579185482 | 89.06% |
| CA | 22137 | 285146 | 34350 | 6355582118666 | 95.77% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 345 | 940 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 119 | 36988538929 | 97.17% |
| CH | 6997 | 98176 | 57454 | 15878800956213568856 | 99.98% |
| CI | 346 | 6712 | 170 | 194921285946 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2069 | 41383 | 2387 | 1419536887486 | 96.56% |
| CM | 189 | 2439 | 160 | 67900151884 | 98.18% |
| CN | 6923 | 1347611 | 12716 | 294809843479996 | 99.94% |
| CO | 3176 | 70287 | 1947 | 1981632272271 | 98.03% |
| CR | 603 | 9166 | 367 | 461701216877 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 82 | 241 | 144 | 13715367983 | 97.22% |
| CW | 139 | 718 | 101 | 63125202989 | 92.43% |
| CX | 8 | 0 | 82 | 727443495 | 41.43% |
| CY | 712 | 4169 | 707 | 5999835487071 | 94.03% |
| CZ | 3876 | 33844 | 6447 | 13338200868511 | 97.34% |
| DE | 38179 | 550914 | 84549 | 115476572992762 | 96.04% |
| DJ | 68 | 246 | 119 | 9641322543 | 94.96% |
| DK | 3614 | 50103 | 2080 | 9667282682796 | 93.40% |
| DM | 74 | 71 | 116 | 44401879600 | 73.08% |
| DO | 763 | 8082 | 626 | 932803540323 | 98.88% |
| DZ | 233 | 18723 | 256 | 44176730297 | 95.48% |
| EC | 1090 | 11267 | 657 | 1158571649927 | 98.76% |
| EE | 1206 | 6513 | 1984 | 2725759609291 | 96.05% |
| EG | 794 | 94817 | 371 | 17726611071391 | 100.00% |
| EH | 4 | 0 | 55 | 457959450 | 46.78% |
| ER | 46 | 29 | 91 | 966146091 | 59.37% |
| ES | 11172 | 139979 | 7002 | 26203112510206 | 97.15% |
| ET | 94 | 1478 | 133 | 1117547837495 | 99.96% |
| FI | 4061 | 58687 | 33678 | 6895833453179 | 96.68% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 38 | 25 | 84 | 612231207 | 46.31% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 86 | 108187543852 | 99.57% |
| FR | 27554 | 325700 | 21152 | 67482227727965 | 98.47% |
| GA | 91 | 1559 | 107 | 35779721513 | 94.29% |
| GB | 40291 | 498466 | 61243 | 105755804181633 | 93.92% |
| GD | 98 | 128 | 102 | 82705771051 | 96.88% |
| GE | 488 | 5116 | 357 | 1831825930361 | 95.38% |
| GF | 262 | 255 | 80 | 23010009128 | 97.39% |
| GG | 145 | 263 | 94 | 112129796399 | 81.66% |
| GH | 365 | 8924 | 264 | 1232412718691 | 99.85% |
| GI | 149 | 330 | 106 | 197924939820 | 90.89% |
| GL | 103 | 144 | 169 | 43621736571 | 99.27% |
| GM | 75 | 1103 | 109 | 39586920746 | 98.99% |
| GN | 76 | 167 | 116 | 48293804074 | 97.24% |
| GP | 225 | 804 | 102 | 54408919095 | 73.07% |
| GQ | 53 | 95 | 92 | 9454958635 | 96.02% |
| GR | 1537 | 23120 | 999 | 2592595276558 | 96.51% |
| GS | 4 | 2 | 41 | 406904854 | 59.19% |
| GT | 673 | 2894 | 276 | 327768439410 | 99.41% |
| GU | 190 | 995 | 129 | 34757730620 | 99.56% |
| GW | 53 | 42 | 94 | 9320609836 | 96.97% |
| GY | 57 | 288 | 108 | 35023939628 | 96.09% |
| HK | 13835 | 95184 | 8039 | 5095572367175 | 85.87% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 845 | 1995 | 266 | 391924755057 | 94.76% |
| HR | 835 | 9687 | 431 | 1591451110613 | 96.38% |
| HT | 83 | 624 | 133 | 26468805184 | 93.34% |
| HU | 2011 | 23373 | 1301 | 3217143031235 | 98.33% |
| ID | 6924 | 81284 | 5267 | 3840379596908 | 98.07% |
| IE | 8336 | 72247 | 8910 | 4956488435128 | 88.88% |
| IL | 2504 | 37146 | 1565 | 4293581347438 | 97.47% |
| IM | 348 | 477 | 190 | 361489639998 | 96.14% |
| IN | 13394 | 197369 | 16792 | 28970330779824 | 97.84% |
| IO | 27 | 14 | 84 | 4755679276 | 95.94% |
| IQ | 678 | 3347 | 980 | 3415584355240 | 96.90% |
| IR | 1971 | 42734 | 2212 | 16962237844363 | 98.49% |
| IS | 337 | 3638 | 493 | 1597264102267 | 96.30% |
| IT | 9934 | 227752 | 9589 | 47054042051718 | 99.08% |
| JE | 242 | 640 | 116 | 219604756782 | 95.45% |
| JM | 251 | 1131 | 152 | 1122142710361 | 99.96% |
| JO | 446 | 2703 | 254 | 1121936402781 | 97.90% |
| JP | 12556 | 782511 | 14864 | 53262073080987 | 88.23% |
| KE | 1080 | 25121 | 662 | 688052414647 | 99.34% |
| KG | 236 | 1184 | 220 | 529225198189 | 99.38% |
| KH | 559 | 2015 | 300 | 430692688487 | 81.58% |
| KI | 79 | 36 | 136 | 13520201802 | 97.88% |
| KM | 66 | 52 | 85 | 9337298983 | 96.78% |
| KN | 100 | 133 | 103 | 5765065014 | 76.99% |
| KP | 9 | 11 | 420 | 499188016 | 55.29% |
| KR | 4038 | 455108 | 3641 | 22712174393931 | 99.87% |
| KW | 432 | 7581 | 245 | 877741020068 | 99.72% |
| KY | 127 | 450 | 106 | 17861635378 | 99.08% |
| KZ | 1400 | 13978 | 1211 | 3150641399341 | 86.82% |
| LA | 148 | 450 | 151 | 121260925231 | 88.17% |
| LB | 391 | 2387 | 293 | 3015926620999 | 98.08% |
| LC | 96 | 168 | 91 | 78392387877 | 97.53% |
| LI | 336 | 450 | 200 | 425652452394 | 93.84% |
| LK | 253 | 2326 | 264 | 82523808116 | 94.32% |
| LR | 82 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1449 | 11967 | 964 | 4177275226000 | 84.79% |
| LU | 1553 | 6704 | 883 | 2222680422257 | 96.44% |
| LV | 1250 | 8161 | 774 | 3235775225052 | 65.64% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 340 | 48132 | 266 | 1171100463782 | 99.78% |
| MC | 285 | 285 | 107 | 142166714925 | 99.88% |
| MD | 760 | 4975 | 972 | 5953066429771 | 64.00% |
| ME | 153 | 828 | 129 | 348120038202 | 99.48% |
| MF | 56 | 103 | 88 | 5443266607 | 92.31% |
| MG | 85 | 2307 | 114 | 22644159028 | 98.09% |
| MH | 84 | 94 | 81 | 4806418475 | 95.24% |
| MK | 253 | 2800 | 186 | 1142897781613 | 97.05% |
| ML | 83 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 236 | 737 | 286 | 133921418861 | 99.23% |
| MO | 104 | 1500 | 709 | 39116108346 | 99.15% |
| MP | 94 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 325 | 822 | 70 | 104815667236 | 89.69% |
| MR | 75 | 233 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 77 | 644671791 | 51.26% |
| MT | 320 | 2676 | 227 | 508545919800 | 92.95% |
| MU | 1981 | 23863 | 8192 | 4455188187935035 | 100.00% |
| MV | 120 | 403 | 132 | 18180677685 | 97.48% |
| MW | 123 | 2200 | 146 | 1242315518001 | 99.85% |
| MX | 4420 | 122459 | 2517 | 3642801429923 | 97.12% |
| MY | 2836 | 34650 | 1622 | 911924741194 | 96.08% |
| MZ | 186 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 336 | 1857 | 186 | 29169103150 | 99.37% |
| NC | 77 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1096 | 12922 | 710 | 604382963839 | 98.31% |
| NI | 168 | 1731 | 119 | 95171111235 | 98.22% |
| NL | 24892 | 214455 | 496762 | 8407869892888658 | 99.94% |
| NO | 3339 | 63983 | 2972 | 10986138580787 | 97.26% |
| NP | 467 | 2390 | 343 | 430195425381 | 94.05% |
| NR | 41 | 52 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2309 | 28869 | 1783 | 1355927646481 | 94.92% |
| OM | 316 | 4416 | 202 | 595072376918 | 96.04% |
| PA | 541 | 7093 | 469 | 429360248116 | 92.96% |
| PE | 1254 | 14661 | 1077 | 1072415464975 | 96.36% |
| PF | 93 | 312 | 143 | 21875058780 | 99.30% |
| PG | 232 | 347 | 153 | 73449136460 | 87.10% |
| PH | 2263 | 26803 | 1272 | 1050699670018 | 97.59% |
| PK | 1452 | 23636 | 740 | 907296026869 | 97.95% |
| PL | 8740 | 83094 | 5377 | 28613658986573 | 96.76% |
| PM | 42 | 25 | 80 | 4671531309 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 653 | 4572 | 335 | 230687480713 | 99.07% |
| PS | 303 | 3312 | 178 | 979424835919 | 93.84% |
| PT | 1904 | 27060 | 1979 | 3328261717208 | 93.01% |
| PW | 47 | 43 | 87 | 9085575215 | 97.53% |
| PY | 435 | 5090 | 274 | 451733299855 | 97.52% |
| QA | 499 | 5196 | 367 | 526445600733 | 96.23% |
| RE | 333 | 1981 | 100 | 95622438962 | 96.43% |
| RO | 3972 | 28288 | 2516 | 5572250970599 | 95.57% |
| RS | 1070 | 9078 | 1344 | 1887339857336 | 97.41% |
| RU | 13726 | 180713 | 15969 | 54635182677843 | 90.14% |
| RW | 189 | 1265 | 125 | 1147559424301 | 99.91% |
| SA | 1671 | 48717 | 920 | 2815427188195 | 96.56% |
| SB | 55 | 71 | 99 | 17944013095 | 90.06% |
| SC | 945 | 2645 | 3187 | 3305379439146 | 62.02% |
| SD | 115 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16234 | 126499 | 61293 | 30393371127166 | 79.91% |
| SG | 10995 | 111037 | 52760 | 133667671875968 | 98.93% |
| SH | 6 | 2 | 128 | 575137797 | 54.83% |
| SI | 1063 | 10333 | 614 | 2079447644091 | 95.52% |
| SJ | 7 | 2 | 40 | 374466579 | 60.68% |
| SK | 1235 | 11162 | 933 | 3009684262165 | 96.94% |
| SL | 157 | 358 | 111 | 26584803376 | 96.08% |
| SM | 73 | 173 | 97 | 197980645926 | 93.41% |
| SN | 118 | 1645 | 149 | 27161196614 | 98.08% |
| SO | 77 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 341 | 107 | 17813725490 | 98.45% |
| SS | 102 | 135 | 110 | 31098290222 | 98.66% |
| ST | 49 | 60 | 90 | 4773832747 | 96.05% |
| SV | 339 | 2780 | 208 | 211465050985 | 97.09% |
| SX | 48 | 128 | 81 | 30575683622 | 94.55% |
| SY | 297 | 5172 | 138 | 876931813423 | 99.64% |
| SZ | 76 | 242 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 83 | 1451583785 | 71.10% |
| TD | 94 | 133 | 104 | 13816293422 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2294 | 43585 | 1269 | 595753118291 | 95.52% |
| TJ | 158 | 402 | 167 | 641204726332 | 98.14% |
| TK | 25 | 15 | 114 | 4789889072 | 95.34% |
| TL | 87 | 125 | 109 | 43596832814 | 92.49% |
| TM | 63 | 126 | 102 | 69114915129 | 99.77% |
| TN | 148 | 30817 | 148 | 116780507711 | 97.52% |
| TO | 49 | 56 | 85 | 17523466282 | 99.25% |
| TR | 3600 | 68043 | 2908 | 22512863294328 | 93.80% |
| TT | 167 | 2226 | 144 | 1177257585227 | 70.56% |
| TV | 40 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9287 | 160447 | 3110 | 11548367958830 | 98.68% |
| TZ | 396 | 4498 | 249 | 267366409333 | 98.77% |
| UA | 4899 | 34641 | 3346 | 11756586919000 | 93.17% |
| UG | 265 | 5773 | 183 | 132375003453 | 96.35% |
| UM | 5 | 2 | 81 | 355002411 | 58.92% |
| US | 135863 | 5870418 | 544468 | 4835846553572249 | 99.76% |
| UY | 247 | 10117 | 193 | 177623867459 | 99.80% |
| UZ | 548 | 1661 | 277 | 1083134229104 | 99.45% |
| VA | 36 | 51 | 106 | 103325444397 | 99.88% |
| VC | 84 | 103 | 89 | 69500512298 | 99.67% |
| VE | 998 | 30364 | 658 | 2279265730999 | 99.37% |
| VG | 351 | 765 | 404 | 255473977403 | 89.93% |
| VI | 90 | 547 | 154 | 30616602693 | 94.74% |
| VN | 2263 | 65531 | 1784 | 1860462113803 | 99.17% |
| VU | 67 | 87 | 274 | 51951561521 | 80.96% |
| WF | 22 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 189 | 623 | 72 | 111806775286 | 60.60% |
| YE | 115 | 958 | 119 | 207096833082 | 99.81% |
| YT | 82 | 54 | 75 | 602988584 | 56.00% |
| ZA | 5199 | 117662 | 2695 | 41031802270534 | 78.31% |
| ZM | 365 | 6515 | 161 | 48484905299 | 96.08% |
| ZW | 291 | 938 | 196 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
