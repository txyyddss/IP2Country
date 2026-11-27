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

Last run: `2026-11-27T06:47:50Z`

Input rows: **40636639** · Published blocks: **2371307**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 572377 | 0 | 0 | 0 | 11298949 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 500651 | 0 | 0 | 15254894 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 573842 | 0 | 0 | 0 | 12017274 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 500773 | 0 | 0 | 34553337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 410091 | 0 | 0 | 0 | 8942009 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 489000 | 0 | 1 | 16145289 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 409856 | 0 | 0 | 0 | 6101222 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 488635 | 0 | 1 | 19037514 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 429038 | 0 | 0 | 0 | 9352775 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 582548 | 0 | 1 | 17410699 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1090217 | 0 | 0 | 0 | 13967301 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 3109580 | 0 | 0 | 54872043 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1447887 | 3349202 | 0 | 0 | 219495721 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 806371 | 1412280 | 0 | 0 | 17756624 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177554 | 69982 | 0 | 0 | 3407965 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 303205 | 324864 | 0 | 8 | 29317705 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 303205 | 324864 | 0 | 8 | 13437327 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4343 | 1726 | 0 | 0 | 98795 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3882 | 0 | 0 | 0 | 82959 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327930 | 0 | 0 | 0 | 537841 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 819428 | 436767 | 0 | 412894 | 18659568 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262823 | 71239 | 0 | 9 | 34658874 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4957269 | 0 | 0 | 0 | 66550081 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380534 | 0 | 0 | 11707152 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 803015 | 0 | 0 | 0 | 11123858 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6235 | 3450 | 0 | 0 | 154474 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1751 | 577 | 0 | 0 | 36096 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 574636 | 501526 | 0 | 0 | 3410892 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 614321 | 3702257408 | 1756736 | 293421847207842436812546030166054273024 |
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
| AD | 216 | 308 | 190 | 43728565320 | 92.57% |
| AE | 2840 | 25465 | 2848 | 12757802464218 | 90.24% |
| AF | 224 | 875 | 477 | 121136965472 | 97.42% |
| AG | 113 | 166 | 186 | 40393307195 | 65.28% |
| AI | 45 | 45 | 100 | 5847643190 | 79.64% |
| AL | 553 | 1410 | 1268 | 2508965240734 | 95.10% |
| AM | 362 | 2325 | 525 | 1293644175574 | 94.63% |
| AO | 517 | 5123 | 507 | 147599979092 | 98.68% |
| AQ | 59 | 5 | 287 | 404842625 | 44.24% |
| AR | 3897 | 77019 | 2407 | 23635284631748 | 98.86% |
| AS | 42 | 29 | 317 | 9243193524 | 96.04% |
| AT | 4738 | 45105 | 4236 | 10501627723253 | 97.49% |
| AU | 13538 | 219884 | 71465 | 2309962465005491159 | 100.00% |
| AW | 51 | 449 | 96 | 11534578740 | 92.96% |
| AX | 77 | 132 | 608 | 69552506698 | 72.57% |
| AZ | 399 | 3071 | 337 | 1963968488043 | 98.36% |
| BA | 325 | 3264 | 744 | 833827513885 | 95.13% |
| BB | 195 | 507 | 109 | 69400590027 | 99.70% |
| BD | 2666 | 8217 | 2596 | 3581889172776 | 96.45% |
| BE | 8550 | 58136 | 4870 | 4398097288931128 | 66.62% |
| BF | 130 | 1388 | 145 | 91245528373 | 99.43% |
| BG | 2366 | 16710 | 1436 | 5129179011889 | 96.11% |
| BH | 340 | 5248 | 317 | 406108709709 | 97.65% |
| BI | 60 | 168 | 109 | 26952917544 | 92.82% |
| BJ | 93 | 622 | 128 | 56582575407 | 96.84% |
| BL | 24 | 10 | 87 | 1047851055 | 57.46% |
| BM | 303 | 500 | 107 | 43426918710 | 99.37% |
| BN | 112 | 878 | 118 | 26250616878 | 99.06% |
| BO | 327 | 4753 | 306 | 168302404200 | 98.07% |
| BQ | 39 | 103 | 70 | 17538741295 | 98.90% |
| BR | 11726 | 335516 | 35869 | 37427976358371 | 99.23% |
| BS | 141 | 634 | 163 | 14003799640 | 96.66% |
| BT | 126 | 226 | 157 | 60963021390 | 99.36% |
| BV | 1 | 1 | 3 | 17825792 | 34.31% |
| BW | 218 | 722 | 123 | 82923761967 | 96.40% |
| BY | 396 | 7406 | 442 | 700163615037 | 99.49% |
| BZ | 215 | 469 | 199 | 192579188553 | 89.06% |
| CA | 19852 | 285236 | 33076 | 6442150363912 | 95.59% |
| CC | 7 | 0 | 50 | 827256852 | 42.84% |
| CD | 356 | 940 | 203 | 155963863379 | 98.74% |
| CF | 50 | 40 | 96 | 1117335591 | 59.87% |
| CG | 94 | 586 | 116 | 36988538926 | 97.17% |
| CH | 7002 | 98595 | 57291 | 15878800912925665426 | 99.98% |
| CI | 349 | 6713 | 167 | 194887800117 | 98.94% |
| CK | 36 | 41 | 92 | 9234282542 | 96.05% |
| CL | 2088 | 41384 | 2034 | 1419552346774 | 96.68% |
| CM | 192 | 2438 | 162 | 67900154955 | 98.18% |
| CN | 6948 | 1347808 | 13010 | 294845036705758 | 99.94% |
| CO | 3215 | 70317 | 1922 | 1994517177200 | 97.92% |
| CR | 596 | 9169 | 371 | 461702268525 | 97.36% |
| CU | 95 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 84 | 241 | 145 | 13715370030 | 97.22% |
| CW | 140 | 718 | 103 | 63125206060 | 92.42% |
| CX | 7 | 0 | 81 | 727429155 | 41.43% |
| CY | 683 | 4140 | 711 | 6042768253794 | 93.92% |
| CZ | 3925 | 33841 | 6292 | 13303777558428 | 97.50% |
| DE | 38719 | 550850 | 82232 | 115551039375157 | 96.09% |
| DJ | 66 | 247 | 122 | 9641325614 | 94.96% |
| DK | 3651 | 50106 | 2117 | 9637217826981 | 93.52% |
| DM | 76 | 71 | 116 | 44401817134 | 73.08% |
| DO | 744 | 8097 | 632 | 941393510755 | 98.70% |
| DZ | 239 | 18724 | 243 | 47397892265 | 92.39% |
| EC | 1099 | 11267 | 663 | 1171457603462 | 98.47% |
| EE | 1202 | 6498 | 1523 | 2725767437087 | 96.04% |
| EG | 806 | 94821 | 276 | 17726575600021 | 100.00% |
| EH | 4 | 0 | 54 | 457961495 | 46.78% |
| ER | 47 | 29 | 95 | 1100366891 | 57.01% |
| ES | 11475 | 139877 | 6441 | 26203111871613 | 97.15% |
| ET | 97 | 1483 | 134 | 1117547839542 | 99.96% |
| FI | 4144 | 58796 | 33572 | 6895814559506 | 96.80% |
| FJ | 115 | 590 | 118 | 43379783467 | 99.48% |
| FK | 40 | 26 | 83 | 578678822 | 46.09% |
| FM | 59 | 49 | 107 | 8921740345 | 78.74% |
| FO | 47 | 181 | 88 | 108187546923 | 99.57% |
| FR | 27889 | 325784 | 18462 | 67659301670172 | 98.41% |
| GA | 92 | 1558 | 106 | 35779724581 | 94.29% |
| GB | 41597 | 498428 | 61043 | 105974292125871 | 93.90% |
| GD | 102 | 128 | 104 | 82705774122 | 96.88% |
| GE | 492 | 5119 | 363 | 1831824881791 | 95.38% |
| GF | 294 | 261 | 79 | 23010009127 | 97.39% |
| GG | 142 | 268 | 96 | 112129799470 | 81.66% |
| GH | 362 | 8925 | 264 | 1232362655330 | 99.85% |
| GI | 155 | 329 | 107 | 197924942890 | 90.89% |
| GL | 104 | 144 | 166 | 43621802103 | 99.27% |
| GM | 77 | 1102 | 110 | 39586922793 | 98.99% |
| GN | 78 | 168 | 117 | 48293807144 | 97.24% |
| GP | 227 | 808 | 104 | 54408922166 | 73.06% |
| GQ | 55 | 94 | 94 | 9454961706 | 96.02% |
| GR | 1555 | 23125 | 885 | 2592596398215 | 96.51% |
| GS | 4 | 2 | 40 | 406904853 | 59.18% |
| GT | 671 | 2899 | 281 | 327769491059 | 99.41% |
| GU | 184 | 996 | 121 | 34757222715 | 99.56% |
| GW | 54 | 42 | 95 | 9320611883 | 96.97% |
| GY | 59 | 288 | 109 | 35023941675 | 96.09% |
| HK | 13939 | 95098 | 8291 | 5099938585189 | 85.97% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 846 | 1995 | 268 | 391925675633 | 94.76% |
| HR | 882 | 9691 | 435 | 1591451098328 | 96.38% |
| HT | 88 | 624 | 134 | 26468807231 | 93.34% |
| HU | 2077 | 23382 | 1551 | 3217144546928 | 98.33% |
| ID | 6913 | 81271 | 5312 | 3830720295023 | 98.20% |
| IE | 8294 | 72102 | 8669 | 4986687822386 | 88.64% |
| IL | 2520 | 37067 | 1451 | 4296153420908 | 97.59% |
| IM | 352 | 477 | 192 | 361490691644 | 96.14% |
| IN | 13535 | 197430 | 14190 | 28991840813008 | 97.87% |
| IO | 29 | 14 | 83 | 4755682344 | 95.94% |
| IQ | 693 | 3347 | 845 | 3415585403677 | 96.90% |
| IR | 1981 | 42741 | 2211 | 17060955966344 | 97.98% |
| IS | 342 | 3642 | 485 | 1562838145905 | 97.32% |
| IT | 10222 | 227805 | 8617 | 47055420340510 | 99.07% |
| JE | 218 | 636 | 117 | 219604759852 | 95.45% |
| JM | 254 | 1129 | 156 | 1122143762009 | 99.96% |
| JO | 451 | 2702 | 256 | 1121936404829 | 97.90% |
| JP | 12833 | 782654 | 13681 | 53267695102862 | 88.22% |
| KE | 1273 | 25127 | 666 | 696608794810 | 99.07% |
| KG | 240 | 1186 | 212 | 529225201250 | 99.38% |
| KH | 555 | 2011 | 298 | 430675911270 | 81.58% |
| KI | 80 | 36 | 137 | 13520203849 | 97.88% |
| KM | 68 | 52 | 81 | 9320521764 | 96.86% |
| KN | 95 | 106 | 123 | 14354937141 | 66.82% |
| KP | 9 | 11 | 423 | 500515115 | 55.26% |
| KR | 4120 | 455054 | 3538 | 22712176556649 | 99.88% |
| KW | 442 | 7581 | 219 | 877742051777 | 99.72% |
| KY | 127 | 450 | 108 | 17861638449 | 99.08% |
| KZ | 1388 | 13991 | 1228 | 3150641336140 | 86.93% |
| LA | 147 | 402 | 150 | 121260925230 | 88.17% |
| LB | 392 | 2384 | 284 | 3016027284287 | 97.70% |
| LC | 106 | 168 | 93 | 78392390948 | 97.53% |
| LI | 286 | 451 | 201 | 429948732453 | 91.87% |
| LK | 257 | 2325 | 261 | 82523791730 | 94.32% |
| LR | 85 | 440 | 107 | 27275920681 | 97.70% |
| LS | 69 | 507 | 93 | 26802794535 | 98.25% |
| LT | 1443 | 11970 | 969 | 4177006780183 | 83.88% |
| LU | 1588 | 6714 | 863 | 2222694958211 | 96.44% |
| LV | 1246 | 8166 | 782 | 3235474534630 | 65.64% |
| LY | 153 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 354 | 48137 | 261 | 1171100465823 | 99.78% |
| MC | 280 | 281 | 108 | 142166650414 | 99.88% |
| MD | 753 | 4973 | 976 | 5953032877391 | 64.04% |
| ME | 147 | 822 | 132 | 348120630073 | 99.48% |
| MF | 60 | 104 | 90 | 5443335213 | 92.30% |
| MG | 87 | 2308 | 112 | 22627381811 | 98.13% |
| MH | 75 | 83 | 79 | 4806418473 | 95.22% |
| MK | 257 | 2800 | 187 | 1138602865495 | 97.19% |
| ML | 84 | 388 | 87 | 22658490412 | 97.55% |
| MM | 462 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 239 | 737 | 275 | 133921418850 | 99.23% |
| MO | 105 | 1500 | 736 | 39117660493 | 99.15% |
| MP | 83 | 15 | 78 | 4805746733 | 95.20% |
| MQ | 303 | 810 | 72 | 104815735842 | 89.69% |
| MR | 77 | 234 | 97 | 26517696554 | 98.88% |
| MS | 33 | 15 | 79 | 644674607 | 51.26% |
| MT | 325 | 2677 | 232 | 508545920828 | 92.95% |
| MU | 2018 | 23914 | 8199 | 4455162418082120 | 100.00% |
| MV | 120 | 403 | 131 | 18180677684 | 97.49% |
| MW | 125 | 2200 | 145 | 1242315518000 | 99.85% |
| MX | 4580 | 122528 | 2551 | 3651392486448 | 96.80% |
| MY | 2806 | 34648 | 1634 | 911924741717 | 96.17% |
| MZ | 191 | 1912 | 132 | 69652197436 | 97.38% |
| NA | 297 | 1860 | 184 | 29152325933 | 99.39% |
| NC | 79 | 571 | 109 | 47910351157 | 99.48% |
| NE | 64 | 218 | 90 | 22105940004 | 98.76% |
| NF | 37 | 3 | 85 | 357560364 | 59.06% |
| NG | 1122 | 12922 | 710 | 604382767233 | 98.31% |
| NI | 172 | 1736 | 123 | 95171114308 | 98.22% |
| NL | 24120 | 214562 | 487305 | 8406844303443854 | 99.94% |
| NO | 3443 | 63971 | 2460 | 10986017998645 | 97.32% |
| NP | 471 | 2393 | 339 | 430213185632 | 94.05% |
| NR | 43 | 53 | 90 | 8951231533 | 98.41% |
| NU | 28 | 11 | 73 | 4654560295 | 81.58% |
| NZ | 2299 | 28845 | 1837 | 1360222344288 | 94.94% |
| OM | 324 | 4418 | 203 | 595055599704 | 96.04% |
| PA | 558 | 7095 | 464 | 427213619508 | 93.35% |
| PE | 1304 | 14783 | 1073 | 1085300366853 | 96.10% |
| PF | 94 | 312 | 144 | 21875061850 | 99.30% |
| PG | 234 | 348 | 152 | 73449136459 | 87.10% |
| PH | 2291 | 26817 | 1444 | 1054992607498 | 97.48% |
| PK | 1469 | 23617 | 741 | 907296026852 | 98.17% |
| PL | 8883 | 83160 | 5243 | 28608832203704 | 97.09% |
| PM | 43 | 25 | 79 | 4671532331 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.40% |
| PR | 651 | 4572 | 332 | 230687368073 | 98.76% |
| PS | 301 | 3298 | 176 | 979424837963 | 93.84% |
| PT | 1964 | 27060 | 1851 | 3362639241100 | 92.85% |
| PW | 49 | 43 | 84 | 9085575212 | 97.53% |
| PY | 458 | 5092 | 275 | 451733299856 | 97.52% |
| QA | 527 | 5197 | 360 | 560788496330 | 94.93% |
| RE | 323 | 1983 | 96 | 95622389806 | 96.43% |
| RO | 4115 | 28290 | 2169 | 5568744413160 | 95.78% |
| RS | 1079 | 9068 | 1317 | 1887339431630 | 97.41% |
| RU | 13826 | 181041 | 16850 | 54798394220800 | 90.11% |
| RW | 174 | 1266 | 122 | 1147559424298 | 99.91% |
| SA | 1710 | 48873 | 924 | 2815443747391 | 96.57% |
| SB | 56 | 71 | 100 | 17944015142 | 94.05% |
| SC | 859 | 2573 | 2995 | 3256762933801 | 62.42% |
| SD | 118 | 7412 | 110 | 69736980787 | 97.36% |
| SE | 15283 | 126603 | 53758 | 30393168416134 | 79.90% |
| SG | 10826 | 110838 | 53162 | 133706055774527 | 98.92% |
| SH | 6 | 2 | 130 | 575140868 | 54.82% |
| SI | 1045 | 10333 | 623 | 2079447594940 | 95.39% |
| SJ | 7 | 2 | 42 | 374469650 | 60.67% |
| SK | 1229 | 11154 | 949 | 3009701131543 | 96.99% |
| SL | 127 | 359 | 112 | 26584806446 | 98.77% |
| SM | 74 | 172 | 98 | 202275614245 | 92.49% |
| SN | 122 | 1646 | 150 | 27161199684 | 98.08% |
| SO | 76 | 224 | 160 | 48216207680 | 99.14% |
| SR | 53 | 340 | 104 | 17780171057 | 98.54% |
| SS | 103 | 135 | 108 | 31098290220 | 98.66% |
| ST | 50 | 60 | 92 | 4773835818 | 96.05% |
| SV | 337 | 2786 | 209 | 211465054055 | 97.10% |
| SX | 65 | 128 | 82 | 30575621157 | 94.55% |
| SY | 297 | 5171 | 138 | 945651290157 | 97.85% |
| SZ | 78 | 243 | 104 | 34670981172 | 96.15% |
| TC | 45 | 56 | 83 | 1451537704 | 71.10% |
| TD | 95 | 133 | 105 | 13816296492 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 73 | 1403 | 112 | 26517259558 | 98.88% |
| TH | 2304 | 43569 | 1268 | 595836857921 | 95.69% |
| TJ | 161 | 403 | 166 | 641204728377 | 98.14% |
| TK | 27 | 15 | 114 | 4789891118 | 95.34% |
| TL | 85 | 121 | 107 | 43596832812 | 92.49% |
| TM | 63 | 126 | 102 | 69081429303 | 99.79% |
| TN | 153 | 30817 | 144 | 116746955323 | 97.54% |
| TO | 49 | 55 | 83 | 17523466280 | 99.25% |
| TR | 3765 | 68060 | 2992 | 22538500591254 | 93.73% |
| TT | 169 | 2226 | 146 | 1177257588298 | 70.56% |
| TV | 42 | 46 | 98 | 11112280102 | 83.81% |
| TW | 9346 | 160439 | 3157 | 11549444964675 | 98.68% |
| TZ | 575 | 4494 | 247 | 267349632116 | 98.78% |
| UA | 4970 | 34572 | 3311 | 11756588163557 | 93.17% |
| UG | 266 | 5773 | 180 | 132375003450 | 96.35% |
| UM | 5 | 2 | 83 | 354939947 | 58.92% |
| US | 135037 | 5870489 | 540225 | 4835794990671808 | 99.76% |
| UY | 247 | 10117 | 191 | 177623867457 | 99.80% |
| UZ | 553 | 1661 | 272 | 1083134183018 | 99.45% |
| VA | 36 | 51 | 103 | 103326428457 | 99.88% |
| VC | 82 | 103 | 90 | 69500515368 | 99.67% |
| VE | 986 | 30432 | 655 | 2283578524081 | 99.32% |
| VG | 358 | 312 | 260 | 255456170034 | 92.19% |
| VI | 101 | 548 | 170 | 30615426118 | 94.73% |
| VN | 2288 | 65518 | 1805 | 1860462171161 | 99.17% |
| VU | 67 | 86 | 274 | 51951564590 | 80.96% |
| WF | 23 | 16 | 71 | 4788515881 | 95.30% |
| WS | 63 | 97 | 89 | 21837242666 | 92.81% |
| XK | 203 | 626 | 72 | 111806775286 | 60.60% |
| YE | 119 | 960 | 118 | 207096833081 | 99.81% |
| YT | 82 | 51 | 74 | 602988583 | 56.00% |
| ZA | 5170 | 117559 | 2759 | 41040391955558 | 78.31% |
| ZM | 325 | 6508 | 162 | 52779872595 | 92.33% |
| ZW | 252 | 924 | 302 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
