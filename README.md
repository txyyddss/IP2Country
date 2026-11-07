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

Last run: `2026-11-07T04:42:52Z`

Input rows: **39688630** · Published blocks: **2409006**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 561853 | 0 | 0 | 0 | 11091375 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 495045 | 0 | 0 | 15042161 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564711 | 0 | 0 | 0 | 11825981 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 494974 | 0 | 0 | 34153206 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 403488 | 0 | 0 | 0 | 8805351 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 474740 | 0 | 1 | 15609841 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 403391 | 0 | 0 | 0 | 6011040 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 474570 | 0 | 1 | 18442665 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 420669 | 0 | 0 | 0 | 9187361 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 566648 | 0 | 1 | 16824618 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1038161 | 0 | 0 | 0 | 13686543 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2930268 | 0 | 0 | 93936600 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1397018 | 3162970 | 0 | 0 | 224352529 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 389960 | 0 | 0 | 0 | 6095165 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1556723 | 0 | 0 | 24472437 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 798377 | 1392042 | 0 | 0 | 17530768 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177223 | 69825 | 0 | 0 | 3400135 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 294035 | 494411 | 0 | 98115 | 41312853 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 294035 | 494411 | 0 | 98635 | 18687179 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4346 | 1676 | 0 | 0 | 98038 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3880 | 0 | 0 | 0 | 82491 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 815122 | 434639 | 0 | 412110 | 18701195 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262543 | 71092 | 0 | 10 | 34614465 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802828 | 0 | 0 | 0 | 11122714 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6204 | 3432 | 0 | 0 | 153579 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 577 | 0 | 0 | 36049 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 564571 | 495852 | 0 | 0 | 3389216 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 614400 | 3702257408 | 1794356 | 293421847207842436812546030166054273024 |
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
| AE | 2820 | 25516 | 2568 | 12835044487897 | 89.85% |
| AF | 219 | 875 | 524 | 121136965519 | 97.42% |
| AG | 106 | 167 | 184 | 40393304124 | 65.26% |
| AI | 44 | 45 | 101 | 5847641145 | 67.40% |
| AL | 579 | 1414 | 1244 | 2500693679258 | 95.22% |
| AM | 349 | 2322 | 525 | 1263612958932 | 95.01% |
| AO | 605 | 5142 | 298 | 147599966562 | 98.68% |
| AQ | 56 | 5 | 283 | 472063214 | 43.66% |
| AR | 3767 | 76979 | 2824 | 23618121490663 | 98.83% |
| AS | 41 | 29 | 311 | 9243191474 | 96.04% |
| AT | 4750 | 45183 | 4492 | 10471310189943 | 97.55% |
| AU | 13338 | 220216 | 71577 | 2309962509067527260 | 100.00% |
| AW | 49 | 449 | 98 | 11534575673 | 92.96% |
| AX | 93 | 131 | 603 | 69552372553 | 72.57% |
| AZ | 398 | 3070 | 346 | 1964102738558 | 98.31% |
| BA | 337 | 3265 | 742 | 833827495454 | 95.13% |
| BB | 193 | 507 | 109 | 69400586958 | 99.69% |
| BD | 2661 | 8191 | 2713 | 3577593750770 | 96.48% |
| BE | 8515 | 57895 | 4826 | 4398024048007804 | 66.61% |
| BF | 124 | 1383 | 144 | 91245525303 | 99.43% |
| BG | 2425 | 16721 | 1434 | 5137769043777 | 96.12% |
| BH | 330 | 5246 | 320 | 406108644177 | 97.65% |
| BI | 57 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 85 | 622 | 126 | 56582572336 | 96.84% |
| BL | 39 | 10 | 85 | 1047717936 | 57.46% |
| BM | 316 | 496 | 107 | 43426851129 | 99.37% |
| BN | 110 | 879 | 119 | 26250616879 | 99.06% |
| BO | 327 | 4752 | 307 | 168302404201 | 98.07% |
| BQ | 38 | 103 | 69 | 17538738225 | 98.90% |
| BR | 11447 | 335458 | 42585 | 37423965139026 | 99.22% |
| BS | 139 | 633 | 163 | 14003518038 | 96.66% |
| BT | 123 | 226 | 164 | 60978750033 | 99.35% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 189 | 718 | 123 | 78628794672 | 97.43% |
| BY | 390 | 7400 | 428 | 700163287348 | 99.59% |
| BZ | 217 | 469 | 197 | 192579185482 | 89.06% |
| CA | 22117 | 285510 | 31549 | 6351227774626 | 95.82% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 320 | 940 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 118 | 36988538928 | 97.17% |
| CH | 6858 | 98181 | 57380 | 15878800943285074475 | 99.98% |
| CI | 348 | 6713 | 168 | 194921285944 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2081 | 41384 | 2376 | 1415240953030 | 96.61% |
| CM | 188 | 2438 | 160 | 67900151884 | 98.18% |
| CN | 6928 | 1347605 | 12720 | 294809140311534 | 99.94% |
| CO | 3171 | 70317 | 1946 | 1981632206736 | 98.03% |
| CR | 598 | 9164 | 369 | 461701217134 | 97.36% |
| CU | 96 | 1023 | 129 | 26439696697 | 99.17% |
| CV | 82 | 241 | 143 | 13715433519 | 97.22% |
| CW | 139 | 718 | 101 | 63125202989 | 92.43% |
| CX | 8 | 0 | 83 | 727509031 | 41.43% |
| CY | 759 | 4160 | 706 | 5999831227235 | 94.03% |
| CZ | 3874 | 33839 | 6526 | 13303848404132 | 97.55% |
| DE | 38499 | 550720 | 84725 | 115451222369373 | 96.06% |
| DJ | 73 | 246 | 119 | 9641322543 | 94.96% |
| DK | 3667 | 50106 | 2080 | 9667282682795 | 93.43% |
| DM | 74 | 71 | 116 | 44401879600 | 73.08% |
| DO | 736 | 8073 | 625 | 928508573027 | 98.99% |
| DZ | 232 | 18723 | 257 | 44176730298 | 95.48% |
| EC | 1087 | 11258 | 658 | 1158571649928 | 98.76% |
| EE | 1199 | 6501 | 1965 | 2725759163293 | 96.02% |
| EG | 794 | 94816 | 289 | 17730906038683 | 99.99% |
| EH | 4 | 0 | 54 | 457959449 | 46.78% |
| ER | 44 | 29 | 92 | 1100363819 | 57.01% |
| ES | 11274 | 139920 | 7000 | 26203112050157 | 97.15% |
| ET | 92 | 1478 | 133 | 1117547837495 | 99.96% |
| FI | 4015 | 58675 | 33556 | 6865768731112 | 96.76% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 38 | 25 | 84 | 612231207 | 46.31% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 86 | 108187543852 | 99.57% |
| FR | 27428 | 325641 | 21229 | 67379180444188 | 98.52% |
| GA | 91 | 1559 | 106 | 35779721512 | 94.29% |
| GB | 40214 | 498360 | 60595 | 105798192798376 | 94.13% |
| GD | 98 | 128 | 103 | 82705836587 | 96.88% |
| GE | 488 | 5114 | 355 | 1831824881784 | 95.38% |
| GF | 286 | 259 | 80 | 23010009128 | 97.39% |
| GG | 152 | 263 | 94 | 112129796399 | 81.66% |
| GH | 373 | 8926 | 264 | 1232412718691 | 99.85% |
| GI | 140 | 326 | 106 | 197924939820 | 90.89% |
| GL | 102 | 144 | 170 | 43621736572 | 99.27% |
| GM | 76 | 1104 | 109 | 39586920746 | 98.99% |
| GN | 74 | 167 | 116 | 48293804074 | 97.24% |
| GP | 240 | 802 | 102 | 54408919095 | 73.06% |
| GQ | 53 | 95 | 92 | 9454958635 | 96.02% |
| GR | 1532 | 23124 | 997 | 2592092025610 | 96.52% |
| GS | 4 | 2 | 41 | 406904854 | 59.19% |
| GT | 661 | 2886 | 276 | 327768439410 | 99.41% |
| GU | 181 | 995 | 130 | 34891948348 | 99.33% |
| GW | 51 | 42 | 94 | 9320609836 | 96.97% |
| GY | 57 | 288 | 108 | 35023939628 | 96.09% |
| HK | 13794 | 95167 | 8154 | 5086951464715 | 85.92% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 833 | 1992 | 266 | 391924755057 | 94.76% |
| HR | 837 | 9687 | 437 | 1591451061520 | 96.38% |
| HT | 83 | 624 | 133 | 26468805184 | 93.34% |
| HU | 1996 | 23376 | 1319 | 3217142900166 | 98.33% |
| ID | 6897 | 81283 | 5292 | 3843600478319 | 98.03% |
| IE | 8284 | 72339 | 8838 | 4921857128581 | 88.98% |
| IL | 2468 | 37167 | 1408 | 4297867840577 | 97.42% |
| IM | 331 | 475 | 190 | 361489639998 | 96.14% |
| IN | 13356 | 197343 | 17018 | 28970331567719 | 97.84% |
| IO | 25 | 14 | 83 | 4755679275 | 95.94% |
| IQ | 673 | 3347 | 963 | 3415584355222 | 96.90% |
| IR | 1973 | 42727 | 2206 | 16962237631368 | 98.49% |
| IS | 339 | 3639 | 492 | 1596995666811 | 96.34% |
| IT | 9950 | 227761 | 9579 | 47046789635152 | 99.09% |
| JE | 223 | 641 | 116 | 219604756782 | 95.45% |
| JM | 251 | 1131 | 152 | 1122142710361 | 99.96% |
| JO | 441 | 2701 | 253 | 1121936402780 | 97.90% |
| JP | 12525 | 782385 | 14902 | 53257923800237 | 88.23% |
| KE | 1147 | 25119 | 666 | 688052349112 | 99.34% |
| KG | 236 | 1184 | 218 | 529225198187 | 99.38% |
| KH | 558 | 2014 | 301 | 430692688488 | 81.58% |
| KI | 79 | 36 | 137 | 13520201803 | 97.75% |
| KM | 66 | 52 | 85 | 9337298983 | 96.78% |
| KN | 99 | 133 | 103 | 5765065014 | 76.99% |
| KP | 9 | 11 | 416 | 500187428 | 55.29% |
| KR | 4037 | 455069 | 3736 | 22712188282206 | 99.76% |
| KW | 425 | 7581 | 218 | 877741035968 | 99.72% |
| KY | 123 | 437 | 106 | 17861635378 | 99.08% |
| KZ | 1393 | 13974 | 1208 | 3150641268013 | 86.82% |
| LA | 146 | 447 | 151 | 121260925231 | 88.17% |
| LB | 392 | 2388 | 294 | 3016060838727 | 98.08% |
| LC | 103 | 168 | 90 | 78392322341 | 97.53% |
| LI | 299 | 453 | 201 | 429947419690 | 93.40% |
| LK | 250 | 2326 | 265 | 82540585332 | 94.31% |
| LR | 83 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1442 | 12005 | 965 | 4211635078928 | 84.64% |
| LU | 1549 | 6703 | 879 | 2222679225977 | 96.04% |
| LV | 1247 | 8158 | 769 | 3235774962912 | 65.66% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 340 | 48132 | 265 | 1171100463781 | 99.78% |
| MC | 285 | 284 | 108 | 142166649391 | 99.88% |
| MD | 761 | 4975 | 974 | 5953066380627 | 64.08% |
| ME | 154 | 828 | 129 | 348120038202 | 99.48% |
| MF | 56 | 103 | 88 | 5443266607 | 92.31% |
| MG | 85 | 2307 | 114 | 22644159028 | 98.09% |
| MH | 84 | 94 | 81 | 4806418475 | 95.24% |
| MK | 254 | 2800 | 186 | 1142897781613 | 97.13% |
| ML | 81 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1170 | 331 | 353155015028 | 97.86% |
| MN | 236 | 737 | 283 | 133921418858 | 99.23% |
| MO | 106 | 1500 | 706 | 39116173878 | 99.15% |
| MP | 91 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 335 | 844 | 70 | 104815667236 | 89.69% |
| MR | 73 | 233 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 77 | 644671791 | 51.26% |
| MT | 325 | 2681 | 230 | 509630147384 | 92.89% |
| MU | 2090 | 23867 | 8213 | 4455188187869499 | 100.00% |
| MV | 119 | 403 | 135 | 18180595767 | 97.48% |
| MW | 123 | 2200 | 146 | 1242315518001 | 99.85% |
| MX | 4391 | 122514 | 2603 | 3642801370787 | 97.00% |
| MY | 2830 | 34573 | 1622 | 911924937798 | 96.08% |
| MZ | 195 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 411 | 1856 | 186 | 29169103150 | 99.37% |
| NC | 76 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1118 | 12923 | 872 | 604382931070 | 98.31% |
| NI | 168 | 1731 | 120 | 95171111236 | 98.22% |
| NL | 25004 | 214492 | 497073 | 8408185627986205 | 99.94% |
| NO | 3292 | 63892 | 3053 | 10986135353121 | 97.26% |
| NP | 469 | 2390 | 343 | 430195425381 | 94.05% |
| NR | 41 | 52 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2405 | 28872 | 1787 | 1359417217362 | 94.87% |
| OM | 308 | 4416 | 202 | 595072376918 | 96.04% |
| PA | 536 | 7090 | 469 | 429360248116 | 92.79% |
| PE | 1256 | 14659 | 1078 | 1072415448593 | 96.36% |
| PF | 93 | 312 | 144 | 21875058781 | 99.30% |
| PG | 231 | 347 | 153 | 73449136460 | 87.10% |
| PH | 2241 | 26838 | 1271 | 1042109604357 | 97.81% |
| PK | 1446 | 23643 | 744 | 907296026873 | 97.95% |
| PL | 8724 | 83072 | 5355 | 28613407337931 | 96.79% |
| PM | 42 | 25 | 80 | 4671531309 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 641 | 4572 | 346 | 230687497864 | 99.07% |
| PS | 302 | 3305 | 178 | 979424835919 | 93.84% |
| PT | 1910 | 27057 | 2076 | 3328261766358 | 93.01% |
| PW | 47 | 43 | 87 | 9085575215 | 97.53% |
| PY | 437 | 5086 | 270 | 451733561976 | 97.52% |
| QA | 492 | 5196 | 366 | 526445600732 | 96.23% |
| RE | 366 | 1997 | 99 | 95622438961 | 96.44% |
| RO | 3927 | 28289 | 2488 | 5569029810664 | 95.61% |
| RS | 1077 | 9076 | 1349 | 1887339660726 | 97.41% |
| RU | 13752 | 180678 | 15928 | 54566193930318 | 90.19% |
| RW | 689 | 1264 | 125 | 1147559424301 | 99.91% |
| SA | 1713 | 48716 | 921 | 2815427171817 | 96.58% |
| SB | 55 | 71 | 99 | 17944013095 | 90.06% |
| SC | 884 | 2557 | 3200 | 3327139111469 | 62.03% |
| SD | 115 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16284 | 126464 | 61345 | 30392834030288 | 79.90% |
| SG | 11084 | 111055 | 52549 | 133650227863925 | 98.93% |
| SH | 6 | 2 | 130 | 574482436 | 54.36% |
| SI | 1066 | 10334 | 621 | 2079455901623 | 95.52% |
| SJ | 7 | 2 | 40 | 374466579 | 60.68% |
| SK | 1221 | 11163 | 931 | 3009684262163 | 97.04% |
| SL | 125 | 358 | 111 | 26584803376 | 96.08% |
| SM | 73 | 173 | 98 | 202275613222 | 92.49% |
| SN | 119 | 1645 | 149 | 27161196614 | 98.08% |
| SO | 75 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 341 | 107 | 17813725490 | 98.45% |
| SS | 99 | 135 | 110 | 31098290222 | 98.66% |
| ST | 47 | 60 | 90 | 4773832747 | 96.05% |
| SV | 333 | 2776 | 208 | 211465050985 | 97.09% |
| SX | 48 | 128 | 81 | 30575618087 | 94.55% |
| SY | 296 | 5171 | 138 | 876931813423 | 99.64% |
| SZ | 76 | 242 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 83 | 1451583785 | 71.10% |
| TD | 93 | 133 | 104 | 13816293422 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2299 | 43580 | 1265 | 595585280338 | 95.53% |
| TJ | 157 | 401 | 167 | 641204726332 | 98.14% |
| TK | 25 | 15 | 114 | 4789889072 | 95.34% |
| TL | 93 | 126 | 109 | 43596832814 | 92.49% |
| TM | 63 | 126 | 102 | 69114915129 | 99.77% |
| TN | 148 | 30817 | 148 | 116780507711 | 97.52% |
| TO | 49 | 56 | 85 | 17523466282 | 99.25% |
| TR | 3588 | 68037 | 2883 | 22512863245686 | 93.84% |
| TT | 167 | 2226 | 144 | 1177257585227 | 70.56% |
| TV | 40 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9263 | 160469 | 3111 | 11552662477866 | 98.66% |
| TZ | 452 | 4500 | 248 | 267366409332 | 98.77% |
| UA | 4887 | 34671 | 3311 | 11756586720903 | 93.14% |
| UG | 269 | 5773 | 182 | 132375003453 | 96.35% |
| UM | 5 | 2 | 84 | 355002414 | 58.92% |
| US | 136054 | 5871984 | 543313 | 4835783197533922 | 99.76% |
| UY | 246 | 10117 | 192 | 177623867458 | 99.80% |
| UZ | 549 | 1661 | 274 | 1083134229101 | 99.45% |
| VA | 36 | 51 | 107 | 103325509933 | 99.88% |
| VC | 84 | 103 | 89 | 69500512298 | 99.67% |
| VE | 968 | 30268 | 658 | 2279265730999 | 99.33% |
| VG | 337 | 271 | 420 | 251178960947 | 90.61% |
| VI | 88 | 547 | 159 | 30643734597 | 94.69% |
| VN | 2258 | 65528 | 1664 | 1860194014225 | 99.13% |
| VU | 66 | 87 | 274 | 51951561521 | 80.96% |
| WF | 21 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 184 | 623 | 76 | 111807037430 | 60.60% |
| YE | 113 | 958 | 119 | 207096833082 | 99.81% |
| YT | 46 | 48 | 76 | 602988585 | 56.00% |
| ZA | 5221 | 117609 | 3381 | 41031802180933 | 78.30% |
| ZM | 387 | 6516 | 161 | 48484905299 | 96.08% |
| ZW | 280 | 939 | 197 | 65582730601 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
