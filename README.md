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

Last run: `2026-11-06T05:08:29Z`

Input rows: **39670327** · Published blocks: **2408436**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 561853 | 0 | 0 | 0 | 11091375 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 495045 | 0 | 0 | 15042161 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 563803 | 0 | 0 | 0 | 11806955 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 495128 | 0 | 0 | 34163832 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 403391 | 0 | 0 | 0 | 8800891 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 474570 | 0 | 1 | 15606604 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 402990 | 0 | 0 | 0 | 6005752 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 474092 | 0 | 1 | 18430521 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 420562 | 0 | 0 | 0 | 9182723 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 566478 | 0 | 1 | 16821381 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1038275 | 0 | 0 | 0 | 13678339 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2923573 | 0 | 0 | 93803844 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1397364 | 3155986 | 0 | 0 | 223911427 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 389960 | 0 | 0 | 0 | 6095165 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1556723 | 0 | 0 | 24472437 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 796771 | 1392242 | 0 | 0 | 17519520 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177204 | 69823 | 0 | 0 | 3399937 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 294016 | 494290 | 0 | 98126 | 41306676 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 294016 | 494290 | 0 | 98646 | 18684334 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4346 | 1673 | 0 | 0 | 97983 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3884 | 0 | 0 | 0 | 82521 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 814057 | 434534 | 0 | 412085 | 18679671 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262513 | 71089 | 0 | 10 | 34611022 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802828 | 0 | 0 | 0 | 11122714 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6204 | 3434 | 0 | 0 | 153617 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 564571 | 495852 | 0 | 0 | 3389216 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 613557 | 3702257408 | 1794629 | 293421847207842436812546030166054273024 |
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
| AD | 214 | 308 | 183 | 43728628810 | 92.57% |
| AE | 2847 | 25531 | 2569 | 12835044553433 | 89.85% |
| AF | 219 | 875 | 524 | 121136965519 | 97.42% |
| AG | 106 | 167 | 184 | 40393304124 | 65.26% |
| AI | 44 | 45 | 101 | 5847641145 | 67.40% |
| AL | 574 | 1415 | 1244 | 2500693679258 | 95.22% |
| AM | 347 | 2322 | 526 | 1263612958933 | 95.01% |
| AO | 603 | 5142 | 298 | 147599966562 | 98.68% |
| AQ | 56 | 5 | 314 | 472063119 | 43.66% |
| AR | 3765 | 76979 | 2821 | 23617874026730 | 98.83% |
| AS | 39 | 29 | 311 | 9243191474 | 96.04% |
| AT | 4757 | 45180 | 4492 | 10505668879735 | 97.43% |
| AU | 13328 | 220216 | 71575 | 2309962513361658971 | 100.00% |
| AW | 49 | 449 | 98 | 11534575673 | 92.96% |
| AX | 92 | 132 | 603 | 69552372553 | 72.57% |
| AZ | 399 | 3071 | 347 | 1964102738559 | 98.31% |
| BA | 337 | 3265 | 742 | 833827495454 | 95.13% |
| BB | 193 | 507 | 109 | 69400586958 | 99.69% |
| BD | 2660 | 8189 | 2713 | 3577593750770 | 96.48% |
| BE | 8498 | 57897 | 4827 | 4398024048073339 | 66.61% |
| BF | 124 | 1383 | 144 | 91245525303 | 99.43% |
| BG | 2426 | 16723 | 1434 | 5137769043777 | 96.12% |
| BH | 330 | 5246 | 320 | 406108644177 | 97.65% |
| BI | 57 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 85 | 622 | 126 | 56582572336 | 96.84% |
| BL | 39 | 10 | 85 | 1047717936 | 57.46% |
| BM | 316 | 496 | 107 | 43426851129 | 99.37% |
| BN | 110 | 879 | 119 | 26250616879 | 99.06% |
| BO | 326 | 4752 | 307 | 168302404201 | 98.07% |
| BQ | 36 | 103 | 69 | 17538738225 | 98.90% |
| BR | 11436 | 335445 | 42583 | 37419670171730 | 99.22% |
| BS | 138 | 632 | 163 | 14003518038 | 96.66% |
| BT | 123 | 226 | 164 | 60978750033 | 99.35% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 189 | 718 | 123 | 78628794672 | 97.43% |
| BY | 390 | 7400 | 429 | 700163287349 | 99.59% |
| BZ | 217 | 469 | 198 | 192713403210 | 89.02% |
| CA | 22096 | 285518 | 31547 | 6351226660514 | 95.83% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 320 | 940 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 118 | 36988538928 | 97.17% |
| CH | 6839 | 98182 | 57390 | 15878800948516633387 | 99.98% |
| CI | 348 | 6714 | 169 | 194921285945 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2074 | 41383 | 2375 | 1415240953030 | 96.61% |
| CM | 188 | 2438 | 160 | 67900151884 | 98.18% |
| CN | 6928 | 1347604 | 12724 | 294808205112842 | 99.94% |
| CO | 3170 | 70315 | 1946 | 1981632206736 | 98.03% |
| CR | 598 | 9164 | 369 | 461701217134 | 97.36% |
| CU | 96 | 1023 | 129 | 26439696697 | 99.17% |
| CV | 82 | 241 | 143 | 13715433519 | 97.22% |
| CW | 139 | 718 | 101 | 63125202989 | 92.43% |
| CX | 8 | 0 | 83 | 727509031 | 41.43% |
| CY | 712 | 4163 | 706 | 5999831227235 | 94.03% |
| CZ | 3876 | 33838 | 6525 | 13303848469666 | 97.55% |
| DE | 38274 | 550725 | 84543 | 115451474519070 | 96.05% |
| DJ | 73 | 246 | 119 | 9641322543 | 94.96% |
| DK | 3666 | 50130 | 2081 | 9667282682796 | 93.43% |
| DM | 74 | 71 | 116 | 44401879600 | 73.08% |
| DO | 735 | 8072 | 625 | 928508573027 | 98.99% |
| DZ | 232 | 18723 | 257 | 44176730298 | 95.48% |
| EC | 1087 | 11258 | 658 | 1158571649928 | 98.76% |
| EE | 1193 | 6503 | 1965 | 2725759163293 | 96.02% |
| EG | 794 | 94816 | 292 | 17730906038686 | 99.99% |
| EH | 4 | 0 | 54 | 457959449 | 46.78% |
| ER | 44 | 29 | 91 | 966146091 | 59.37% |
| ES | 11268 | 139923 | 7003 | 26203112050160 | 97.15% |
| ET | 92 | 1478 | 133 | 1117547837495 | 99.96% |
| FI | 4010 | 58674 | 33490 | 6865768731105 | 96.76% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 38 | 25 | 84 | 612231207 | 46.31% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 86 | 108187543852 | 99.57% |
| FR | 27355 | 325623 | 21225 | 67379180313114 | 98.52% |
| GA | 91 | 1559 | 106 | 35779721512 | 94.29% |
| GB | 40291 | 498349 | 60601 | 105798454811317 | 94.12% |
| GD | 98 | 128 | 103 | 82705836587 | 96.88% |
| GE | 486 | 5114 | 355 | 1831824816249 | 95.38% |
| GF | 248 | 252 | 80 | 23010009128 | 97.39% |
| GG | 129 | 263 | 94 | 112129796399 | 81.66% |
| GH | 373 | 8926 | 264 | 1232412718691 | 99.85% |
| GI | 142 | 326 | 106 | 197924939820 | 90.89% |
| GL | 103 | 144 | 170 | 43621736572 | 99.27% |
| GM | 76 | 1104 | 109 | 39586855210 | 98.99% |
| GN | 74 | 167 | 116 | 48293804074 | 97.24% |
| GP | 226 | 800 | 102 | 54408919095 | 73.06% |
| GQ | 53 | 95 | 92 | 9454958635 | 96.02% |
| GR | 1530 | 23124 | 997 | 2592092025610 | 96.52% |
| GS | 4 | 2 | 41 | 406904854 | 59.19% |
| GT | 661 | 2886 | 276 | 327768439410 | 99.41% |
| GU | 181 | 995 | 130 | 34891948348 | 99.33% |
| GW | 51 | 42 | 94 | 9320609836 | 96.97% |
| GY | 57 | 288 | 108 | 35023939628 | 96.09% |
| HK | 13767 | 95191 | 8153 | 5086950416140 | 86.02% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 833 | 1992 | 266 | 391924755057 | 94.76% |
| HR | 841 | 9688 | 437 | 1591451061520 | 96.38% |
| HT | 83 | 624 | 133 | 26468805184 | 93.34% |
| HU | 1989 | 23374 | 1318 | 3217142834629 | 98.33% |
| ID | 6873 | 81194 | 5288 | 3843600216175 | 98.03% |
| IE | 8270 | 72345 | 8838 | 4921857128581 | 88.98% |
| IL | 2465 | 37169 | 1389 | 4297867840651 | 97.42% |
| IM | 329 | 475 | 190 | 361489639998 | 96.14% |
| IN | 13349 | 197343 | 17021 | 29000396387947 | 97.80% |
| IO | 25 | 14 | 83 | 4755679275 | 95.94% |
| IQ | 672 | 3347 | 963 | 3415584355222 | 96.90% |
| IR | 1977 | 42727 | 2209 | 16962237631371 | 98.49% |
| IS | 339 | 3639 | 492 | 1596995666811 | 96.34% |
| IT | 9931 | 227760 | 9577 | 47046789569616 | 99.09% |
| JE | 207 | 641 | 116 | 219604756782 | 95.45% |
| JM | 251 | 1131 | 152 | 1122142710361 | 99.96% |
| JO | 438 | 2702 | 253 | 1121936402780 | 97.90% |
| JP | 12535 | 782384 | 14909 | 53257923914932 | 88.23% |
| KE | 1152 | 25122 | 666 | 688052349112 | 99.34% |
| KG | 236 | 1184 | 218 | 529225198187 | 99.38% |
| KH | 558 | 2014 | 301 | 430692688488 | 81.58% |
| KI | 79 | 36 | 137 | 13520201803 | 97.75% |
| KM | 66 | 52 | 85 | 9337298983 | 96.78% |
| KN | 99 | 133 | 103 | 5765065014 | 76.99% |
| KP | 9 | 11 | 422 | 500187443 | 55.29% |
| KR | 4035 | 455045 | 3739 | 22712188282218 | 99.76% |
| KW | 425 | 7581 | 218 | 877741035968 | 99.72% |
| KY | 124 | 439 | 106 | 17861635378 | 99.08% |
| KZ | 1389 | 13968 | 1214 | 3150641399344 | 86.82% |
| LA | 145 | 443 | 151 | 121260925231 | 88.17% |
| LB | 392 | 2387 | 293 | 3015926620999 | 98.08% |
| LC | 103 | 168 | 90 | 78392322341 | 97.53% |
| LI | 298 | 452 | 201 | 429947419690 | 93.40% |
| LK | 245 | 2325 | 265 | 82540585332 | 94.31% |
| LR | 83 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1449 | 12010 | 965 | 4211635078928 | 84.64% |
| LU | 1544 | 6703 | 879 | 2222679225977 | 96.04% |
| LV | 1245 | 8158 | 770 | 3235776011488 | 65.66% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 340 | 48132 | 265 | 1171100463781 | 99.78% |
| MC | 285 | 284 | 108 | 142166649391 | 99.88% |
| MD | 760 | 4975 | 974 | 5953066380627 | 63.99% |
| ME | 154 | 828 | 129 | 348120038202 | 99.48% |
| MF | 56 | 103 | 88 | 5443266607 | 92.31% |
| MG | 85 | 2307 | 114 | 22644159028 | 98.09% |
| MH | 84 | 94 | 81 | 4806418475 | 95.24% |
| MK | 254 | 2800 | 186 | 1142897781613 | 97.13% |
| ML | 81 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1171 | 337 | 353155015037 | 97.86% |
| MN | 236 | 738 | 286 | 133921418861 | 99.23% |
| MO | 106 | 1500 | 711 | 39116173883 | 99.15% |
| MP | 91 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 336 | 838 | 70 | 104815667236 | 89.69% |
| MR | 73 | 233 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 77 | 644671791 | 51.26% |
| MT | 325 | 2681 | 230 | 509630147384 | 92.89% |
| MU | 2090 | 23867 | 8213 | 4455188187869499 | 100.00% |
| MV | 120 | 403 | 135 | 18180595767 | 97.48% |
| MW | 123 | 2200 | 146 | 1242315518001 | 99.85% |
| MX | 4389 | 122513 | 2601 | 3642801370785 | 97.00% |
| MY | 2830 | 34573 | 1625 | 911924937801 | 96.07% |
| MZ | 197 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 416 | 1855 | 186 | 29169103150 | 99.37% |
| NC | 76 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1119 | 12923 | 872 | 604382931070 | 98.31% |
| NI | 168 | 1731 | 120 | 95171111236 | 98.22% |
| NL | 25001 | 214358 | 496997 | 8408185829166343 | 99.94% |
| NO | 3287 | 63893 | 3050 | 10986135287583 | 97.26% |
| NP | 469 | 2390 | 343 | 430195425381 | 94.05% |
| NR | 41 | 52 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2405 | 28872 | 1787 | 1359417217362 | 94.87% |
| OM | 308 | 4416 | 202 | 595072376918 | 96.04% |
| PA | 534 | 7088 | 469 | 429360248116 | 92.79% |
| PE | 1256 | 14658 | 1078 | 1072415448593 | 96.36% |
| PF | 93 | 312 | 144 | 21875058781 | 99.30% |
| PG | 231 | 347 | 153 | 73449136460 | 87.10% |
| PH | 2244 | 26841 | 1271 | 1042109604357 | 97.81% |
| PK | 1446 | 23643 | 744 | 907296026873 | 97.95% |
| PL | 8726 | 83062 | 5357 | 28613474446796 | 96.79% |
| PM | 42 | 25 | 80 | 4671531309 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 641 | 4572 | 346 | 230687497864 | 99.07% |
| PS | 301 | 3304 | 178 | 979424835919 | 93.84% |
| PT | 1909 | 27057 | 2076 | 3328261766358 | 93.01% |
| PW | 47 | 43 | 87 | 9085575215 | 97.53% |
| PY | 437 | 5086 | 270 | 451733561976 | 97.52% |
| QA | 492 | 5196 | 366 | 526445600732 | 96.23% |
| RE | 362 | 2022 | 99 | 95622438961 | 96.44% |
| RO | 3916 | 28287 | 2489 | 5569029810665 | 95.61% |
| RS | 1072 | 9075 | 1349 | 1887339660726 | 97.41% |
| RU | 13758 | 180670 | 15961 | 54566193864535 | 90.19% |
| RW | 689 | 1264 | 125 | 1147559424301 | 99.91% |
| SA | 1713 | 48716 | 921 | 2815427171817 | 96.58% |
| SB | 55 | 71 | 87 | 17675643175 | 90.67% |
| SC | 893 | 2555 | 3201 | 3331434078765 | 61.99% |
| SD | 115 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16288 | 126465 | 61345 | 30392834030288 | 79.90% |
| SG | 11038 | 111056 | 52574 | 133650228257138 | 98.94% |
| SH | 6 | 2 | 130 | 574482436 | 54.36% |
| SI | 1064 | 10334 | 622 | 2079455901624 | 95.76% |
| SJ | 7 | 2 | 40 | 374466579 | 60.68% |
| SK | 1217 | 11163 | 931 | 3009684262163 | 97.04% |
| SL | 125 | 358 | 111 | 26584803376 | 96.08% |
| SM | 73 | 173 | 98 | 202275613222 | 92.49% |
| SN | 119 | 1645 | 149 | 27161196614 | 98.08% |
| SO | 75 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 341 | 107 | 17813725490 | 98.45% |
| SS | 99 | 135 | 110 | 31098290222 | 98.66% |
| ST | 47 | 60 | 90 | 4773832747 | 96.05% |
| SV | 333 | 2776 | 208 | 211465050985 | 97.09% |
| SX | 48 | 128 | 80 | 30441400359 | 94.79% |
| SY | 296 | 5171 | 138 | 876931813423 | 99.64% |
| SZ | 76 | 242 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 83 | 1451583785 | 71.10% |
| TD | 93 | 133 | 104 | 13816293422 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2292 | 43539 | 1268 | 595719498068 | 95.52% |
| TJ | 157 | 401 | 167 | 641204726332 | 98.14% |
| TK | 25 | 15 | 114 | 4789889072 | 95.34% |
| TL | 93 | 126 | 109 | 43596832814 | 92.49% |
| TM | 63 | 126 | 102 | 69114915129 | 99.77% |
| TN | 148 | 30817 | 148 | 116780507711 | 97.52% |
| TO | 49 | 56 | 85 | 17523466282 | 99.25% |
| TR | 3684 | 68036 | 2884 | 22512846402937 | 93.84% |
| TT | 167 | 2226 | 144 | 1177257585227 | 70.56% |
| TV | 40 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9252 | 160470 | 3116 | 11552662412338 | 98.66% |
| TZ | 452 | 4500 | 248 | 267366409332 | 98.77% |
| UA | 4887 | 34679 | 3314 | 11756586720906 | 93.13% |
| UG | 269 | 5773 | 182 | 132375003453 | 96.35% |
| UM | 5 | 2 | 84 | 355002414 | 58.92% |
| US | 135726 | 5872243 | 543783 | 4835710069105097 | 99.76% |
| UY | 245 | 10117 | 192 | 177623867458 | 99.80% |
| UZ | 549 | 1661 | 277 | 1083134229104 | 99.45% |
| VA | 36 | 51 | 107 | 103325509933 | 99.88% |
| VC | 84 | 103 | 89 | 69500512298 | 99.67% |
| VE | 966 | 30266 | 658 | 2279265730999 | 99.33% |
| VG | 349 | 271 | 420 | 251178960947 | 90.61% |
| VI | 88 | 547 | 159 | 30643734597 | 94.69% |
| VN | 2257 | 65524 | 1665 | 1860194014226 | 99.18% |
| VU | 65 | 86 | 274 | 51951561521 | 80.96% |
| WF | 21 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 182 | 622 | 76 | 111807037430 | 60.60% |
| YE | 113 | 958 | 119 | 207096833082 | 99.81% |
| YT | 46 | 48 | 76 | 602988585 | 56.00% |
| ZA | 5212 | 117606 | 3380 | 41031802180932 | 78.30% |
| ZM | 393 | 6516 | 161 | 48484905299 | 96.08% |
| ZW | 280 | 939 | 197 | 65582730601 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
