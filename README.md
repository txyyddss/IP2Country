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

Last run: `2026-10-24T03:20:57Z`

Input rows: **38902209** · Published blocks: **2325900**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 597286 | 0 | 0 | 0 | 11032548 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 785668 | 0 | 0 | 20177287 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 597286 | 0 | 0 | 0 | 7580021 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 785668 | 0 | 0 | 24060852 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562273 | 0 | 0 | 0 | 11078750 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 499547 | 0 | 0 | 14984389 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564252 | 0 | 0 | 532 | 11826650 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 499624 | 0 | 153 | 34484307 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 399731 | 0 | 0 | 0 | 8712628 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 471910 | 0 | 1 | 15539292 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 399513 | 0 | 0 | 0 | 5946795 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 471895 | 0 | 1 | 18361866 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 416844 | 0 | 0 | 0 | 9093016 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 563816 | 0 | 1 | 16753949 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 722011 | 0 | 0 | 2137 | 9618305 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288565 | 0 | 3995 | 8458984 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321994 | 0 | 0 | 0 | 7039282 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 87831 | 0 | 0 | 3813168 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 978654 | 0 | 0 | 0 | 12966588 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2708038 | 0 | 0 | 85202442 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1337977 | 2937794 | 0 | 0 | 208656999 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 786311 | 1315385 | 0 | 0 | 16820984 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177161 | 69725 | 0 | 0 | 3395371 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 289856 | 492472 | 0 | 97800 | 41026977 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 289856 | 492472 | 0 | 98389 | 18562285 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4317 | 1673 | 0 | 0 | 97542 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3881 | 0 | 0 | 0 | 82395 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327613 | 0 | 0 | 0 | 537556 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 808893 | 422767 | 0 | 411267 | 18234970 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262318 | 70985 | 0 | 10 | 34579854 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802825 | 0 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6526 | 3489 | 0 | 0 | 159325 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 574 | 0 | 0 | 36004 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 565547 | 500485 | 0 | 0 | 3406491 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 600896 | 3702257408 | 1724754 | 293421847207842436812546030166054273024 |
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
| AD | 212 | 307 | 162 | 43780996399 | 91.72% |
| AE | 2849 | 25472 | 2400 | 12826428474011 | 89.83% |
| AF | 232 | 878 | 314 | 121052755649 | 97.45% |
| AG | 120 | 177 | 172 | 40375548207 | 65.28% |
| AI | 44 | 46 | 80 | 5830802724 | 67.48% |
| AL | 561 | 1422 | 1273 | 2500709001837 | 95.22% |
| AM | 334 | 2307 | 414 | 1259266648420 | 94.71% |
| AO | 552 | 5126 | 278 | 143305134668 | 98.64% |
| AQ | 63 | 6 | 171 | 467798310 | 43.57% |
| AR | 3742 | 76924 | 2808 | 23596365820716 | 98.84% |
| AS | 39 | 28 | 202 | 9240951108 | 96.06% |
| AT | 4676 | 45106 | 4506 | 10476626624650 | 97.51% |
| AU | 12997 | 219011 | 8005 | 2305920540254768361 | 100.00% |
| AW | 49 | 449 | 84 | 11516754218 | 93.04% |
| AX | 100 | 131 | 585 | 69601594420 | 74.62% |
| AZ | 398 | 3068 | 301 | 1934086992724 | 99.30% |
| BA | 341 | 3264 | 734 | 803811947281 | 94.68% |
| BB | 190 | 508 | 98 | 69382748968 | 99.71% |
| BD | 2636 | 8174 | 2661 | 3577608746435 | 96.42% |
| BE | 8521 | 58087 | 73177 | 8440506885712926 | 66.66% |
| BF | 121 | 1381 | 137 | 91261192750 | 99.42% |
| BG | 2370 | 16721 | 1393 | 5172161964033 | 95.72% |
| BH | 327 | 5245 | 305 | 406107403329 | 97.65% |
| BI | 58 | 167 | 106 | 26902524710 | 92.88% |
| BJ | 83 | 620 | 109 | 56564816429 | 96.86% |
| BL | 42 | 10 | 76 | 1029962023 | 57.76% |
| BM | 316 | 498 | 109 | 43678792242 | 99.00% |
| BN | 109 | 878 | 111 | 26232664360 | 95.82% |
| BO | 336 | 4758 | 294 | 168267739998 | 97.23% |
| BQ | 36 | 103 | 67 | 17538021676 | 98.90% |
| BR | 11494 | 336097 | 42641 | 37419318504633 | 99.16% |
| BS | 135 | 632 | 148 | 13972409160 | 96.79% |
| BT | 124 | 226 | 152 | 60945265220 | 99.38% |
| BV | 4 | 1 | 3 | 1048577 | 50.00% |
| BW | 182 | 717 | 120 | 78595244588 | 97.52% |
| BY | 409 | 7402 | 334 | 700178953434 | 99.59% |
| BZ | 210 | 475 | 187 | 192696613689 | 89.03% |
| CA | 19401 | 285272 | 30771 | 6346741659917 | 95.55% |
| CC | 6 | 0 | 44 | 842725650 | 43.17% |
| CD | 313 | 932 | 197 | 155930247756 | 98.74% |
| CF | 49 | 40 | 91 | 1050231075 | 60.02% |
| CG | 89 | 585 | 109 | 36970586407 | 97.19% |
| CH | 6922 | 98331 | 57102 | 15878800913693744497 | 99.98% |
| CI | 348 | 6837 | 153 | 194903530024 | 98.93% |
| CK | 33 | 41 | 78 | 9217637672 | 96.15% |
| CL | 2023 | 41393 | 2360 | 1415225228815 | 96.62% |
| CM | 191 | 2438 | 154 | 67915819077 | 98.17% |
| CN | 7017 | 1348059 | 10808 | 294817154178366 | 99.94% |
| CO | 3139 | 70218 | 1879 | 1977320047690 | 97.73% |
| CR | 597 | 9171 | 380 | 461687065703 | 97.36% |
| CU | 99 | 1022 | 114 | 26422592046 | 99.21% |
| CV | 81 | 240 | 133 | 13663861034 | 97.38% |
| CW | 140 | 720 | 90 | 62973098276 | 92.55% |
| CX | 7 | 0 | 57 | 726448403 | 41.65% |
| CY | 687 | 4147 | 671 | 5999830445131 | 93.92% |
| CZ | 3768 | 33866 | 6909 | 13290450264388 | 97.58% |
| DE | 37891 | 551103 | 79456 | 115188760523383 | 96.07% |
| DJ | 70 | 243 | 110 | 9808836647 | 94.09% |
| DK | 3711 | 50109 | 1996 | 9667298694473 | 93.46% |
| DM | 77 | 71 | 110 | 44384058154 | 73.09% |
| DO | 740 | 8064 | 638 | 928490784366 | 98.99% |
| DZ | 226 | 18725 | 219 | 44143245715 | 93.89% |
| EC | 1122 | 11253 | 643 | 1158522015826 | 98.69% |
| EE | 1176 | 6516 | 1676 | 2730082613690 | 95.97% |
| EG | 790 | 94821 | 266 | 17730956161666 | 99.99% |
| EH | 4 | 0 | 45 | 456849683 | 46.77% |
| ER | 46 | 29 | 89 | 998656293 | 59.40% |
| ES | 11192 | 139785 | 6650 | 26207440788315 | 97.13% |
| ET | 95 | 1478 | 130 | 1117513238835 | 99.97% |
| FI | 4114 | 58635 | 27839 | 6857211958547 | 96.71% |
| FJ | 110 | 594 | 115 | 43363073064 | 99.51% |
| FK | 38 | 26 | 76 | 544078114 | 44.62% |
| FM | 56 | 49 | 98 | 8953002292 | 88.23% |
| FO | 48 | 181 | 88 | 108270320169 | 99.54% |
| FR | 28120 | 325491 | 20191 | 67346430700977 | 98.51% |
| GA | 89 | 1559 | 93 | 35744925983 | 94.32% |
| GB | 40653 | 499002 | 57340 | 106033846706647 | 94.09% |
| GD | 100 | 128 | 96 | 82687883815 | 96.90% |
| GE | 489 | 5108 | 321 | 1797448223071 | 95.68% |
| GF | 271 | 259 | 80 | 23008899366 | 97.40% |
| GG | 146 | 261 | 86 | 146471778853 | 85.47% |
| GH | 373 | 8927 | 257 | 1232395880282 | 99.79% |
| GI | 149 | 327 | 108 | 198008895782 | 91.23% |
| GL | 108 | 145 | 157 | 43603849584 | 99.30% |
| GM | 76 | 1102 | 108 | 39552453159 | 99.04% |
| GN | 75 | 172 | 109 | 48242362660 | 97.28% |
| GP | 230 | 795 | 98 | 54391032115 | 73.08% |
| GQ | 56 | 95 | 90 | 9487272231 | 95.90% |
| GR | 1529 | 23125 | 970 | 2592006046721 | 96.46% |
| GS | 4 | 2 | 39 | 389148947 | 59.18% |
| GT | 670 | 2887 | 275 | 327885867886 | 99.39% |
| GU | 158 | 990 | 116 | 34908664372 | 99.32% |
| GW | 50 | 42 | 90 | 9319434535 | 96.98% |
| GY | 57 | 288 | 104 | 35006052649 | 99.18% |
| HK | 13604 | 94998 | 7466 | 5142575726030 | 85.95% |
| HM | 1 | 2 | 7 | 17965056 | 49.38% |
| HN | 831 | 1992 | 250 | 391906802547 | 94.76% |
| HR | 865 | 9688 | 427 | 1591501320707 | 96.38% |
| HT | 89 | 624 | 120 | 26450983730 | 93.37% |
| HU | 1942 | 23373 | 1298 | 3217193271680 | 98.25% |
| ID | 7054 | 85406 | 4665 | 3792027746136 | 98.45% |
| IE | 8322 | 72150 | 8088 | 4921993985719 | 89.13% |
| IL | 2428 | 39290 | 1477 | 4302329931905 | 97.44% |
| IM | 345 | 474 | 174 | 361538861867 | 96.14% |
| IN | 13408 | 197171 | 17724 | 28987603755270 | 97.91% |
| IO | 33 | 15 | 76 | 4737988898 | 96.08% |
| IQ | 681 | 3342 | 631 | 3415550477127 | 96.90% |
| IR | 1963 | 42727 | 2211 | 16859158416487 | 98.53% |
| IS | 339 | 3642 | 470 | 1605333718111 | 96.23% |
| IT | 10011 | 227836 | 8986 | 46978083107006 | 99.09% |
| JE | 224 | 643 | 111 | 219655158307 | 95.45% |
| JM | 254 | 1132 | 144 | 1122124888911 | 99.96% |
| JO | 445 | 2705 | 240 | 1121935423825 | 97.90% |
| JP | 12604 | 782359 | 14855 | 52720665353468 | 88.45% |
| KE | 1090 | 25116 | 639 | 671057099184 | 98.83% |
| KG | 234 | 1179 | 190 | 529224022862 | 99.25% |
| KH | 559 | 2011 | 277 | 434970817357 | 81.10% |
| KI | 78 | 35 | 122 | 13519026501 | 92.47% |
| KM | 61 | 51 | 85 | 9353097507 | 96.74% |
| KN | 116 | 164 | 94 | 5746850355 | 77.11% |
| KP | 9 | 11 | 313 | 490570269 | 55.54% |
| KR | 3964 | 455209 | 3535 | 22750793070178 | 99.65% |
| KW | 418 | 7567 | 208 | 876851815098 | 99.77% |
| KY | 126 | 439 | 94 | 17845583401 | 99.15% |
| KZ | 1328 | 13962 | 974 | 3150623753541 | 86.62% |
| LA | 150 | 452 | 145 | 121244086823 | 88.18% |
| LB | 391 | 2388 | 276 | 3015891891256 | 98.08% |
| LC | 89 | 167 | 86 | 78374500897 | 97.54% |
| LI | 327 | 451 | 203 | 429998673192 | 93.40% |
| LK | 293 | 2326 | 228 | 82539459136 | 92.58% |
| LR | 79 | 437 | 101 | 27241187880 | 97.76% |
| LS | 71 | 507 | 86 | 26751287588 | 98.32% |
| LT | 1490 | 12217 | 899 | 4368449689747 | 85.05% |
| LU | 1551 | 6715 | 856 | 2188410848866 | 96.22% |
| LV | 1236 | 8162 | 697 | 3235808980373 | 65.68% |
| LY | 151 | 1923 | 105 | 99581264185 | 89.62% |
| MA | 346 | 48129 | 251 | 1171116196756 | 99.78% |
| MC | 282 | 284 | 108 | 142217116456 | 99.87% |
| MD | 773 | 4958 | 952 | 5931591680571 | 64.00% |
| ME | 150 | 824 | 129 | 348170440753 | 99.48% |
| MF | 57 | 103 | 79 | 5359122725 | 92.88% |
| MG | 88 | 2308 | 110 | 22676538158 | 98.05% |
| MH | 71 | 90 | 76 | 4789645605 | 95.44% |
| MK | 251 | 2798 | 189 | 1142897918059 | 96.63% |
| ML | 85 | 388 | 87 | 22640603435 | 97.58% |
| MM | 461 | 1167 | 320 | 353154101865 | 97.56% |
| MN | 242 | 739 | 247 | 133921255753 | 99.24% |
| MO | 107 | 1500 | 470 | 39131579201 | 99.13% |
| MP | 76 | 18 | 71 | 4771082536 | 95.56% |
| MQ | 307 | 791 | 69 | 104798763298 | 89.70% |
| MR | 70 | 232 | 92 | 26483030311 | 98.94% |
| MS | 30 | 15 | 75 | 626850349 | 51.57% |
| MT | 320 | 2678 | 218 | 501360819246 | 93.67% |
| MU | 2033 | 23887 | 8141 | 4455217983291954 | 100.00% |
| MV | 118 | 401 | 128 | 18197508394 | 97.47% |
| MW | 124 | 2200 | 140 | 1242314342699 | 99.85% |
| MX | 4431 | 122269 | 2574 | 3621259010206 | 96.79% |
| MY | 2865 | 34674 | 1336 | 911705199289 | 95.78% |
| MZ | 196 | 1911 | 131 | 69652201784 | 97.38% |
| NA | 359 | 1856 | 182 | 29151347241 | 99.39% |
| NC | 74 | 571 | 102 | 47875881518 | 99.52% |
| NE | 61 | 217 | 92 | 22121672996 | 98.71% |
| NF | 33 | 2 | 74 | 358742312 | 59.47% |
| NG | 1113 | 12938 | 846 | 604634397021 | 98.29% |
| NI | 176 | 1731 | 120 | 95154338115 | 98.22% |
| NL | 23925 | 215104 | 501385 | 8408983993214750 | 99.94% |
| NO | 3225 | 63856 | 2764 | 10943233034235 | 97.37% |
| NP | 472 | 2389 | 339 | 434489217375 | 93.37% |
| NR | 40 | 53 | 79 | 8917678373 | 98.61% |
| NU | 27 | 11 | 74 | 4671469862 | 96.82% |
| NZ | 2285 | 29284 | 1717 | 1359949185486 | 95.35% |
| OM | 304 | 4418 | 183 | 595054620993 | 96.04% |
| PA | 536 | 7101 | 460 | 429343311405 | 92.60% |
| PE | 1286 | 14892 | 1066 | 1072377655297 | 96.46% |
| PF | 91 | 312 | 133 | 21857171793 | 99.35% |
| PG | 263 | 349 | 144 | 73449140800 | 91.78% |
| PH | 2246 | 26770 | 1233 | 1046351426025 | 97.62% |
| PK | 1425 | 23634 | 772 | 907362025941 | 97.65% |
| PL | 8633 | 83057 | 5255 | 28443874825330 | 97.03% |
| PM | 44 | 25 | 70 | 4670356004 | 96.79% |
| PN | 5 | 2 | 38 | 355531027 | 60.85% |
| PR | 665 | 4576 | 330 | 230653880960 | 99.07% |
| PS | 308 | 3298 | 164 | 979424774722 | 93.84% |
| PT | 1898 | 27023 | 2081 | 3328571444923 | 92.98% |
| PW | 48 | 43 | 75 | 9084268838 | 97.54% |
| PY | 444 | 5087 | 261 | 451683300207 | 97.34% |
| QA | 476 | 5173 | 363 | 526342861010 | 96.21% |
| RE | 351 | 1988 | 91 | 95638106407 | 96.44% |
| RO | 4048 | 28314 | 2457 | 5569648758511 | 95.60% |
| RS | 1052 | 9072 | 1373 | 1853001230987 | 98.28% |
| RU | 13636 | 180633 | 12475 | 54493278544581 | 90.28% |
| RW | 212 | 1265 | 120 | 1147541537319 | 99.91% |
| SA | 1653 | 46657 | 907 | 2815426323898 | 96.13% |
| SB | 54 | 71 | 86 | 17657887269 | 94.77% |
| SC | 1036 | 2749 | 3344 | 3336783000342 | 61.78% |
| SD | 115 | 7412 | 104 | 69686719021 | 97.39% |
| SE | 16093 | 126285 | 58111 | 30392289950307 | 79.88% |
| SG | 10736 | 107647 | 52576 | 151211297831883 | 97.13% |
| SH | 6 | 2 | 127 | 590543106 | 54.52% |
| SI | 1085 | 10325 | 653 | 2113819469482 | 95.56% |
| SJ | 7 | 2 | 40 | 390134034 | 60.68% |
| SK | 1223 | 11163 | 773 | 2980507950534 | 97.23% |
| SL | 126 | 356 | 105 | 26701068586 | 95.84% |
| SM | 71 | 172 | 99 | 202294426403 | 92.49% |
| SN | 121 | 1644 | 145 | 27193641279 | 98.05% |
| SO | 74 | 226 | 141 | 48199434795 | 99.16% |
| SR | 51 | 341 | 103 | 17812484653 | 98.46% |
| SS | 101 | 135 | 105 | 31080403241 | 98.68% |
| ST | 47 | 60 | 86 | 4772854053 | 96.06% |
| SV | 344 | 2774 | 204 | 211430563802 | 97.10% |
| SX | 44 | 129 | 75 | 30423513379 | 94.82% |
| SY | 295 | 5167 | 138 | 842622210350 | 99.62% |
| SZ | 76 | 242 | 100 | 30358044977 | 99.16% |
| TC | 45 | 56 | 77 | 1569880611 | 68.81% |
| TD | 91 | 132 | 98 | 13816166695 | 97.15% |
| TF | 2 | 0 | 36 | 305459474 | 59.88% |
| TG | 71 | 1402 | 110 | 26516146725 | 98.87% |
| TH | 2292 | 43569 | 1175 | 595602323714 | 95.58% |
| TJ | 148 | 392 | 158 | 636925557553 | 98.12% |
| TK | 24 | 15 | 100 | 4788975911 | 95.39% |
| TL | 88 | 126 | 99 | 43595854121 | 96.44% |
| TM | 63 | 126 | 94 | 69265717808 | 99.64% |
| TN | 150 | 30817 | 142 | 116746957624 | 97.54% |
| TO | 47 | 55 | 83 | 17540182310 | 99.20% |
| TR | 3718 | 68028 | 2643 | 22486789216117 | 93.64% |
| TT | 167 | 2226 | 138 | 1177239763781 | 70.56% |
| TV | 41 | 46 | 89 | 11247679780 | 88.42% |
| TW | 2550 | 147481 | 2576 | 11501069720397 | 98.79% |
| TZ | 357 | 4488 | 239 | 267349570921 | 98.78% |
| UA | 4994 | 34813 | 2741 | 11756605513280 | 93.32% |
| UG | 271 | 5773 | 170 | 132373828144 | 95.54% |
| UM | 5 | 2 | 55 | 322828569 | 60.28% |
| US | 134172 | 5881871 | 498048 | 4835665282220572 | 99.76% |
| UY | 248 | 10380 | 181 | 177573408826 | 99.81% |
| UZ | 554 | 1659 | 235 | 1083166673730 | 99.45% |
| VA | 35 | 51 | 94 | 103208860193 | 99.95% |
| VC | 82 | 104 | 84 | 69482625317 | 99.68% |
| VE | 979 | 30275 | 629 | 2253462311583 | 99.23% |
| VG | 344 | 554 | 397 | 251162925352 | 87.77% |
| VI | 85 | 532 | 167 | 30583835956 | 92.45% |
| VN | 2264 | 65537 | 1536 | 1860176610447 | 99.18% |
| VU | 65 | 86 | 271 | 51950517285 | 79.82% |
| WF | 22 | 16 | 65 | 4770822436 | 95.50% |
| WS | 60 | 97 | 87 | 21851927079 | 92.79% |
| XK | 184 | 622 | 58 | 111804022773 | 60.60% |
| YE | 116 | 958 | 119 | 207080387640 | 99.82% |
| YT | 42 | 45 | 76 | 635498789 | 56.23% |
| ZA | 5221 | 118030 | 3435 | 23435221558865 | 99.55% |
| ZM | 359 | 6506 | 148 | 44204360265 | 95.69% |
| ZW | 257 | 931 | 192 | 65564843615 | 98.33% |

<!-- GEOIP-STATISTICS:END -->
