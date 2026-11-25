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

Last run: `2026-11-25T07:02:39Z`

Input rows: **38761493** · Published blocks: **2367081**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 569628 | 0 | 0 | 0 | 11251768 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 496255 | 0 | 0 | 15185935 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 573842 | 0 | 0 | 0 | 12017274 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 500773 | 0 | 0 | 34553337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 409705 | 0 | 0 | 0 | 8930754 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 488340 | 0 | 1 | 16115010 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 409745 | 0 | 0 | 0 | 6098067 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 488077 | 0 | 1 | 18996390 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 428689 | 0 | 0 | 0 | 9342172 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 581888 | 0 | 1 | 17380420 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1089226 | 0 | 0 | 0 | 13988929 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2186664 | 0 | 0 | 52260186 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1447574 | 2421933 | 0 | 0 | 184984895 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 806645 | 1411510 | 0 | 0 | 17752656 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177521 | 69962 | 0 | 0 | 3407047 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 302944 | 322077 | 0 | 9 | 29154539 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 302944 | 322077 | 0 | 9 | 13353189 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4346 | 1723 | 0 | 0 | 98784 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3882 | 0 | 0 | 0 | 82911 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327930 | 0 | 0 | 0 | 537841 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 819260 | 436592 | 0 | 412783 | 18650742 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262798 | 71231 | 0 | 10 | 34655561 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4957269 | 0 | 0 | 0 | 66550081 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380534 | 0 | 0 | 11707152 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 803015 | 0 | 0 | 0 | 11123858 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6241 | 3447 | 0 | 0 | 154514 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1751 | 578 | 0 | 0 | 36111 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 572321 | 497070 | 0 | 0 | 3411761 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 619986 | 3702257408 | 1746845 | 293421847207842436812546030166054273024 |
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
| AD | 217 | 308 | 190 | 43728565320 | 92.57% |
| AE | 2878 | 25506 | 2798 | 12770686841820 | 90.03% |
| AF | 224 | 875 | 477 | 121136965472 | 97.42% |
| AG | 138 | 176 | 186 | 40393307195 | 65.28% |
| AI | 45 | 45 | 100 | 5847643190 | 67.40% |
| AL | 555 | 1409 | 1254 | 2508745170846 | 95.10% |
| AM | 361 | 2324 | 525 | 1293644175574 | 94.63% |
| AO | 518 | 5123 | 507 | 147599979092 | 98.68% |
| AQ | 60 | 5 | 287 | 404842625 | 44.23% |
| AR | 3878 | 77017 | 2406 | 23630989664452 | 98.87% |
| AS | 42 | 29 | 317 | 9243193524 | 96.04% |
| AT | 4792 | 45215 | 4245 | 10540281576956 | 97.36% |
| AU | 13553 | 219872 | 71425 | 2309962469356257580 | 100.00% |
| AW | 51 | 449 | 96 | 11534578740 | 92.96% |
| AX | 77 | 132 | 608 | 69552506698 | 72.57% |
| AZ | 399 | 3070 | 323 | 1963968619115 | 98.36% |
| BA | 326 | 3264 | 744 | 833827513885 | 95.13% |
| BB | 195 | 507 | 109 | 69400590027 | 99.69% |
| BD | 2670 | 8217 | 2597 | 3581889172777 | 96.45% |
| BE | 8639 | 58024 | 4859 | 4398097011451959 | 66.61% |
| BF | 126 | 1387 | 145 | 91245528373 | 99.43% |
| BG | 2354 | 16715 | 1436 | 5129179011889 | 96.12% |
| BH | 340 | 5248 | 317 | 406108709709 | 97.65% |
| BI | 60 | 168 | 109 | 26952917544 | 92.82% |
| BJ | 86 | 622 | 128 | 56582575407 | 96.84% |
| BL | 24 | 10 | 87 | 1047851055 | 57.46% |
| BM | 317 | 501 | 107 | 43426918710 | 99.37% |
| BN | 112 | 878 | 118 | 26250616878 | 99.06% |
| BO | 327 | 4753 | 306 | 168302404200 | 98.07% |
| BQ | 39 | 103 | 70 | 17538741295 | 98.90% |
| BR | 11785 | 335619 | 35862 | 37440877054429 | 99.23% |
| BS | 141 | 634 | 162 | 14003275352 | 96.66% |
| BT | 126 | 226 | 157 | 60963021390 | 99.36% |
| BV | 1 | 1 | 3 | 17825792 | 34.31% |
| BW | 204 | 723 | 123 | 82923761967 | 96.34% |
| BY | 396 | 7406 | 442 | 700163615037 | 99.49% |
| BZ | 215 | 469 | 200 | 192579188554 | 89.06% |
| CA | 19908 | 285232 | 27322 | 6442132671007 | 95.46% |
| CC | 7 | 0 | 50 | 827256852 | 42.84% |
| CD | 352 | 940 | 203 | 155963863379 | 98.74% |
| CF | 50 | 40 | 96 | 1117335591 | 59.87% |
| CG | 93 | 585 | 116 | 36988538926 | 97.17% |
| CH | 6995 | 98601 | 57261 | 15878800904407847056 | 99.98% |
| CI | 349 | 6713 | 167 | 194887800117 | 98.94% |
| CK | 36 | 41 | 92 | 9234282542 | 96.05% |
| CL | 2088 | 41389 | 2027 | 1419535897238 | 96.66% |
| CM | 192 | 2438 | 162 | 67900154955 | 98.18% |
| CN | 6947 | 1347805 | 13012 | 294844728751866 | 99.94% |
| CO | 3212 | 70319 | 1922 | 1994517111665 | 97.92% |
| CR | 598 | 9169 | 371 | 461702268525 | 97.36% |
| CU | 96 | 1023 | 128 | 26439696696 | 99.17% |
| CV | 84 | 241 | 145 | 13715370030 | 97.22% |
| CW | 140 | 718 | 103 | 63125206060 | 92.42% |
| CX | 7 | 0 | 81 | 727429155 | 41.43% |
| CY | 678 | 4128 | 711 | 6042768253794 | 93.92% |
| CZ | 3950 | 33839 | 6281 | 13303776378776 | 97.50% |
| DE | 38779 | 550750 | 82268 | 115541268747713 | 96.04% |
| DJ | 66 | 247 | 121 | 9641325614 | 94.96% |
| DK | 3666 | 50121 | 2097 | 9637217892262 | 93.45% |
| DM | 74 | 71 | 116 | 44401817134 | 73.08% |
| DO | 748 | 8098 | 632 | 941393477988 | 98.70% |
| DZ | 239 | 18724 | 243 | 47397892265 | 92.39% |
| EC | 1098 | 11267 | 661 | 1162867668870 | 98.70% |
| EE | 1206 | 6491 | 1506 | 2725768289056 | 96.02% |
| EG | 808 | 94821 | 279 | 17726577500565 | 100.00% |
| EH | 4 | 0 | 54 | 457961495 | 46.78% |
| ER | 47 | 29 | 95 | 1100366891 | 57.01% |
| ES | 11484 | 139958 | 6440 | 26202036769917 | 97.12% |
| ET | 97 | 1483 | 134 | 1117547839542 | 99.96% |
| FI | 4142 | 58808 | 33770 | 6895814559378 | 96.70% |
| FJ | 115 | 590 | 118 | 43379783467 | 99.48% |
| FK | 40 | 26 | 83 | 578678822 | 46.09% |
| FM | 59 | 49 | 107 | 8921740345 | 78.74% |
| FO | 47 | 181 | 88 | 108187546923 | 99.57% |
| FR | 28454 | 325741 | 18440 | 67589378067783 | 98.43% |
| GA | 91 | 1558 | 106 | 35779724581 | 94.29% |
| GB | 41817 | 498463 | 61168 | 105923055535775 | 93.88% |
| GD | 102 | 128 | 104 | 82705774122 | 96.88% |
| GE | 489 | 5120 | 364 | 1831825930367 | 95.38% |
| GF | 257 | 249 | 79 | 23010009127 | 97.39% |
| GG | 133 | 266 | 96 | 112129799470 | 81.66% |
| GH | 361 | 8925 | 264 | 1232362655330 | 99.85% |
| GI | 155 | 329 | 107 | 197924942890 | 90.89% |
| GL | 104 | 144 | 165 | 43621736567 | 99.27% |
| GM | 77 | 1102 | 110 | 39586922793 | 98.99% |
| GN | 78 | 168 | 117 | 48293807144 | 97.24% |
| GP | 225 | 816 | 104 | 54408922166 | 73.06% |
| GQ | 55 | 94 | 94 | 9454961706 | 96.02% |
| GR | 1551 | 23125 | 885 | 2592595349640 | 96.51% |
| GS | 4 | 2 | 40 | 406904853 | 59.18% |
| GT | 670 | 2898 | 281 | 327769491059 | 99.41% |
| GU | 182 | 995 | 129 | 34757681467 | 99.56% |
| GW | 54 | 42 | 95 | 9320611883 | 96.97% |
| GY | 59 | 288 | 109 | 35023941675 | 96.09% |
| HK | 13964 | 95126 | 8298 | 5099970307938 | 85.92% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 848 | 1994 | 268 | 391925675633 | 94.76% |
| HR | 882 | 9691 | 435 | 1591451098328 | 96.38% |
| HT | 88 | 624 | 134 | 26468807231 | 93.34% |
| HU | 2084 | 23382 | 1551 | 3217144284785 | 98.33% |
| ID | 6912 | 81266 | 5292 | 3827496775791 | 98.19% |
| IE | 8331 | 72146 | 8616 | 4986552097330 | 88.66% |
| IL | 2532 | 37073 | 1575 | 4297793000526 | 97.42% |
| IM | 362 | 476 | 192 | 361490691644 | 96.14% |
| IN | 13523 | 197414 | 14184 | 28991808176340 | 97.82% |
| IO | 29 | 14 | 83 | 4755682344 | 95.94% |
| IQ | 693 | 3347 | 845 | 3415585403677 | 96.90% |
| IR | 1980 | 42741 | 2211 | 17060955966344 | 97.97% |
| IS | 342 | 3642 | 486 | 1562905254769 | 97.32% |
| IT | 10317 | 227808 | 8580 | 47055385606186 | 99.06% |
| JE | 220 | 638 | 117 | 219604759852 | 95.45% |
| JM | 255 | 1131 | 156 | 1122143762009 | 99.96% |
| JO | 451 | 2702 | 256 | 1121936404829 | 97.90% |
| JP | 12889 | 782644 | 13684 | 53263163206539 | 88.22% |
| KE | 1190 | 25133 | 667 | 696608794810 | 99.07% |
| KG | 241 | 1186 | 212 | 529225201250 | 99.38% |
| KH | 555 | 2011 | 298 | 430675911270 | 81.58% |
| KI | 80 | 36 | 137 | 13520203849 | 97.88% |
| KM | 68 | 52 | 81 | 9320521764 | 96.86% |
| KN | 95 | 106 | 123 | 14354937141 | 66.82% |
| KP | 9 | 11 | 419 | 499269931 | 55.27% |
| KR | 4114 | 455050 | 3544 | 22712175780713 | 99.87% |
| KW | 442 | 7581 | 219 | 877742051777 | 99.72% |
| KY | 131 | 450 | 108 | 17861638449 | 99.08% |
| KZ | 1383 | 13994 | 1248 | 3150641139280 | 86.75% |
| LA | 148 | 418 | 150 | 121260925230 | 88.17% |
| LB | 395 | 2384 | 284 | 3016027284287 | 97.70% |
| LC | 106 | 168 | 93 | 78392390948 | 97.53% |
| LI | 286 | 451 | 200 | 429947683877 | 93.12% |
| LK | 259 | 2325 | 261 | 82523791730 | 94.32% |
| LR | 85 | 440 | 107 | 27275920681 | 97.70% |
| LS | 69 | 507 | 93 | 26802794535 | 98.25% |
| LT | 1439 | 11949 | 970 | 4177006845719 | 83.73% |
| LU | 1582 | 6712 | 878 | 2222695023746 | 96.44% |
| LV | 1249 | 8165 | 789 | 3235473567974 | 65.64% |
| LY | 153 | 1923 | 107 | 99614748732 | 89.61% |
| MA | 355 | 48138 | 261 | 1171100465823 | 99.78% |
| MC | 280 | 281 | 108 | 142166650414 | 99.88% |
| MD | 754 | 4981 | 976 | 5953032877391 | 64.00% |
| ME | 148 | 822 | 130 | 348120040249 | 99.48% |
| MF | 57 | 105 | 90 | 5443335213 | 92.30% |
| MG | 87 | 2308 | 112 | 22627381811 | 98.13% |
| MH | 75 | 83 | 79 | 4806418473 | 95.22% |
| MK | 258 | 2800 | 187 | 1138602865495 | 97.19% |
| ML | 84 | 388 | 87 | 22658490412 | 97.55% |
| MM | 462 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 239 | 737 | 275 | 133921418850 | 99.23% |
| MO | 105 | 1500 | 737 | 39116611919 | 99.15% |
| MP | 84 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 305 | 811 | 72 | 104815670307 | 89.69% |
| MR | 77 | 234 | 97 | 26517696554 | 98.88% |
| MS | 33 | 15 | 79 | 644674607 | 51.26% |
| MT | 326 | 2678 | 232 | 508545920828 | 92.95% |
| MU | 2015 | 23894 | 8201 | 4455162418082116 | 100.00% |
| MV | 120 | 403 | 131 | 18180677684 | 97.49% |
| MW | 125 | 2200 | 145 | 1242315518000 | 99.85% |
| MX | 4602 | 122525 | 2552 | 3651392617520 | 96.87% |
| MY | 2800 | 34642 | 1633 | 911924676181 | 96.08% |
| MZ | 191 | 1914 | 132 | 69652197436 | 97.38% |
| NA | 302 | 1861 | 184 | 29152325933 | 97.58% |
| NC | 79 | 571 | 109 | 47910351157 | 99.48% |
| NE | 64 | 218 | 90 | 22105940004 | 98.76% |
| NF | 37 | 3 | 85 | 357560364 | 59.06% |
| NG | 1122 | 12923 | 710 | 604382767233 | 98.31% |
| NI | 172 | 1736 | 123 | 95171114308 | 98.22% |
| NL | 24137 | 214571 | 487207 | 8406949536042032 | 99.94% |
| NO | 3430 | 63964 | 2428 | 10985932081719 | 97.20% |
| NP | 471 | 2393 | 341 | 430212202594 | 94.05% |
| NR | 43 | 53 | 90 | 8951231533 | 98.41% |
| NU | 28 | 11 | 73 | 4654560295 | 81.58% |
| NZ | 2297 | 28862 | 1843 | 1360223697409 | 94.94% |
| OM | 324 | 4418 | 203 | 595055599704 | 96.04% |
| PA | 558 | 7095 | 464 | 427213619508 | 93.34% |
| PE | 1302 | 14775 | 1072 | 1081005399557 | 96.21% |
| PF | 94 | 312 | 144 | 21875061850 | 99.30% |
| PG | 235 | 349 | 152 | 73449136459 | 87.10% |
| PH | 2288 | 26819 | 1281 | 1050699737354 | 97.67% |
| PK | 1481 | 23640 | 742 | 907296026853 | 97.95% |
| PL | 8929 | 83164 | 5240 | 28613124484023 | 96.73% |
| PM | 43 | 25 | 79 | 4671532331 | 96.85% |
| PN | 6 | 3 | 45 | 355464214 | 60.40% |
| PR | 649 | 4572 | 339 | 230687435145 | 98.76% |
| PS | 301 | 3298 | 176 | 979424837963 | 93.84% |
| PT | 1972 | 27063 | 1845 | 3362638192524 | 92.83% |
| PW | 49 | 43 | 84 | 9085575212 | 97.53% |
| PY | 456 | 5092 | 275 | 451733299856 | 97.52% |
| QA | 526 | 5197 | 360 | 560788496330 | 94.93% |
| RE | 314 | 1998 | 95 | 95622373422 | 96.43% |
| RO | 4128 | 28294 | 2173 | 5572233615336 | 95.57% |
| RS | 1078 | 9068 | 1311 | 1887339792334 | 97.41% |
| RU | 13780 | 180974 | 16483 | 54739356967187 | 90.03% |
| RW | 174 | 1266 | 122 | 1147559424298 | 99.91% |
| SA | 1702 | 48872 | 924 | 2815443747391 | 96.57% |
| SB | 56 | 71 | 100 | 17944015142 | 90.06% |
| SC | 836 | 2547 | 3264 | 3288165884457 | 62.19% |
| SD | 118 | 7412 | 110 | 69736980787 | 97.36% |
| SE | 17793 | 126531 | 53272 | 30393335795089 | 79.94% |
| SG | 10872 | 110847 | 53106 | 133710348513861 | 98.91% |
| SH | 6 | 2 | 130 | 575140868 | 54.83% |
| SI | 1042 | 10330 | 622 | 2079447594940 | 95.35% |
| SJ | 7 | 2 | 42 | 374469650 | 60.67% |
| SK | 1231 | 11154 | 949 | 3009701131543 | 96.94% |
| SL | 135 | 359 | 112 | 26584806446 | 96.08% |
| SM | 74 | 172 | 98 | 202275614245 | 92.49% |
| SN | 122 | 1646 | 150 | 27161199684 | 98.08% |
| SO | 76 | 224 | 160 | 48216207680 | 99.14% |
| SR | 53 | 340 | 104 | 17780171057 | 98.54% |
| SS | 103 | 135 | 108 | 31098290220 | 98.66% |
| ST | 50 | 60 | 92 | 4773835818 | 96.05% |
| SV | 337 | 2786 | 209 | 211465054055 | 97.09% |
| SX | 65 | 128 | 82 | 30575621157 | 94.55% |
| SY | 297 | 5171 | 138 | 945651290157 | 97.85% |
| SZ | 78 | 243 | 104 | 34670981172 | 96.15% |
| TC | 45 | 56 | 84 | 1451586855 | 71.10% |
| TD | 94 | 133 | 105 | 13816296492 | 97.12% |
| TF | 2 | 0 | 34 | 305455122 | 59.87% |
| TG | 73 | 1403 | 112 | 26517259558 | 98.88% |
| TH | 2284 | 43568 | 1258 | 595753037375 | 95.52% |
| TJ | 162 | 403 | 166 | 641204728377 | 98.14% |
| TK | 27 | 15 | 114 | 4789891118 | 95.34% |
| TL | 85 | 121 | 107 | 43596832812 | 92.49% |
| TM | 64 | 126 | 102 | 69081429303 | 99.79% |
| TN | 153 | 30817 | 144 | 116746955323 | 97.54% |
| TO | 48 | 55 | 83 | 17523466280 | 99.25% |
| TR | 3742 | 68059 | 2914 | 22542777191829 | 93.64% |
| TT | 169 | 2226 | 146 | 1177257588298 | 70.56% |
| TV | 42 | 46 | 98 | 11112280102 | 82.95% |
| TW | 9318 | 160436 | 3154 | 11549442310724 | 98.68% |
| TZ | 430 | 4493 | 247 | 267349632116 | 98.78% |
| UA | 4921 | 34590 | 3307 | 11756587049445 | 93.17% |
| UG | 266 | 5773 | 180 | 132375003450 | 96.35% |
| UM | 5 | 2 | 84 | 355005483 | 58.92% |
| US | 137047 | 5870397 | 536977 | 4835805630571952 | 99.76% |
| UY | 246 | 10118 | 191 | 177623867457 | 99.80% |
| UZ | 553 | 1661 | 272 | 1083134183018 | 99.45% |
| VA | 36 | 51 | 103 | 103326428457 | 99.88% |
| VC | 82 | 103 | 90 | 69500515368 | 99.67% |
| VE | 988 | 30443 | 654 | 2283561746865 | 99.33% |
| VG | 353 | 312 | 260 | 255456170034 | 89.93% |
| VI | 101 | 548 | 170 | 30615360582 | 94.73% |
| VN | 2289 | 65517 | 1808 | 1860462236699 | 99.17% |
| VU | 67 | 86 | 274 | 51951564590 | 80.96% |
| WF | 23 | 16 | 71 | 4788515881 | 95.30% |
| WS | 63 | 97 | 89 | 21837242666 | 92.81% |
| XK | 205 | 627 | 72 | 111806775286 | 60.60% |
| YE | 119 | 960 | 118 | 207096833081 | 99.81% |
| YT | 71 | 43 | 74 | 602988583 | 56.00% |
| ZA | 5196 | 117567 | 2694 | 41040391955563 | 78.31% |
| ZM | 375 | 6511 | 161 | 48484905299 | 96.08% |
| ZW | 270 | 928 | 302 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
