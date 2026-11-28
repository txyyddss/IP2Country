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

Last run: `2026-11-28T06:21:07Z`

Input rows: **37998991** · Published blocks: **2375195**

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
| Server-Country via sapics IPv4 | fresh | 410297 | 0 | 0 | 0 | 8946792 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 488996 | 0 | 1 | 16146460 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 410091 | 0 | 0 | 0 | 6105549 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 489000 | 0 | 1 | 19044828 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 429243 | 0 | 0 | 0 | 9357525 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 582544 | 0 | 1 | 17411870 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1053265 | 0 | 0 | 0 | 13609923 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 1823341 | 0 | 0 | 43325997 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1414261 | 2066593 | 0 | 0 | 166115144 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 806513 | 1412736 | 0 | 0 | 17761408 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177578 | 69977 | 0 | 0 | 3408079 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 303205 | 324864 | 0 | 8 | 29317705 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 303205 | 324864 | 0 | 8 | 13437327 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4343 | 1727 | 0 | 0 | 98815 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3882 | 0 | 0 | 0 | 82959 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327930 | 0 | 0 | 0 | 537841 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 819521 | 436769 | 0 | 412925 | 18661145 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262862 | 71263 | 0 | 9 | 34665421 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4957269 | 0 | 0 | 0 | 66550081 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380534 | 0 | 0 | 11707152 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 803015 | 0 | 0 | 0 | 11123858 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6235 | 3449 | 0 | 0 | 154457 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1751 | 577 | 0 | 0 | 36096 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 574636 | 501526 | 0 | 0 | 3410892 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 628787 | 3702257408 | 1746158 | 293421847207842436812546030166054273024 |
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
| AD | 216 | 308 | 188 | 43729089605 | 92.57% |
| AE | 2825 | 25514 | 2533 | 12787799897027 | 90.53% |
| AF | 225 | 875 | 484 | 121136965479 | 97.42% |
| AG | 138 | 178 | 184 | 40393307192 | 65.28% |
| AI | 45 | 45 | 100 | 5847577655 | 79.64% |
| AL | 556 | 1416 | 1222 | 2504668640920 | 95.09% |
| AM | 360 | 2325 | 513 | 1293644175562 | 94.68% |
| AO | 540 | 5122 | 382 | 147599913554 | 98.68% |
| AQ | 59 | 5 | 301 | 405104771 | 44.19% |
| AR | 3873 | 77013 | 2867 | 23635048716611 | 98.86% |
| AS | 42 | 29 | 307 | 9242865836 | 96.05% |
| AT | 4806 | 45081 | 4205 | 10497348642682 | 97.52% |
| AU | 13524 | 219613 | 69281 | 2309962450473299717 | 100.00% |
| AW | 51 | 449 | 93 | 11534529587 | 92.96% |
| AX | 80 | 133 | 604 | 69539858249 | 72.57% |
| AZ | 403 | 3073 | 340 | 1963968553586 | 98.36% |
| BA | 325 | 3264 | 744 | 833827534396 | 95.13% |
| BB | 195 | 507 | 103 | 69400458797 | 99.70% |
| BD | 2664 | 8218 | 2576 | 3590479173780 | 96.38% |
| BE | 8998 | 58146 | 3796 | 4398091927418641 | 66.62% |
| BF | 133 | 1387 | 145 | 91245528373 | 99.43% |
| BG | 2397 | 16705 | 1500 | 5129178552843 | 96.11% |
| BH | 339 | 4976 | 319 | 406108644176 | 97.65% |
| BI | 61 | 169 | 109 | 26952917544 | 92.82% |
| BJ | 94 | 623 | 128 | 56582575407 | 96.84% |
| BL | 24 | 10 | 85 | 1047851053 | 57.46% |
| BM | 317 | 501 | 108 | 43426918711 | 99.37% |
| BN | 113 | 878 | 118 | 26250616879 | 99.06% |
| BO | 336 | 4754 | 306 | 168302404200 | 98.07% |
| BQ | 39 | 103 | 70 | 21833643055 | 91.24% |
| BR | 11745 | 334950 | 34371 | 37440874366475 | 99.22% |
| BS | 141 | 634 | 175 | 17996662357 | 84.09% |
| BT | 126 | 226 | 157 | 60963021390 | 99.36% |
| BV | 1 | 1 | 3 | 17825792 | 34.31% |
| BW | 186 | 723 | 123 | 82923761967 | 96.40% |
| BY | 396 | 7406 | 425 | 700163615020 | 99.49% |
| BZ | 221 | 468 | 183 | 188586276681 | 90.09% |
| CA | 20090 | 283673 | 24098 | 6442060031509 | 95.33% |
| CC | 7 | 0 | 51 | 827256853 | 42.84% |
| CD | 347 | 940 | 201 | 155963797842 | 98.74% |
| CF | 50 | 40 | 97 | 1117335592 | 59.87% |
| CG | 94 | 586 | 116 | 36988538926 | 97.17% |
| CH | 6999 | 98586 | 57190 | 15878800900020021546 | 99.98% |
| CI | 350 | 6713 | 167 | 194887734582 | 98.94% |
| CK | 36 | 41 | 92 | 9234282542 | 96.05% |
| CL | 2097 | 41393 | 2061 | 1419556809166 | 96.68% |
| CM | 193 | 2439 | 162 | 67900154955 | 98.18% |
| CN | 6976 | 1348458 | 12862 | 294845227332818 | 99.94% |
| CO | 3217 | 70316 | 1868 | 1994516930231 | 97.97% |
| CR | 594 | 9167 | 361 | 461702202977 | 97.36% |
| CU | 94 | 1022 | 127 | 26439631160 | 99.17% |
| CV | 84 | 241 | 145 | 13715370030 | 97.22% |
| CW | 170 | 718 | 102 | 63124943916 | 92.45% |
| CX | 7 | 0 | 80 | 727429154 | 41.43% |
| CY | 706 | 4151 | 720 | 6042768229986 | 94.51% |
| CZ | 3911 | 33840 | 6453 | 13303781088145 | 97.50% |
| DE | 40596 | 550761 | 83518 | 115514206855730 | 96.10% |
| DJ | 68 | 286 | 122 | 9641325614 | 94.96% |
| DK | 3727 | 50110 | 2108 | 9637217075581 | 93.37% |
| DM | 79 | 71 | 115 | 44401817133 | 73.08% |
| DO | 744 | 8095 | 629 | 945682022754 | 98.56% |
| DZ | 239 | 18724 | 242 | 47397892264 | 92.29% |
| EC | 1119 | 11279 | 659 | 1171423901574 | 98.47% |
| EE | 1232 | 6512 | 1420 | 2725767830209 | 96.28% |
| EG | 812 | 94822 | 278 | 17726577369491 | 100.00% |
| EH | 4 | 0 | 54 | 457961495 | 46.78% |
| ER | 47 | 29 | 94 | 1100366890 | 57.01% |
| ES | 11854 | 139959 | 6756 | 26173008287410 | 97.18% |
| ET | 99 | 1484 | 133 | 1117547774006 | 99.96% |
| FI | 4233 | 58799 | 37114 | 6895827105044 | 96.80% |
| FJ | 115 | 590 | 118 | 43379783467 | 99.48% |
| FK | 40 | 26 | 82 | 578613286 | 46.09% |
| FM | 59 | 49 | 106 | 8921478201 | 78.74% |
| FO | 47 | 181 | 86 | 108187555622 | 99.57% |
| FR | 29767 | 325590 | 18563 | 67693577124193 | 98.36% |
| GA | 92 | 1558 | 107 | 35779708197 | 94.29% |
| GB | 44975 | 497927 | 69379 | 106025162888109 | 93.87% |
| GD | 102 | 128 | 104 | 82705774122 | 96.88% |
| GE | 494 | 5120 | 356 | 1831825848440 | 95.48% |
| GF | 247 | 269 | 78 | 23010009126 | 97.39% |
| GG | 142 | 268 | 97 | 112129996079 | 81.66% |
| GH | 365 | 8926 | 261 | 1232362589792 | 99.85% |
| GI | 152 | 328 | 106 | 197924942889 | 90.89% |
| GL | 103 | 144 | 167 | 43621802104 | 99.27% |
| GM | 77 | 1102 | 110 | 39586922793 | 98.99% |
| GN | 79 | 168 | 117 | 48293807144 | 97.24% |
| GP | 241 | 814 | 104 | 54408922166 | 73.07% |
| GQ | 55 | 94 | 93 | 9454961705 | 96.02% |
| GR | 1572 | 23124 | 919 | 2592595808398 | 96.51% |
| GS | 4 | 2 | 39 | 406904852 | 58.50% |
| GT | 765 | 2901 | 278 | 327769491056 | 99.41% |
| GU | 168 | 999 | 129 | 34756957255 | 99.56% |
| GW | 54 | 42 | 95 | 9320611883 | 96.97% |
| GY | 59 | 288 | 108 | 35023876139 | 96.09% |
| HK | 13857 | 94662 | 8037 | 5086145518432 | 86.09% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 885 | 1993 | 268 | 396220642928 | 94.39% |
| HR | 890 | 9691 | 465 | 1591433992322 | 96.38% |
| HT | 88 | 624 | 133 | 26468807230 | 93.34% |
| HU | 2132 | 23385 | 1446 | 3217142580588 | 98.33% |
| ID | 6960 | 81229 | 5268 | 3830718849051 | 98.13% |
| IE | 8428 | 71385 | 10538 | 5029099464792 | 88.31% |
| IL | 2519 | 36326 | 1409 | 4297150174417 | 97.61% |
| IM | 365 | 476 | 203 | 361490617307 | 96.14% |
| IN | 13541 | 197449 | 14478 | 28996126546853 | 97.93% |
| IO | 29 | 14 | 82 | 4755682343 | 95.94% |
| IQ | 704 | 3348 | 867 | 3415568560951 | 96.90% |
| IR | 1984 | 42738 | 2209 | 17026663058312 | 98.08% |
| IS | 342 | 3642 | 504 | 1562770866261 | 97.32% |
| IT | 10400 | 227582 | 8614 | 47042526101973 | 99.07% |
| JE | 211 | 636 | 117 | 219604444460 | 95.45% |
| JM | 278 | 1130 | 151 | 1122143762003 | 99.96% |
| JO | 451 | 2702 | 255 | 1121936404828 | 97.90% |
| JP | 12887 | 782233 | 12846 | 53267655011509 | 88.21% |
| KE | 1133 | 25124 | 663 | 700903631031 | 98.88% |
| KG | 241 | 1190 | 217 | 529225266790 | 99.38% |
| KH | 558 | 2012 | 299 | 430675911271 | 81.58% |
| KI | 78 | 36 | 137 | 13520203849 | 97.88% |
| KM | 69 | 52 | 81 | 9320521764 | 96.86% |
| KN | 91 | 103 | 124 | 14354937142 | 66.82% |
| KP | 9 | 11 | 440 | 501678378 | 54.76% |
| KR | 4112 | 454231 | 3577 | 22707880393555 | 99.89% |
| KW | 443 | 7582 | 219 | 877742066619 | 99.72% |
| KY | 134 | 448 | 108 | 17861638449 | 99.08% |
| KZ | 1399 | 13994 | 1178 | 3185001074457 | 86.53% |
| LA | 149 | 402 | 150 | 121260925230 | 88.17% |
| LB | 396 | 2387 | 285 | 3016027284288 | 97.70% |
| LC | 101 | 168 | 93 | 78392390948 | 97.53% |
| LI | 291 | 461 | 203 | 429948404778 | 91.87% |
| LK | 254 | 2326 | 262 | 82523791731 | 94.32% |
| LR | 84 | 440 | 108 | 27275920682 | 97.70% |
| LS | 69 | 507 | 94 | 26802794536 | 98.25% |
| LT | 1477 | 11979 | 1000 | 4281696717792 | 84.30% |
| LU | 1595 | 6711 | 788 | 2222694086776 | 96.44% |
| LV | 1262 | 8173 | 771 | 3239768866761 | 65.60% |
| LY | 155 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 353 | 48137 | 258 | 1171100138143 | 99.77% |
| MC | 285 | 286 | 108 | 142166650414 | 99.88% |
| MD | 752 | 4970 | 974 | 5953032888680 | 63.94% |
| ME | 148 | 822 | 132 | 348120630073 | 99.48% |
| MF | 61 | 105 | 86 | 5442237484 | 92.31% |
| MG | 89 | 2308 | 111 | 22627316275 | 98.13% |
| MH | 67 | 82 | 78 | 4806156329 | 95.23% |
| MK | 258 | 2801 | 185 | 1138602865493 | 97.19% |
| ML | 85 | 388 | 88 | 22658490413 | 97.55% |
| MM | 462 | 1170 | 338 | 353155080573 | 97.86% |
| MN | 239 | 737 | 276 | 133921484386 | 99.23% |
| MO | 112 | 1500 | 722 | 39117611329 | 98.24% |
| MP | 66 | 13 | 77 | 4805484589 | 95.20% |
| MQ | 298 | 796 | 72 | 104815604772 | 89.69% |
| MR | 77 | 234 | 97 | 26517696554 | 98.88% |
| MS | 33 | 15 | 77 | 644674860 | 51.26% |
| MT | 352 | 2675 | 230 | 508545969975 | 92.95% |
| MU | 2052 | 23916 | 8230 | 4455158122901833 | 100.00% |
| MV | 120 | 403 | 136 | 18180642868 | 97.49% |
| MW | 125 | 2200 | 145 | 1242315518000 | 99.85% |
| MX | 4613 | 122526 | 2675 | 3651391111850 | 96.66% |
| MY | 2797 | 34634 | 1618 | 911924446758 | 96.25% |
| MZ | 194 | 1912 | 132 | 69652197436 | 97.38% |
| NA | 288 | 1860 | 183 | 29152260397 | 99.40% |
| NC | 79 | 571 | 110 | 47910416693 | 99.48% |
| NE | 64 | 218 | 90 | 22105940004 | 98.76% |
| NF | 37 | 3 | 85 | 357560364 | 59.06% |
| NG | 1149 | 12930 | 724 | 604382242937 | 98.31% |
| NI | 172 | 1736 | 119 | 95171048766 | 98.22% |
| NL | 23881 | 214551 | 484106 | 8406901143421050 | 99.94% |
| NO | 3537 | 64024 | 2464 | 10986021029635 | 97.32% |
| NP | 471 | 2391 | 341 | 430212137059 | 94.05% |
| NR | 43 | 53 | 90 | 8951231533 | 98.41% |
| NU | 28 | 11 | 73 | 4654560295 | 81.58% |
| NZ | 2304 | 28845 | 1964 | 1360220964352 | 94.94% |
| OM | 325 | 4418 | 200 | 595055534166 | 96.04% |
| PA | 574 | 7097 | 479 | 427213685041 | 93.35% |
| PE | 1326 | 14781 | 1076 | 1081078013256 | 96.20% |
| PF | 94 | 312 | 144 | 21875061850 | 99.30% |
| PG | 234 | 348 | 153 | 73449136460 | 87.10% |
| PH | 2281 | 26767 | 1436 | 1054992279814 | 97.48% |
| PK | 1474 | 23625 | 747 | 907295911795 | 98.17% |
| PL | 8886 | 83164 | 5370 | 28643195717412 | 97.33% |
| PM | 43 | 25 | 79 | 4671532331 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.41% |
| PR | 714 | 4580 | 366 | 230686933655 | 98.78% |
| PS | 301 | 3291 | 174 | 979407798603 | 93.84% |
| PT | 2038 | 27059 | 1718 | 3362638749845 | 92.85% |
| PW | 49 | 43 | 83 | 9085313068 | 97.53% |
| PY | 456 | 5091 | 275 | 451733299856 | 97.52% |
| QA | 505 | 5171 | 359 | 560788384712 | 94.93% |
| RE | 361 | 2001 | 96 | 95622357035 | 96.43% |
| RO | 4138 | 28308 | 1980 | 5568744039163 | 95.99% |
| RS | 1089 | 9069 | 1300 | 1887339181748 | 97.41% |
| RU | 13865 | 180958 | 16246 | 54953446923777 | 90.07% |
| RW | 170 | 1266 | 121 | 1147559424297 | 99.91% |
| SA | 1708 | 48874 | 917 | 2815359795804 | 96.57% |
| SB | 56 | 71 | 100 | 17944015142 | 94.05% |
| SC | 825 | 2589 | 3025 | 3030203343403 | 63.60% |
| SD | 118 | 7412 | 109 | 69736915251 | 97.36% |
| SE | 18506 | 124724 | 51977 | 30393168557929 | 79.95% |
| SG | 11175 | 110283 | 53491 | 133688914296473 | 98.92% |
| SH | 6 | 2 | 130 | 576123908 | 54.78% |
| SI | 1049 | 10331 | 783 | 2079447070429 | 95.43% |
| SJ | 7 | 2 | 40 | 373158930 | 59.96% |
| SK | 1237 | 11154 | 898 | 3009700705481 | 97.04% |
| SL | 140 | 359 | 113 | 26584806447 | 97.43% |
| SM | 75 | 172 | 98 | 202275614245 | 92.49% |
| SN | 122 | 1646 | 151 | 27161199685 | 98.08% |
| SO | 77 | 224 | 160 | 48216207680 | 99.14% |
| SR | 53 | 340 | 103 | 17780171056 | 98.54% |
| SS | 95 | 134 | 108 | 31098290220 | 98.66% |
| ST | 50 | 60 | 92 | 4773835818 | 96.05% |
| SV | 346 | 2786 | 201 | 211465053787 | 97.09% |
| SX | 49 | 128 | 80 | 30575555620 | 94.55% |
| SY | 302 | 5173 | 138 | 945651290157 | 97.85% |
| SZ | 78 | 243 | 105 | 34670981173 | 96.15% |
| TC | 45 | 56 | 84 | 1451488555 | 71.10% |
| TD | 96 | 133 | 105 | 13816296492 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 73 | 1403 | 112 | 26517259558 | 98.88% |
| TH | 2334 | 43522 | 1274 | 595836660987 | 95.57% |
| TJ | 161 | 403 | 167 | 641204728378 | 98.14% |
| TK | 27 | 15 | 114 | 4789891118 | 95.34% |
| TL | 84 | 120 | 107 | 43596832812 | 92.49% |
| TM | 63 | 126 | 102 | 69081429303 | 99.79% |
| TN | 153 | 30817 | 143 | 116746889787 | 97.54% |
| TO | 48 | 55 | 83 | 17523466280 | 99.25% |
| TR | 3740 | 68055 | 3049 | 22542775914094 | 93.73% |
| TT | 169 | 2226 | 142 | 1177257588290 | 70.56% |
| TV | 42 | 46 | 101 | 12186021926 | 79.29% |
| TW | 9363 | 160369 | 3080 | 11547011417672 | 98.70% |
| TZ | 438 | 4520 | 245 | 267349566579 | 98.78% |
| UA | 4856 | 34578 | 12050 | 11799331579300 | 93.40% |
| UG | 269 | 5774 | 180 | 132374937915 | 96.35% |
| UM | 5 | 2 | 81 | 354939945 | 58.93% |
| US | 136817 | 5879906 | 525669 | 4835584439952045 | 99.76% |
| UY | 247 | 10118 | 187 | 177623867456 | 99.80% |
| UZ | 556 | 1663 | 276 | 1083134166637 | 99.45% |
| VA | 37 | 52 | 104 | 103326428458 | 99.88% |
| VC | 81 | 103 | 88 | 69500515366 | 99.67% |
| VE | 994 | 30459 | 655 | 2283578196403 | 99.32% |
| VG | 356 | 314 | 257 | 255456120881 | 92.19% |
| VI | 102 | 548 | 165 | 30615098437 | 94.73% |
| VN | 2299 | 65524 | 1779 | 1860478632804 | 99.17% |
| VU | 67 | 86 | 274 | 51951564590 | 80.96% |
| WF | 23 | 16 | 71 | 4788515881 | 95.30% |
| WS | 63 | 97 | 89 | 21837242666 | 92.81% |
| XK | 172 | 621 | 74 | 111806906358 | 60.60% |
| YE | 119 | 960 | 118 | 207096833081 | 99.81% |
| YT | 64 | 44 | 74 | 602988583 | 56.00% |
| ZA | 5218 | 116666 | 2646 | 41040390093491 | 78.32% |
| ZM | 344 | 6491 | 160 | 52779807058 | 92.33% |
| ZW | 275 | 923 | 264 | 65582730599 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
