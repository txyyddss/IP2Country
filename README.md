# IP2Country GeoIP Generator

A Go program that combines public IP-to-country feeds into compact country databases and country-specific CSV files. It produces Xray/Mihomo V2Fly protobuf GeoIP, current sing-box rule sets, and a MaxMind DB country database.

The input inventory, parser settings, enabled state, vote groups, and source order live in [`config/sources.json`](config/sources.json). Feeds are enabled unless explicitly disabled in the config. The IANA feed supplies country-tagged allocation prefixes; built-in special-use ranges are excluded. Public Cloudflare and hosting ranges remain eligible.

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

Voting follows exact source boundaries. If countries differ inside an IPv4 /24 or IPv6 /64, the voter splits that cell into smaller CIDRs instead of replacing minority assignments with the cell's majority country. Partial assignments do not extend into uncovered addresses. Overlapping assignments within one vote group use the most specific prefix, then source and input-row order. Each group casts at most one vote per address; configured mirrors of DB-IP, GeoLite2, and IPinfo share their provider's group. Across groups, country ties follow config order, with country code as a deterministic fallback.

Confidence is winning votes divided by all participating group votes, averaged by address space when adjacent winning ranges merge. Country CSV vote totals count vote-address pairs for that output CIDR. No-country records, such as metadata-only MMDB entries and regional tags like `EU`, are reported separately from malformed records. Country ISO-code columns take precedence over country names. Integer address feeds can declare `ip_encoding` as `hex` or `decimal`; tdulcet TSV feeds use hexadecimal endpoints.

Range processing sorts exact boundaries on disk without expanding addresses, /24s, or /64s. It handles broad IPv6 networks with bounded event-sort memory. Preprocessing is O(n log n); final sorted-prefix lookups are O(log n). README /24 and /64 coverage totals are summed exactly and rounded once to whole units; JSON totals retain fractional units.

## Automated updates

The GeoIP workflow runs at 00:00 UTC each day and supports `workflow_dispatch`. It publishes one dated GitHub release per UTC day, reuses that release on reruns, and updates this README's generated statistics section. GitHub Actions caches source files between runs. The quality workflow runs formatting, tests, vet, and build checks on pushes and pull requests.

<!-- GEOIP-STATISTICS:START -->
## Generated data status

Last run: `2026-10-20T15:45:52Z`

Input rows: **39261406** · Published blocks: **1234383**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 597286 | 0 | 0 | 0 | 11032548 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 785668 | 0 | 0 | 20177287 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 597286 | 0 | 0 | 0 | 7580021 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 785668 | 0 | 0 | 24060852 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562225 | 0 | 0 | 0 | 11070001 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 500017 | 0 | 0 | 14970875 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 564219 | 0 | 0 | 532 | 11825975 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 500090 | 0 | 153 | 34516461 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 398706 | 0 | 0 | 0 | 8688499 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 471750 | 0 | 1 | 15531894 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 398438 | 0 | 0 | 0 | 5930350 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 471707 | 0 | 1 | 18351723 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 415649 | 0 | 0 | 0 | 9066810 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 563998 | 0 | 1 | 16757331 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 722193 | 0 | 0 | 2152 | 9616617 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288944 | 0 | 3996 | 8466227 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321682 | 0 | 0 | 0 | 7032792 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 87671 | 0 | 0 | 3803866 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 982030 | 0 | 0 | 0 | 12997637 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2707304 | 0 | 0 | 85118814 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPINFO Lite CSV | fresh | 1341028 | 2936383 | 0 | 0 | 208702865 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 785303 | 1368186 | 0 | 0 | 17235320 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177093 | 69680 | 0 | 0 | 3393343 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPGeo IPv4 CSV | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPGeo IPv6 CSV | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPRanges | fresh | 287188 | 480983 | 0 | 97783 | 40299556 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 287188 | 480983 | 0 | 98370 | 18199920 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4318 | 1675 | 0 | 0 | 97591 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3879 | 0 | 0 | 0 | 82305 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327605 | 0 | 0 | 0 | 537559 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 806857 | 417982 | 0 | 410891 | 18173463 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262213 | 70936 | 0 | 10 | 34563823 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802825 | 0 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| V2ray GeoIP | fresh | 597286 | 785617 | 0 | 24 | 23258601 | https://github.com/v2fly/geoip/releases/latest/download/geoip.dat |
| Meta Rules Dat | fresh | 512893 | 413486 | 0 | 10664 | 7647337 | https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/country.mmdb |
| Loyalsoldier GeoIP Enhanced | fresh | 557938 | 492270 | 0 | 10655 | 16587463 | https://raw.githubusercontent.com/Loyalsoldier/geoip/release/geoip.dat |
| gaoyifan China operator IPv4+IPv6 | fresh | 6526 | 3489 | 0 | 0 | 159327 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 574 | 0 | 0 | 36004 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| IPLocate CSV ZIP alternative | disabled | 0 | 0 | 0 | 0 | 0 | — |
| GeoLite2 Country via P3TERX | fresh | 565511 | 500934 | 0 | 0 | 3408573 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |
| GeoIP All-in-One consensus MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| laincat GeoIP MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| 17mon China CIDRs (stale backup) | disabled | 0 | 0 | 0 | 0 | 0 | — |

### Final country coverage

| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |
|---|---:|---:|---:|---:|---:|
| AD | 74 | 266 | 81 | 43816520498 | 91.87% |
| AE | 3008 | 25035 | 1608 | 16439317035124 | 85.10% |
| AF | 199 | 730 | 127 | 121093948030 | 96.89% |
| AG | 139 | 525 | 150 | 39401670966 | 67.04% |
| AI | 34 | 41 | 49 | 4787995953 | 77.82% |
| AL | 370 | 2005 | 953 | 2702872620619 | 94.50% |
| AM | 316 | 2330 | 195 | 1340409591876 | 92.62% |
| AO | 347 | 5149 | 112 | 177342976338 | 91.16% |
| AQ | 61 | 9 | 341 | 4411523107 | 68.38% |
| AR | 3201 | 76395 | 1711 | 23548965452189 | 99.19% |
| AS | 32 | 27 | 55 | 8780056887 | 98.89% |
| AT | 3856 | 44546 | 2201 | 10419335827386 | 97.30% |
| AU | 11127 | 212810 | 47294 | 2305941879014637051 | 100.00% |
| AW | 45 | 436 | 48 | 10809836852 | 90.39% |
| AX | 117 | 87 | 65 | 68864182574 | 83.13% |
| AZ | 380 | 3078 | 194 | 1967666905922 | 97.23% |
| BA | 292 | 3209 | 128 | 760855594110 | 97.70% |
| BB | 160 | 531 | 65 | 73508390455 | 97.65% |
| BD | 2456 | 8114 | 2029 | 3572229485220 | 95.49% |
| BE | 4850 | 55533 | 71783 | 8421704102083621 | 60.03% |
| BF | 110 | 1440 | 99 | 90806374973 | 99.15% |
| BG | 2090 | 16764 | 746 | 5154667408231 | 96.10% |
| BH | 253 | 3870 | 169 | 401320264779 | 98.59% |
| BI | 48 | 158 | 62 | 26163085877 | 92.85% |
| BJ | 71 | 612 | 77 | 56228029756 | 96.25% |
| BL | 20 | 6 | 47 | 324208932 | 94.28% |
| BM | 311 | 531 | 78 | 47319220800 | 95.89% |
| BN | 96 | 864 | 77 | 25912017211 | 97.56% |
| BO | 220 | 4717 | 158 | 167774011995 | 98.04% |
| BQ | 32 | 103 | 50 | 21664629036 | 90.88% |
| BR | 9233 | 325870 | 13238 | 37234247891749 | 99.30% |
| BS | 117 | 615 | 84 | 13262787399 | 99.74% |
| BT | 112 | 218 | 114 | 60743221843 | 99.55% |
| BV | 0 | 0 | 10 | 18022674 | 50.70% |
| BW | 152 | 692 | 85 | 78140231992 | 96.97% |
| BY | 300 | 7366 | 145 | 699981694127 | 99.33% |
| BZ | 242 | 482 | 177 | 233713893968 | 85.26% |
| CA | 14845 | 281069 | 10967 | 6501043243621 | 94.11% |
| CC | 2 | 0 | 22 | 120652051 | 56.09% |
| CD | 226 | 892 | 141 | 155152138589 | 98.56% |
| CF | 41 | 36 | 55 | 577196338 | 55.95% |
| CG | 63 | 573 | 70 | 36380176692 | 95.70% |
| CH | 5605 | 96084 | 54411 | 15878799624612159105 | 66.67% |
| CI | 334 | 6715 | 123 | 194272174375 | 98.75% |
| CK | 30 | 41 | 51 | 8780448039 | 98.89% |
| CL | 1478 | 40284 | 1024 | 1407798787383 | 97.05% |
| CM | 175 | 2428 | 108 | 68764067156 | 97.78% |
| CN | 6554 | 1349219 | 5428 | 294925226852253 | 99.91% |
| CO | 1895 | 69862 | 1317 | 1953010543304 | 98.29% |
| CR | 411 | 9132 | 244 | 461493409165 | 97.67% |
| CU | 78 | 1014 | 67 | 26252018229 | 99.27% |
| CV | 70 | 234 | 104 | 13477413178 | 98.06% |
| CW | 94 | 736 | 66 | 161116389666 | 68.51% |
| CX | 2 | 0 | 18 | 36831507 | 65.95% |
| CY | 650 | 4541 | 650 | 10527816938061 | 80.51% |
| CZ | 3062 | 34226 | 1475 | 13471716256004 | 97.23% |
| DE | 27693 | 529855 | 64461 | 108770354813301 | 97.92% |
| DJ | 39 | 229 | 71 | 9085715761 | 97.57% |
| DK | 3181 | 49403 | 1053 | 9356389697597 | 93.60% |
| DM | 59 | 64 | 54 | 43409736244 | 76.49% |
| DO | 544 | 7965 | 535 | 928004590290 | 98.93% |
| DZ | 187 | 18707 | 117 | 43387912596 | 94.39% |
| EC | 620 | 11182 | 489 | 1154028831534 | 98.94% |
| EE | 1106 | 6357 | 709 | 3173972983324 | 90.66% |
| EG | 482 | 94694 | 163 | 17730888467069 | 99.98% |
| EH | 2 | 0 | 24 | 287768850 | 47.78% |
| ER | 34 | 21 | 57 | 409424182 | 60.59% |
| ES | 9358 | 137662 | 3261 | 25692142628883 | 98.02% |
| ET | 62 | 1462 | 78 | 1117305114940 | 99.98% |
| FI | 3762 | 58138 | 34429 | 6721732545043 | 97.54% |
| FJ | 99 | 588 | 79 | 43175314487 | 97.25% |
| FK | 27 | 21 | 50 | 89784609 | 60.43% |
| FM | 46 | 41 | 66 | 8783726913 | 93.12% |
| FO | 40 | 178 | 58 | 107547395624 | 99.93% |
| FR | 22863 | 327844 | 6214 | 66639717378638 | 98.90% |
| GA | 70 | 1546 | 61 | 34769229105 | 93.87% |
| GB | 30922 | 503888 | 15586 | 97094451643785 | 95.77% |
| GD | 91 | 113 | 65 | 77687095862 | 99.96% |
| GE | 448 | 5075 | 289 | 1903298828362 | 94.10% |
| GF | 45 | 98 | 57 | 21749825828 | 98.20% |
| GG | 112 | 291 | 73 | 189219013153 | 81.27% |
| GH | 310 | 8887 | 177 | 1233647543910 | 99.65% |
| GI | 126 | 412 | 68 | 202139697444 | 90.10% |
| GL | 97 | 144 | 115 | 43164828010 | 99.76% |
| GM | 63 | 1094 | 66 | 39097423156 | 99.56% |
| GN | 62 | 166 | 80 | 47788200244 | 98.31% |
| GP | 212 | 444 | 66 | 47453262129 | 74.27% |
| GQ | 46 | 88 | 59 | 9032913207 | 98.05% |
| GR | 1204 | 23131 | 502 | 2485461141743 | 97.41% |
| GS | 3 | 2 | 22 | 237437203 | 48.48% |
| GT | 774 | 2733 | 193 | 326884938413 | 99.39% |
| GU | 162 | 965 | 76 | 34436088116 | 99.23% |
| GW | 35 | 32 | 62 | 9133577527 | 97.44% |
| GY | 47 | 284 | 69 | 34551431480 | 97.69% |
| HK | 9429 | 81862 | 5895 | 9129148261459 | 70.56% |
| HM | 1 | 2 | 12 | 34865426 | 52.82% |
| HN | 927 | 1969 | 213 | 421377606336 | 93.22% |
| HR | 619 | 9607 | 300 | 1617206228972 | 95.86% |
| HT | 66 | 609 | 70 | 25978012221 | 95.46% |
| HU | 1475 | 23204 | 643 | 3304939994785 | 97.42% |
| ID | 6384 | 82118 | 3613 | 3669143823615 | 99.20% |
| IE | 3983 | 70744 | 5223 | 4940134109505 | 89.07% |
| IL | 1952 | 34802 | 1830 | 4453089117943 | 95.78% |
| IM | 218 | 544 | 194 | 429804684065 | 88.23% |
| IN | 12026 | 190670 | 9771 | 29423465607833 | 97.02% |
| IO | 20 | 14 | 48 | 4552327457 | 97.10% |
| IQ | 603 | 3283 | 332 | 3286471414191 | 98.20% |
| IR | 1911 | 42866 | 756 | 17588591526709 | 97.09% |
| IS | 302 | 3630 | 283 | 1504111841871 | 98.87% |
| IT | 8439 | 219122 | 3573 | 47187222727615 | 98.91% |
| JE | 113 | 639 | 67 | 253561012772 | 91.59% |
| JM | 224 | 1093 | 93 | 1121380681309 | 100.00% |
| JO | 282 | 2564 | 155 | 1091409871448 | 98.19% |
| JP | 9648 | 774115 | 4849 | 51418763765825 | 90.07% |
| KE | 758 | 24833 | 359 | 667084621496 | 98.65% |
| KG | 227 | 1186 | 133 | 563117228885 | 96.28% |
| KH | 463 | 1976 | 191 | 297187820133 | 95.15% |
| KI | 67 | 29 | 89 | 13081313620 | 95.37% |
| KM | 50 | 43 | 55 | 9167044914 | 97.23% |
| KN | 63 | 114 | 58 | 34836119872 | 60.30% |
| KP | 4 | 7 | 262 | 530776594 | 55.10% |
| KR | 2839 | 452498 | 1511 | 22752593521079 | 99.64% |
| KW | 253 | 7501 | 129 | 872396239179 | 99.90% |
| KY | 126 | 429 | 62 | 17675452967 | 99.36% |
| KZ | 1229 | 13881 | 416 | 2415165950849 | 98.16% |
| LA | 131 | 410 | 102 | 86429205046 | 95.17% |
| LB | 362 | 2318 | 209 | 2981241242178 | 99.08% |
| LC | 73 | 155 | 54 | 81963516208 | 97.33% |
| LI | 135 | 427 | 101 | 391173378868 | 91.89% |
| LK | 207 | 2292 | 131 | 73448390986 | 95.86% |
| LR | 63 | 170 | 72 | 26232246583 | 97.38% |
| LS | 56 | 496 | 61 | 26162565428 | 98.51% |
| LT | 1379 | 12345 | 867 | 6279999558367 | 76.37% |
| LU | 1229 | 6948 | 498 | 2277119100275 | 94.14% |
| LV | 1089 | 8161 | 451 | 1530000255861 | 93.71% |
| LY | 130 | 1904 | 82 | 99412037959 | 92.16% |
| MA | 287 | 48059 | 159 | 1164847874703 | 99.96% |
| MC | 82 | 277 | 64 | 146321573169 | 98.69% |
| MD | 587 | 4958 | 1071 | 6584534582295 | 72.97% |
| ME | 111 | 791 | 92 | 395497329214 | 93.82% |
| MF | 31 | 30 | 45 | 4652532004 | 99.60% |
| MG | 76 | 2299 | 76 | 22220722749 | 98.61% |
| MH | 43 | 128 | 49 | 73087910197 | 55.72% |
| MK | 231 | 2766 | 151 | 1069809025637 | 98.67% |
| ML | 72 | 378 | 60 | 21917794618 | 98.11% |
| MM | 354 | 1099 | 250 | 352694882915 | 97.29% |
| MN | 211 | 733 | 158 | 133729960270 | 99.63% |
| MO | 86 | 1403 | 121 | 38872761123 | 98.42% |
| MP | 48 | 32 | 44 | 4350673188 | 99.57% |
| MQ | 73 | 728 | 66 | 112933601568 | 88.87% |
| MR | 54 | 216 | 58 | 26162364725 | 99.43% |
| MS | 21 | 8 | 47 | 189858088 | 48.16% |
| MT | 241 | 2658 | 128 | 466260066875 | 96.94% |
| MU | 1787 | 26613 | 7829 | 4455234700857656 | 100.00% |
| MV | 115 | 385 | 77 | 17741840694 | 98.61% |
| MW | 97 | 2190 | 96 | 1241843192122 | 99.89% |
| MX | 3157 | 118799 | 1685 | 3526319712280 | 97.64% |
| MY | 2023 | 30436 | 960 | 896378243225 | 95.49% |
| MZ | 196 | 1826 | 88 | 69180674375 | 96.54% |
| NA | 226 | 1876 | 68 | 29215183167 | 98.24% |
| NC | 67 | 571 | 78 | 47438561840 | 99.79% |
| NE | 50 | 208 | 59 | 21633761586 | 98.59% |
| NF | 26 | 3 | 45 | 189923621 | 47.19% |
| NG | 836 | 12794 | 417 | 632668292954 | 96.07% |
| NI | 137 | 1700 | 81 | 94716691014 | 98.47% |
| NL | 18256 | 213284 | 162284 | 8407736758896612 | 99.94% |
| NO | 2572 | 63920 | 1112 | 10936663884735 | 95.33% |
| NP | 405 | 2368 | 284 | 430168826212 | 92.22% |
| NR | 30 | 47 | 50 | 8781955381 | 98.88% |
| NU | 23 | 11 | 52 | 4502914341 | 97.63% |
| NZ | 1929 | 29012 | 1116 | 1337206541154 | 96.09% |
| OM | 278 | 4308 | 108 | 550420365631 | 99.51% |
| PA | 486 | 7162 | 420 | 491796571761 | 88.87% |
| PE | 879 | 14153 | 668 | 1068643155645 | 97.05% |
| PF | 76 | 309 | 109 | 21705001297 | 99.47% |
| PG | 164 | 327 | 101 | 73259352654 | 91.37% |
| PH | 1858 | 26027 | 951 | 1029560760734 | 97.11% |
| PK | 1227 | 23352 | 591 | 924354564545 | 95.96% |
| PL | 7544 | 82032 | 2402 | 28263502382847 | 96.94% |
| PM | 35 | 23 | 49 | 4484957733 | 97.82% |
| PN | 3 | 2 | 16 | 52756755 | 66.45% |
| PR | 523 | 4437 | 743 | 223265370521 | 97.21% |
| PS | 288 | 3319 | 127 | 979536384317 | 95.72% |
| PT | 1439 | 26828 | 647 | 3064828422215 | 95.97% |
| PW | 36 | 31 | 48 | 8646885685 | 99.77% |
| PY | 331 | 5032 | 210 | 451219571341 | 97.83% |
| QA | 246 | 5065 | 168 | 460232077658 | 99.46% |
| RE | 94 | 1652 | 64 | 129308819755 | 84.64% |
| RO | 3745 | 29211 | 1291 | 5580661709080 | 95.84% |
| RS | 870 | 9095 | 656 | 1856073067634 | 98.43% |
| RU | 12321 | 180379 | 4623 | 46629021541215 | 95.40% |
| RW | 92 | 1188 | 84 | 1147200164151 | 99.83% |
| SA | 1216 | 46106 | 619 | 2715990032432 | 98.37% |
| SB | 43 | 62 | 59 | 17253925428 | 97.08% |
| SC | 3028 | 14995 | 4115 | 14823811547956 | 57.23% |
| SD | 88 | 7415 | 73 | 69232361787 | 97.24% |
| SE | 8392 | 125065 | 4493 | 30495016131379 | 87.26% |
| SG | 7814 | 112434 | 4385 | 148579403561309 | 97.63% |
| SH | 3 | 2 | 27 | 120193309 | 58.05% |
| SI | 884 | 10374 | 537 | 2284947961231 | 92.65% |
| SJ | 4 | 2 | 24 | 238486803 | 50.11% |
| SK | 880 | 10948 | 396 | 2900495548413 | 97.34% |
| SL | 74 | 360 | 72 | 26246513978 | 98.79% |
| SM | 55 | 161 | 70 | 197813862705 | 93.27% |
| SN | 89 | 1614 | 99 | 26448712013 | 98.86% |
| SO | 54 | 213 | 78 | 47738671670 | 99.58% |
| SR | 43 | 336 | 73 | 17391616572 | 99.47% |
| SS | 75 | 128 | 75 | 30625849657 | 99.14% |
| ST | 32 | 50 | 56 | 4586930485 | 96.86% |
| SV | 309 | 2751 | 144 | 210718585477 | 97.11% |
| SX | 45 | 146 | 53 | 30120804642 | 96.37% |
| SY | 271 | 5143 | 101 | 876791796027 | 98.67% |
| SZ | 58 | 227 | 65 | 30172074304 | 98.65% |
| TC | 35 | 50 | 46 | 326238756 | 94.13% |
| TD | 61 | 117 | 64 | 13093241142 | 99.34% |
| TF | 2 | 0 | 17 | 170262803 | 44.86% |
| TG | 56 | 1396 | 80 | 26196136245 | 98.34% |
| TH | 1844 | 40454 | 958 | 587491961306 | 94.79% |
| TJ | 130 | 380 | 93 | 670810374973 | 96.76% |
| TK | 21 | 14 | 52 | 4350935336 | 99.57% |
| TL | 80 | 119 | 70 | 43426644279 | 97.53% |
| TM | 52 | 120 | 56 | 68844716601 | 99.92% |
| TN | 125 | 30790 | 93 | 112249614908 | 99.77% |
| TO | 37 | 48 | 53 | 17371103539 | 99.43% |
| TR | 2973 | 67613 | 1649 | 24023035308168 | 89.39% |
| TT | 133 | 2196 | 92 | 1177065767498 | 77.91% |
| TV | 29 | 38 | 60 | 10826023219 | 87.98% |
| TW | 1889 | 144860 | 1436 | 11485421826710 | 98.52% |
| TZ | 315 | 4424 | 173 | 266872471928 | 98.34% |
| UA | 4500 | 34882 | 1725 | 11060589283638 | 94.98% |
| UG | 202 | 5666 | 104 | 125048217917 | 97.26% |
| UM | 3 | 2 | 20 | 187106579 | 47.32% |
| US | 98953 | 5983277 | 155164 | 4837647341043396 | 99.77% |
| UY | 186 | 10325 | 120 | 177248813367 | 99.57% |
| UZ | 421 | 1597 | 153 | 1083206751044 | 99.48% |
| VA | 23 | 43 | 52 | 103155434031 | 99.97% |
| VC | 63 | 86 | 47 | 69044863281 | 99.97% |
| VE | 714 | 29812 | 478 | 2252981809812 | 99.45% |
| VG | 237 | 2945 | 343 | 770134876452 | 64.09% |
| VI | 72 | 537 | 114 | 30126761271 | 94.35% |
| VN | 1970 | 65227 | 1340 | 1902154082831 | 98.01% |
| VU | 52 | 64 | 224 | 51781501748 | 87.57% |
| WF | 19 | 16 | 44 | 4350608676 | 99.55% |
| WS | 49 | 88 | 63 | 21683044919 | 91.59% |
| XK | 7 | 6 | 3 | 65538 | 33.33% |
| YE | 97 | 940 | 86 | 206625964356 | 99.92% |
| YT | 30 | 15 | 51 | 181535012 | 58.27% |
| ZA | 3790 | 114023 | 1235 | 23390913508637 | 99.41% |
| ZM | 144 | 6393 | 90 | 43460285011 | 94.98% |
| ZW | 188 | 690 | 104 | 65341509992 | 98.61% |

<!-- GEOIP-STATISTICS:END -->
