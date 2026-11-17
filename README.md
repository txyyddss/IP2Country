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

Last run: `2026-11-17T06:34:54Z`

Input rows: **40766133** · Published blocks: **2395721**

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
| Server-Country via sapics IPv4 | fresh | 406667 | 0 | 0 | 0 | 8859836 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 483591 | 0 | 1 | 15946655 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 406791 | 0 | 0 | 0 | 6051426 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 483486 | 0 | 1 | 18814506 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 425507 | 0 | 0 | 0 | 9269303 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 577139 | 0 | 1 | 17212065 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 723572 | 0 | 0 | 1487 | 9631726 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288172 | 0 | 116 | 8266337 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321327 | 0 | 0 | 0 | 7039331 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 84215 | 0 | 0 | 3672965 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 1043521 | 0 | 0 | 0 | 13736087 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 3113188 | 0 | 0 | 102955245 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1402246 | 3346141 | 0 | 0 | 234656128 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 377372 | 0 | 0 | 0 | 5932108 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1597048 | 0 | 0 | 26588391 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 804434 | 1405472 | 0 | 0 | 17686664 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177430 | 69881 | 0 | 0 | 3403525 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 294825 | 489937 | 0 | 97510 | 41076425 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 294825 | 489937 | 0 | 97997 | 18564579 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4339 | 1724 | 0 | 0 | 98697 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3895 | 0 | 0 | 0 | 83091 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327710 | 0 | 0 | 0 | 537626 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 814446 | 427499 | 0 | 411732 | 18523034 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262637 | 71141 | 0 | 10 | 34629443 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4958198 | 0 | 0 | 0 | 66567473 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 1380280 | 0 | 0 | 11696034 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802848 | 0 | 0 | 0 | 11122742 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6226 | 3436 | 0 | 0 | 154047 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 578 | 0 | 0 | 36064 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 566994 | 494983 | 0 | 0 | 3396552 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 614293 | 3702257408 | 1781178 | 293421847207842436812546030166054273024 |
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
| AD | 218 | 309 | 190 | 43728563274 | 92.57% |
| AE | 2841 | 25549 | 2535 | 12770686497504 | 90.03% |
| AF | 220 | 875 | 529 | 121136965525 | 97.42% |
| AG | 127 | 177 | 186 | 40393306172 | 65.26% |
| AI | 44 | 45 | 104 | 5847643194 | 67.40% |
| AL | 595 | 1418 | 1248 | 2508745104282 | 95.10% |
| AM | 357 | 2322 | 522 | 1297972697296 | 94.26% |
| AO | 548 | 5126 | 370 | 147599974754 | 98.68% |
| AQ | 56 | 5 | 305 | 472130703 | 43.65% |
| AR | 3799 | 77013 | 2776 | 23626711480170 | 98.86% |
| AS | 41 | 29 | 309 | 9243191472 | 96.04% |
| AT | 4743 | 45213 | 4397 | 10505922004668 | 97.35% |
| AU | 13414 | 219805 | 71511 | 2309962499421325409 | 100.00% |
| AW | 50 | 449 | 101 | 11534577722 | 92.96% |
| AX | 93 | 132 | 627 | 69552488267 | 72.57% |
| AZ | 398 | 3069 | 348 | 1963968455298 | 98.36% |
| BA | 333 | 3265 | 743 | 833827496478 | 95.13% |
| BB | 194 | 507 | 110 | 69400589005 | 99.69% |
| BD | 2666 | 8190 | 2711 | 3577593881841 | 96.48% |
| BE | 8431 | 58034 | 4850 | 4398028310707085 | 66.61% |
| BF | 126 | 1388 | 144 | 91245525303 | 99.43% |
| BG | 2386 | 16711 | 1436 | 5137768961859 | 96.07% |
| BH | 336 | 5248 | 320 | 406108709712 | 97.65% |
| BI | 59 | 168 | 111 | 26969694761 | 92.79% |
| BJ | 85 | 622 | 127 | 56582572337 | 96.84% |
| BL | 23 | 10 | 88 | 1047851056 | 57.46% |
| BM | 320 | 501 | 109 | 43426868537 | 99.37% |
| BN | 111 | 879 | 119 | 26250616879 | 99.06% |
| BO | 324 | 4753 | 307 | 168302404201 | 98.07% |
| BQ | 38 | 103 | 71 | 17538740273 | 98.90% |
| BR | 11566 | 336360 | 42478 | 37428260822841 | 99.22% |
| BS | 140 | 634 | 166 | 14004568662 | 96.65% |
| BT | 125 | 226 | 160 | 60963021393 | 99.36% |
| BV | 1 | 1 | 4 | 17825793 | 34.31% |
| BW | 188 | 720 | 123 | 78628794672 | 97.43% |
| BY | 401 | 7405 | 433 | 700163598643 | 99.49% |
| BZ | 224 | 470 | 200 | 192713405258 | 89.02% |
| CA | 22145 | 285207 | 36241 | 6373148559849 | 95.84% |
| CC | 7 | 0 | 50 | 827254806 | 42.84% |
| CD | 339 | 940 | 203 | 155963863379 | 98.74% |
| CF | 54 | 40 | 99 | 1117335594 | 59.87% |
| CG | 92 | 585 | 119 | 36988538929 | 97.17% |
| CH | 6954 | 98598 | 57517 | 15878800934688456270 | 99.98% |
| CI | 348 | 6714 | 169 | 194921285945 | 98.93% |
| CK | 35 | 41 | 91 | 9234279472 | 96.05% |
| CL | 2079 | 41388 | 2387 | 1419535904326 | 96.63% |
| CM | 189 | 2439 | 160 | 67900151884 | 98.18% |
| CN | 6987 | 1347805 | 12881 | 294811339896345 | 99.94% |
| CO | 3184 | 70295 | 1945 | 1985927241618 | 98.00% |
| CR | 590 | 9167 | 372 | 461702267503 | 97.36% |
| CU | 96 | 1023 | 129 | 26439696697 | 99.17% |
| CV | 82 | 242 | 145 | 13715367984 | 97.22% |
| CW | 139 | 718 | 103 | 63125205037 | 92.43% |
| CX | 8 | 0 | 81 | 727377959 | 41.43% |
| CY | 785 | 4143 | 708 | 5999835421537 | 94.09% |
| CZ | 3898 | 33844 | 6431 | 13303842265519 | 97.50% |
| DE | 38336 | 550865 | 84826 | 115514968385974 | 96.03% |
| DJ | 65 | 247 | 119 | 9641322543 | 94.96% |
| DK | 3646 | 50104 | 2095 | 9671577762731 | 93.30% |
| DM | 76 | 71 | 118 | 44401881648 | 73.08% |
| DO | 747 | 8092 | 629 | 932803542372 | 98.92% |
| DZ | 233 | 18724 | 252 | 44176664758 | 95.48% |
| EC | 1092 | 11267 | 662 | 1162867667848 | 98.69% |
| EE | 1201 | 6509 | 1646 | 2725760633112 | 96.02% |
| EG | 800 | 94821 | 291 | 17726610989472 | 100.00% |
| EH | 4 | 0 | 55 | 457959450 | 46.78% |
| ER | 46 | 29 | 92 | 966146092 | 59.37% |
| ES | 11203 | 139916 | 6930 | 26203110723850 | 97.12% |
| ET | 95 | 1482 | 133 | 1117547837495 | 99.96% |
| FI | 4099 | 58760 | 34091 | 6895831421862 | 96.70% |
| FJ | 114 | 590 | 117 | 43379780397 | 99.48% |
| FK | 39 | 26 | 84 | 612231207 | 46.30% |
| FM | 57 | 49 | 105 | 8921737274 | 78.74% |
| FO | 46 | 181 | 87 | 108187544876 | 99.57% |
| FR | 27590 | 325683 | 21092 | 67548566284902 | 98.44% |
| GA | 90 | 1558 | 108 | 35779722537 | 94.29% |
| GB | 40154 | 498256 | 61066 | 105927301382427 | 93.85% |
| GD | 101 | 128 | 104 | 82705773099 | 96.88% |
| GE | 494 | 5119 | 360 | 1831825930363 | 95.38% |
| GF | 302 | 260 | 81 | 23010009129 | 97.39% |
| GG | 142 | 263 | 94 | 112129796399 | 81.66% |
| GH | 360 | 8925 | 264 | 1232412980834 | 99.85% |
| GI | 148 | 327 | 108 | 197924941868 | 90.89% |
| GL | 103 | 144 | 169 | 43621736571 | 99.27% |
| GM | 75 | 1103 | 109 | 39586920746 | 98.99% |
| GN | 77 | 168 | 117 | 48293805098 | 97.24% |
| GP | 227 | 807 | 105 | 54408921144 | 73.07% |
| GQ | 52 | 94 | 95 | 9454960684 | 96.02% |
| GR | 1539 | 23120 | 1004 | 2592595346192 | 96.51% |
| GS | 4 | 2 | 41 | 406904854 | 59.19% |
| GT | 671 | 2899 | 278 | 327769490033 | 99.41% |
| GU | 182 | 995 | 129 | 34757730620 | 99.56% |
| GW | 53 | 42 | 94 | 9320609836 | 96.97% |
| GY | 58 | 288 | 109 | 35023940652 | 96.09% |
| HK | 13920 | 95338 | 8029 | 5091409340242 | 85.96% |
| HM | 1 | 2 | 2 | 33554432 | 41.67% |
| HN | 844 | 1995 | 268 | 391924757105 | 94.76% |
| HR | 863 | 9689 | 431 | 1591451096277 | 96.38% |
| HT | 87 | 624 | 135 | 26468807232 | 93.34% |
| HU | 2024 | 23382 | 1321 | 3217143006102 | 98.33% |
| ID | 6907 | 81225 | 5262 | 3840379777136 | 98.07% |
| IE | 8176 | 72155 | 8712 | 4986553070985 | 88.67% |
| IL | 2506 | 37116 | 1542 | 4297859386436 | 97.41% |
| IM | 348 | 476 | 190 | 361489639998 | 96.14% |
| IN | 13406 | 197398 | 16536 | 29004691681450 | 97.79% |
| IO | 27 | 14 | 83 | 4755679275 | 95.94% |
| IQ | 679 | 3347 | 981 | 3415584355240 | 96.90% |
| IR | 1989 | 42739 | 2187 | 17026595408752 | 98.30% |
| IS | 344 | 3642 | 492 | 1562904218491 | 97.32% |
| IT | 9966 | 227815 | 9353 | 47054042200088 | 99.07% |
| JE | 232 | 640 | 116 | 219604756782 | 95.45% |
| JM | 253 | 1131 | 155 | 1122142712410 | 99.96% |
| JO | 449 | 2703 | 256 | 1121936403806 | 97.90% |
| JP | 12596 | 782643 | 14983 | 53263146314908 | 88.23% |
| KE | 1156 | 25127 | 670 | 692347316409 | 99.09% |
| KG | 238 | 1186 | 222 | 529225200237 | 99.38% |
| KH | 553 | 2010 | 301 | 430692688488 | 81.58% |
| KI | 79 | 36 | 136 | 13520201802 | 97.88% |
| KM | 67 | 53 | 85 | 9337298983 | 96.78% |
| KN | 133 | 135 | 123 | 10060017974 | 69.74% |
| KP | 9 | 11 | 419 | 499122480 | 55.30% |
| KR | 4044 | 455043 | 3654 | 22712174934606 | 99.87% |
| KW | 439 | 7585 | 246 | 877742068644 | 99.72% |
| KY | 129 | 449 | 108 | 17861637426 | 99.08% |
| KZ | 1391 | 13991 | 1215 | 3150641221169 | 86.82% |
| LA | 146 | 414 | 151 | 121260925231 | 88.17% |
| LB | 389 | 2384 | 294 | 3015926621000 | 97.70% |
| LC | 98 | 168 | 93 | 78392389925 | 97.53% |
| LI | 291 | 452 | 201 | 429947486247 | 93.57% |
| LK | 255 | 2325 | 266 | 82523808118 | 94.32% |
| LR | 83 | 440 | 106 | 27275917611 | 97.70% |
| LS | 68 | 507 | 95 | 26802794537 | 98.25% |
| LT | 1445 | 11936 | 972 | 4177006927825 | 84.82% |
| LU | 1559 | 6708 | 875 | 2222696003442 | 96.44% |
| LV | 1252 | 8162 | 780 | 3235507120349 | 65.64% |
| LY | 152 | 1923 | 108 | 99614748733 | 89.61% |
| MA | 342 | 48135 | 267 | 1171100463783 | 99.78% |
| MC | 279 | 281 | 107 | 142166714925 | 99.88% |
| MD | 749 | 4969 | 974 | 5953066430796 | 64.09% |
| ME | 148 | 823 | 131 | 348120040250 | 99.48% |
| MF | 56 | 105 | 90 | 5443268655 | 92.31% |
| MG | 87 | 2309 | 114 | 22644159028 | 98.09% |
| MH | 57 | 102 | 82 | 4806418476 | 95.22% |
| MK | 255 | 2801 | 186 | 1138468597613 | 97.20% |
| ML | 83 | 388 | 90 | 22675267630 | 97.52% |
| MM | 460 | 1170 | 337 | 353155015037 | 97.86% |
| MN | 236 | 737 | 286 | 133921418861 | 99.23% |
| MO | 105 | 1500 | 717 | 39116546364 | 99.15% |
| MP | 88 | 16 | 78 | 4805746733 | 95.20% |
| MQ | 324 | 818 | 72 | 104815669284 | 89.69% |
| MR | 76 | 234 | 96 | 26517694507 | 98.88% |
| MS | 31 | 15 | 79 | 644673839 | 51.26% |
| MT | 320 | 2676 | 226 | 508545919799 | 92.95% |
| MU | 1978 | 23846 | 8206 | 4455171008065851 | 100.00% |
| MV | 119 | 403 | 132 | 18180677685 | 97.48% |
| MW | 124 | 2201 | 146 | 1242315518001 | 99.85% |
| MX | 4427 | 122529 | 2549 | 3651392508422 | 97.02% |
| MY | 2856 | 34659 | 1628 | 911924741200 | 96.08% |
| MZ | 195 | 1914 | 133 | 69652197437 | 97.38% |
| NA | 341 | 1861 | 186 | 29169103150 | 99.37% |
| NC | 77 | 571 | 108 | 47910414645 | 99.48% |
| NE | 64 | 218 | 92 | 22122717221 | 98.72% |
| NF | 37 | 3 | 85 | 357558318 | 59.06% |
| NG | 1100 | 12921 | 709 | 604382767233 | 98.31% |
| NI | 170 | 1736 | 122 | 95171113284 | 98.22% |
| NL | 24890 | 214608 | 486236 | 8407427831422603 | 99.94% |
| NO | 3353 | 63976 | 2936 | 10986138653488 | 97.26% |
| NP | 469 | 2393 | 343 | 430212202596 | 94.05% |
| NR | 42 | 53 | 88 | 8951228462 | 98.41% |
| NU | 27 | 11 | 73 | 4671334441 | 81.46% |
| NZ | 2287 | 28873 | 1784 | 1355928777234 | 94.98% |
| OM | 317 | 4418 | 204 | 595072376920 | 96.04% |
| PA | 560 | 7098 | 487 | 427213830452 | 93.33% |
| PE | 1314 | 14692 | 1082 | 1072415464980 | 96.36% |
| PF | 92 | 312 | 144 | 21875058781 | 99.30% |
| PG | 234 | 349 | 153 | 73449136460 | 87.10% |
| PH | 2270 | 26813 | 1275 | 1050699735302 | 97.67% |
| PK | 1463 | 23636 | 741 | 907295961335 | 97.95% |
| PL | 8783 | 83158 | 5401 | 28613661518414 | 96.74% |
| PM | 42 | 25 | 81 | 4671532333 | 96.85% |
| PN | 6 | 3 | 43 | 355461143 | 60.41% |
| PR | 662 | 4574 | 333 | 230687481737 | 99.07% |
| PS | 303 | 3312 | 179 | 979424835920 | 93.84% |
| PT | 1903 | 27062 | 2059 | 3362638421194 | 92.74% |
| PW | 48 | 43 | 86 | 9085575214 | 97.53% |
| PY | 448 | 5094 | 274 | 451733299855 | 97.52% |
| QA | 501 | 5197 | 368 | 526445600734 | 96.23% |
| RE | 332 | 2007 | 100 | 95622438962 | 96.43% |
| RO | 3987 | 28323 | 2520 | 5572234364396 | 95.57% |
| RS | 1059 | 9073 | 1342 | 1887339792822 | 97.41% |
| RU | 13810 | 180905 | 16093 | 54648599388150 | 90.13% |
| RW | 190 | 1266 | 125 | 1147559424301 | 99.91% |
| SA | 1691 | 48871 | 924 | 2815443834340 | 96.57% |
| SB | 55 | 71 | 99 | 17944013095 | 90.06% |
| SC | 943 | 2678 | 3265 | 3305345753642 | 61.91% |
| SD | 116 | 7412 | 111 | 69736980788 | 97.36% |
| SE | 16268 | 126527 | 56547 | 30393369313467 | 79.90% |
| SG | 10948 | 110810 | 52717 | 133667403981149 | 98.94% |
| SH | 6 | 2 | 129 | 575138821 | 54.83% |
| SI | 1061 | 10331 | 618 | 2079447586748 | 95.35% |
| SJ | 7 | 2 | 42 | 374468627 | 60.68% |
| SK | 1230 | 11152 | 939 | 3009684342037 | 96.94% |
| SL | 164 | 359 | 112 | 26584804400 | 96.08% |
| SM | 73 | 173 | 97 | 197980645926 | 93.41% |
| SN | 119 | 1646 | 149 | 27161196614 | 98.08% |
| SO | 76 | 225 | 163 | 48249762113 | 99.11% |
| SR | 51 | 340 | 107 | 17813725490 | 98.45% |
| SS | 105 | 135 | 110 | 31098290222 | 98.66% |
| ST | 49 | 60 | 91 | 4773833771 | 96.05% |
| SV | 339 | 2787 | 210 | 211465053033 | 97.09% |
| SX | 47 | 128 | 82 | 30441467942 | 94.79% |
| SY | 298 | 5173 | 138 | 876931813423 | 99.64% |
| SZ | 77 | 243 | 107 | 34670981175 | 95.12% |
| TC | 44 | 56 | 85 | 1451585833 | 71.10% |
| TD | 94 | 134 | 105 | 13816294446 | 97.12% |
| TF | 2 | 0 | 35 | 305455123 | 59.87% |
| TG | 72 | 1403 | 112 | 26550810919 | 98.82% |
| TH | 2290 | 43579 | 1273 | 595753052760 | 95.52% |
| TJ | 159 | 402 | 169 | 641204728380 | 98.14% |
| TK | 26 | 15 | 114 | 4789889072 | 95.34% |
| TL | 84 | 121 | 109 | 43596832814 | 92.49% |
| TM | 63 | 126 | 105 | 69114982713 | 99.77% |
| TN | 148 | 30817 | 149 | 116780507712 | 97.52% |
| TO | 47 | 55 | 85 | 17523466282 | 99.25% |
| TR | 3611 | 68050 | 2894 | 22568563917181 | 93.79% |
| TT | 167 | 2226 | 146 | 1177257587275 | 70.56% |
| TV | 41 | 46 | 97 | 11112278055 | 82.95% |
| TW | 9294 | 160420 | 3126 | 11549441815358 | 98.67% |
| TZ | 427 | 4498 | 249 | 267366409333 | 98.77% |
| UA | 4908 | 34630 | 3343 | 11756587902043 | 93.16% |
| UG | 266 | 5774 | 183 | 132375003453 | 96.35% |
| UM | 5 | 2 | 83 | 355003436 | 58.92% |
| US | 135676 | 5869487 | 541595 | 4835808341342091 | 99.76% |
| UY | 246 | 10118 | 194 | 177623867460 | 99.80% |
| UZ | 546 | 1661 | 281 | 1083134232177 | 99.45% |
| VA | 36 | 51 | 104 | 103325378860 | 99.88% |
| VC | 84 | 103 | 91 | 69500514346 | 99.67% |
| VE | 990 | 30432 | 659 | 2279266779575 | 99.37% |
| VG | 384 | 524 | 267 | 255457202227 | 89.93% |
| VI | 100 | 548 | 171 | 30633513029 | 94.72% |
| VN | 2267 | 65519 | 1799 | 1860462138391 | 99.21% |
| VU | 66 | 86 | 274 | 51951561521 | 80.96% |
| WF | 22 | 16 | 69 | 4788512810 | 95.30% |
| WS | 62 | 97 | 90 | 21837242667 | 92.81% |
| XK | 220 | 619 | 73 | 111806775287 | 60.60% |
| YE | 115 | 958 | 119 | 207096833082 | 99.81% |
| YT | 79 | 51 | 76 | 602988585 | 56.00% |
| ZA | 5159 | 117576 | 2696 | 41040392008521 | 78.31% |
| ZM | 348 | 6515 | 161 | 48484905299 | 96.08% |
| ZW | 304 | 934 | 196 | 65582730600 | 98.32% |

<!-- GEOIP-STATISTICS:END -->
