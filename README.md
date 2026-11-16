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

Last run: `2026-11-16T04:48:16Z`

Input rows: **40479873** · Published blocks: **2392862**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 596718 | 0 | 0 | 0 | 11047834 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 814011 | 0 | 0 | 21437479 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 596718 | 0 | 0 | 0 | 7595747 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 814011 | 0 | 0 | 25515786 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 564311 | 0 | 0 | 0 | 11143126 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 494164 | 0 | 0 | 15054413 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 566224 | 0 | 0 | 0 | 11857570 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 494248 | 0 | 0 | 34103112 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 406791 | 0 | 0 | 0 | 8861753 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 483486 | 0 | 1 | 15946283 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 406608 | 0 | 0 | 0 | 6049428 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 482754 | 0 | 1 | 18777039 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 425632 | 0 | 0 | 0 | 9271249 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 577034 | 0 | 1 | 17211693 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1039425 | 0 | 0 | 0 | 13705794 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2976374 | 0 | 0 | 96771741 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1398367 | 3209044 | 0 | 0 | 227037019 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 804555 | 1404435 | 0 | 0 | 17679336 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177309 | 69868 | 0 | 0 | 3402409 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 294706 | 489795 | 0 | 97499 | 41063982 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 294706 | 489795 | 0 | 97986 | 18559058 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4338 | 1725 | 0 | 0 | 98701 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3884 | 0 | 0 | 0 | 82599 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 813853 | 427285 | 0 | 411690 | 18511873 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262628 | 71134 | 0 | 10 | 34627770 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4958198 | 0 | 0 | 0 | 66567473 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380280 | 0 | 0 | 11696034 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6219 | 3439 | 0 | 0 | 153986 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 578 | 0 | 0 | 36064 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 565582 | 495307 | 0 | 0 | 3394049 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 613683 | 3702257408 | 1778929 | 293421847207842436812546030166054273024 |
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
| AE | 2871 | 25534 | 2577 | 12804979159766 | 89.92% |
| AF | 220 | 875 | 526 | 121136965522 | 97.42% |
| AG | 119 | 177 | 185 | 40393306171 | 65.26% |
| AI | 44 | 45 | 104 | 5847643194 | 67.40% |
| AL | 576 | 1413 | 1246 | 2508745169819 | 95.10% |
| AM | 354 | 2322 | 519 | 1263612958926 | 95.01% |
| AO | 561 | 5127 | 297 | 147599966561 | 98.68% |
| AQ | 56 | 5 | 303 | 472114316 | 43.65% |
| AR | 3796 | 77018 | 2852 | 23626711490783 | 98.82% |
| AS | 41 | 29 | 310 | 9243191473 | 96.04% |
| AT | 4743 | 45214 | 4445 | 10505653847671 | 97.35% |
| AU | 13408 | 219813 | 71508 | 2309962499421337693 | 100.00% |
| AW | 50 | 449 | 100 | 11534577721 | 92.96% |
| AX | 86 | 132 | 627 | 69552488267 | 72.57% |
| AZ | 402 | 3073 | 340 | 1963968520834 | 98.36% |
| BA | 333 | 3265 | 743 | 833827496478 | 95.13% |
| BB | 194 | 507 | 111 | 69400589006 | 99.69% |
| BD | 2661 | 8185 | 2712 | 3577593881839 | 96.48% |
| BE | 8420 | 57968 | 4865 | 4398028310494091 | 66.61% |
| BF | 125 | 1387 | 144 | 91245525303 | 99.43% |
| BG | 2376 | 16713 | 1434 | 5137768961857 | 96.07% |
| BH | 336 | 5248 | 320 | 406108709712 | 97.65% |
| BI | 59 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 85 | 622 | 127 | 56582572337 | 96.84% |
| BL | 24 | 10 | 88 | 1047851056 | 57.46% |
| BM | 328 | 498 | 109 | 43426868537 | 99.37% |
| BN | 111 | 879 | 119 | 26250616879 | 99.06% |
| BO | 324 | 4753 | 307 | 168302404201 | 98.07% |
| BQ | 38 | 103 | 71 | 17538740273 | 98.90% |
| BR | 11532 | 336333 | 42520 | 37428261883634 | 99.23% |
| BS | 141 | 634 | 167 | 14004634198 | 96.65% |
| BT | 125 | 226 | 160 | 60963021393 | 99.36% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 193 | 720 | 123 | 78628794672 | 97.43% |
| BY | 395 | 7406 | 432 | 700163615027 | 99.49% |
| BZ | 219 | 469 | 199 | 192579187530 | 89.06% |
| CA | 22138 | 285197 | 34467 | 6373167513159 | 95.83% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 318 | 940 | 203 | 155963863379 | 98.74% |
| CF | 49 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 119 | 36988538929 | 97.17% |
| CH | 7051 | 98185 | 57537 | 15878800939037195872 | 99.98% |
| CI | 348 | 6714 | 170 | 194921285946 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2065 | 41387 | 2396 | 1419536887487 | 96.63% |
| CM | 190 | 2440 | 160 | 67900151884 | 98.18% |
| CN | 6952 | 1347654 | 12776 | 294811339781618 | 99.94% |
| CO | 3181 | 70295 | 1943 | 1985927241616 | 98.00% |
| CR | 589 | 9167 | 370 | 461702267501 | 97.36% |
| CU | 96 | 1023 | 130 | 26439696698 | 99.17% |
| CV | 83 | 242 | 145 | 13715367984 | 97.22% |
| CW | 139 | 718 | 103 | 63125205037 | 92.43% |
| CX | 8 | 0 | 81 | 727377959 | 41.43% |
| CY | 688 | 4144 | 707 | 5999835421536 | 94.09% |
| CZ | 3870 | 33844 | 6462 | 13303858010802 | 97.50% |
| DE | 38233 | 550865 | 84811 | 115519264288441 | 96.02% |
| DJ | 65 | 247 | 119 | 9641322543 | 94.96% |
| DK | 3616 | 50105 | 2095 | 9671577762731 | 93.29% |
| DM | 79 | 71 | 118 | 44401881648 | 73.08% |
| DO | 747 | 8092 | 629 | 932803542372 | 98.92% |
| DZ | 234 | 18724 | 253 | 44176664759 | 95.48% |
| EC | 1090 | 11268 | 662 | 1162867667848 | 98.69% |
| EE | 1202 | 6502 | 1998 | 2725760730573 | 96.05% |
| EG | 800 | 94821 | 291 | 17726611005855 | 100.00% |
| EH | 4 | 0 | 55 | 457959450 | 46.78% |
| ER | 46 | 29 | 91 | 966146091 | 59.37% |
| ES | 11182 | 139969 | 6965 | 26203110609665 | 97.12% |
| ET | 95 | 1482 | 133 | 1117547837495 | 99.96% |
| FI | 4117 | 58732 | 34165 | 6895831438105 | 96.70% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 39 | 26 | 84 | 612231207 | 46.31% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 87 | 108187544876 | 99.57% |
| FR | 27563 | 325719 | 21157 | 67550663173736 | 98.44% |
| GA | 90 | 1558 | 108 | 35779722537 | 94.29% |
| GB | 40113 | 498342 | 61089 | 105927570265834 | 93.85% |
| GD | 101 | 128 | 104 | 82705773099 | 96.88% |
| GE | 494 | 5121 | 360 | 1831825930363 | 95.38% |
| GF | 312 | 257 | 80 | 23010009128 | 97.39% |
| GG | 145 | 263 | 94 | 112129796399 | 81.66% |
| GH | 360 | 8925 | 264 | 1232412718691 | 99.85% |
| GI | 154 | 329 | 108 | 197924941868 | 90.89% |
| GL | 103 | 144 | 169 | 43621736571 | 99.27% |
| GM | 75 | 1103 | 109 | 39586920746 | 98.99% |
| GN | 77 | 168 | 117 | 48293805098 | 97.24% |
| GP | 219 | 810 | 104 | 54408921143 | 73.07% |
| GQ | 52 | 94 | 94 | 9454960683 | 96.02% |
| GR | 1542 | 23121 | 1003 | 2592595346191 | 96.51% |
| GS | 4 | 2 | 41 | 406904854 | 59.19% |
| GT | 675 | 2899 | 279 | 327769490034 | 99.41% |
| GU | 190 | 995 | 129 | 34757730620 | 99.56% |
| GW | 53 | 42 | 94 | 9320609836 | 96.97% |
| GY | 58 | 288 | 109 | 35023940652 | 96.09% |
| HK | 13982 | 95428 | 8052 | 5095704930124 | 85.93% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 843 | 1995 | 268 | 391924757105 | 94.76% |
| HR | 861 | 9689 | 432 | 1591451112661 | 96.38% |
| HT | 87 | 624 | 135 | 26468807232 | 93.34% |
| HU | 2028 | 23382 | 1304 | 3217142990275 | 98.33% |
| ID | 6920 | 81235 | 5260 | 3840379793519 | 98.07% |
| IE | 8176 | 72153 | 8647 | 4986553333178 | 88.65% |
| IL | 2510 | 37110 | 1592 | 4297859406448 | 97.41% |
| IM | 348 | 477 | 190 | 361489639998 | 96.14% |
| IN | 13412 | 197420 | 16846 | 28970333007827 | 97.84% |
| IO | 27 | 14 | 84 | 4755679276 | 95.94% |
| IQ | 677 | 3346 | 980 | 3415584355240 | 96.90% |
| IR | 1984 | 42739 | 2215 | 16992302664587 | 98.44% |
| IS | 340 | 3641 | 492 | 1562904234875 | 97.32% |
| IT | 9959 | 227826 | 9593 | 47054042183563 | 99.07% |
| JE | 242 | 640 | 116 | 219604756782 | 95.45% |
| JM | 253 | 1131 | 155 | 1122142712410 | 99.96% |
| JO | 447 | 2703 | 255 | 1121936403805 | 97.90% |
| JP | 12590 | 782634 | 14934 | 53263147347134 | 88.23% |
| KE | 1086 | 25125 | 665 | 692347381945 | 99.04% |
| KG | 238 | 1186 | 222 | 529225200237 | 99.38% |
| KH | 557 | 2014 | 301 | 430692688488 | 81.58% |
| KI | 79 | 36 | 136 | 13520201802 | 97.88% |
| KM | 67 | 53 | 85 | 9337298983 | 96.78% |
| KN | 100 | 133 | 105 | 5765067062 | 76.99% |
| KP | 9 | 11 | 419 | 499122480 | 55.30% |
| KR | 4055 | 455206 | 3656 | 22712175114832 | 99.87% |
| KW | 439 | 7585 | 246 | 877742068644 | 99.72% |
| KY | 129 | 448 | 108 | 17861637426 | 99.08% |
| KZ | 1408 | 13990 | 1217 | 3150641204788 | 86.82% |
| LA | 146 | 414 | 151 | 121260925231 | 88.17% |
| LB | 391 | 2386 | 293 | 3015926620999 | 97.81% |
| LC | 96 | 168 | 93 | 78392389925 | 97.53% |
| LI | 337 | 451 | 201 | 425652453418 | 93.84% |
| LK | 252 | 2325 | 266 | 82523808118 | 94.32% |
| LR | 82 | 440 | 107 | 27275917612 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1451 | 11931 | 972 | 4177006927762 | 84.82% |
| LU | 1559 | 6707 | 889 | 2222696019825 | 96.44% |
| LV | 1252 | 8168 | 781 | 3235507120349 | 65.64% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 342 | 48135 | 267 | 1171100463783 | 99.78% |
| MC | 279 | 281 | 107 | 142166714925 | 99.88% |
| MD | 748 | 4970 | 973 | 5953066430795 | 64.09% |
| ME | 147 | 822 | 131 | 348120040250 | 99.48% |
| MF | 57 | 105 | 90 | 5443268655 | 92.31% |
| MG | 87 | 2309 | 114 | 22644159028 | 98.09% |
| MH | 85 | 94 | 81 | 4806418475 | 95.24% |
| MK | 255 | 2801 | 187 | 1138602815342 | 97.19% |
| ML | 83 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 236 | 737 | 286 | 133921418861 | 99.23% |
| MO | 104 | 1500 | 719 | 39116567101 | 99.15% |
| MP | 94 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 326 | 814 | 72 | 104815669284 | 89.69% |
| MR | 76 | 234 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 79 | 644673839 | 51.26% |
| MT | 321 | 2677 | 227 | 508545919800 | 92.95% |
| MU | 1968 | 23854 | 8194 | 4455183892967739 | 100.00% |
| MV | 118 | 403 | 132 | 18180677685 | 97.48% |
| MW | 124 | 2201 | 146 | 1242315518001 | 99.85% |
| MX | 4437 | 122492 | 2537 | 3647097565610 | 97.05% |
| MY | 2852 | 34653 | 1627 | 911924806734 | 96.08% |
| MZ | 190 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 347 | 1857 | 186 | 29169103150 | 99.37% |
| NC | 77 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1103 | 12921 | 709 | 604382832767 | 98.31% |
| NI | 170 | 1736 | 121 | 95171113283 | 98.22% |
| NL | 24873 | 214639 | 486033 | 8407492237829808 | 99.94% |
| NO | 3348 | 63979 | 2955 | 10986138604338 | 97.26% |
| NP | 468 | 2392 | 344 | 430212202597 | 94.05% |
| NR | 42 | 53 | 88 | 8951228462 | 98.41% |
| NU | 28 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2287 | 28873 | 1784 | 1355928777234 | 94.98% |
| OM | 317 | 4418 | 202 | 595072376918 | 96.04% |
| PA | 558 | 7098 | 487 | 429361298739 | 92.96% |
| PE | 1309 | 14690 | 1077 | 1072415464975 | 96.36% |
| PF | 93 | 312 | 143 | 21875058780 | 99.30% |
| PG | 233 | 348 | 153 | 73449136460 | 87.10% |
| PH | 2264 | 26809 | 1274 | 1050699735301 | 97.67% |
| PK | 1460 | 23636 | 741 | 907296026870 | 97.95% |
| PL | 8781 | 83154 | 5399 | 28613661410383 | 96.76% |
| PM | 42 | 25 | 81 | 4671532333 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 655 | 4574 | 337 | 230687482761 | 99.07% |
| PS | 303 | 3312 | 179 | 979424835920 | 93.84% |
| PT | 1886 | 27061 | 2137 | 3328261586144 | 93.01% |
| PW | 48 | 43 | 87 | 9085575215 | 97.53% |
| PY | 447 | 5093 | 274 | 451733299855 | 97.52% |
| QA | 499 | 5197 | 367 | 526445600733 | 96.23% |
| RE | 326 | 1985 | 100 | 95622438962 | 96.43% |
| RO | 3961 | 28323 | 2523 | 5572234413544 | 95.57% |
| RS | 1063 | 9072 | 1345 | 1887339858360 | 97.41% |
| RU | 13808 | 180906 | 16070 | 54670074093493 | 90.12% |
| RW | 190 | 1266 | 125 | 1147559424301 | 99.91% |
| SA | 1690 | 48871 | 924 | 2815443834340 | 96.57% |
| SB | 55 | 71 | 99 | 17944013095 | 90.06% |
| SC | 959 | 2659 | 3266 | 3301050835498 | 61.88% |
| SD | 116 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16253 | 126528 | 57030 | 30393369330758 | 79.91% |
| SG | 10977 | 110813 | 52785 | 133667670451351 | 98.94% |
| SH | 6 | 2 | 129 | 575138821 | 54.83% |
| SI | 1062 | 10331 | 618 | 2079447586748 | 95.35% |
| SJ | 7 | 2 | 42 | 374468627 | 60.68% |
| SK | 1234 | 11157 | 940 | 3009684342038 | 96.94% |
| SL | 158 | 359 | 112 | 26584804400 | 96.08% |
| SM | 73 | 173 | 97 | 197980645926 | 93.41% |
| SN | 119 | 1646 | 149 | 27161196614 | 98.08% |
| SO | 76 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 340 | 107 | 17813725490 | 98.45% |
| SS | 102 | 135 | 110 | 31098290222 | 98.66% |
| ST | 49 | 60 | 91 | 4773833771 | 96.05% |
| SV | 343 | 2787 | 210 | 211465053033 | 97.09% |
| SX | 48 | 128 | 83 | 30575685670 | 94.55% |
| SY | 298 | 5173 | 138 | 876931813423 | 99.64% |
| SZ | 77 | 243 | 106 | 30376013879 | 99.14% |
| TC | 44 | 56 | 85 | 1451585833 | 71.10% |
| TD | 94 | 134 | 105 | 13816294446 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2294 | 43584 | 1272 | 595753118294 | 95.52% |
| TJ | 159 | 402 | 169 | 641204728380 | 98.14% |
| TK | 26 | 15 | 114 | 4789889072 | 95.34% |
| TL | 84 | 121 | 109 | 43596832814 | 92.49% |
| TM | 63 | 126 | 105 | 69114982713 | 99.77% |
| TN | 148 | 30817 | 149 | 116780507712 | 97.52% |
| TO | 47 | 55 | 85 | 17523466282 | 99.25% |
| TR | 3601 | 68045 | 2982 | 22547089080188 | 93.78% |
| TT | 167 | 2226 | 146 | 1177257587275 | 70.56% |
| TV | 41 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9294 | 160420 | 3119 | 11549441848116 | 98.67% |
| TZ | 401 | 4500 | 249 | 267366409333 | 98.77% |
| UA | 4886 | 34615 | 3350 | 11756587050073 | 93.15% |
| UG | 266 | 5774 | 183 | 132375003453 | 96.35% |
| UM | 5 | 2 | 82 | 355003435 | 58.92% |
| US | 135490 | 5869837 | 539382 | 4835846586996129 | 99.76% |
| UY | 246 | 10118 | 193 | 177623867459 | 99.80% |
| UZ | 548 | 1661 | 281 | 1083134232177 | 99.45% |
| VA | 36 | 51 | 105 | 103325378861 | 99.88% |
| VC | 84 | 103 | 91 | 69500514346 | 99.67% |
| VE | 993 | 30432 | 662 | 2279266779578 | 99.37% |
| VG | 386 | 523 | 406 | 255473979451 | 89.93% |
| VI | 92 | 548 | 170 | 30616539205 | 92.42% |
| VN | 2266 | 65516 | 1794 | 1860462179344 | 99.21% |
| VU | 67 | 87 | 274 | 51951561521 | 80.96% |
| WF | 22 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 198 | 623 | 72 | 111806775286 | 60.60% |
| YE | 115 | 958 | 119 | 207096833082 | 99.81% |
| YT | 73 | 42 | 75 | 602988584 | 56.00% |
| ZA | 5144 | 117572 | 2690 | 41031802073927 | 78.31% |
| ZM | 366 | 6514 | 161 | 48484905299 | 96.08% |
| ZW | 290 | 937 | 196 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
