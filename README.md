# IP2Country GeoIP Generator

A Go program that combines public IP-to-country feeds into compact country databases and country-specific CSV files. It produces Xray/Mihomo V2Fly protobuf GeoIP, current sing-box rule sets, and a MaxMind DB country database.

The input inventory, parser settings, enabled state, vote groups, and source order live in [`config/sources.json`](config/sources.json). All supplied country feeds are enabled by default. The IANA feed supplies country-tagged allocation prefixes; built-in special-use ranges are excluded. Public Cloudflare and hosting ranges remain eligible.

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

Each source casts at most one vote per normalized /24 or /64 cell. When one source assigns competing countries within a cell, the country covering the most addresses wins; the earliest input row breaks a tie. Across sources, ties follow the order in the JSON config. Confidence is winning votes divided by all votes. Adjacent cells with the same winning country merge; their confidence is averaged by represented address space.

Range processing uses sorted boundaries instead of expanding every /24 or /64. It therefore handles broad IPv6 networks without materializing billions of rows. Preprocessing is O(n log n); final sorted-prefix lookups are O(log n).

## Automated updates

The GeoIP workflow runs at 00:00 UTC each day and supports `workflow_dispatch`. It publishes one dated GitHub release per UTC day, reuses that release on reruns, and updates this README's generated statistics section. GitHub Actions caches source files between runs. The quality workflow runs formatting, tests, vet, and build checks on pushes and pull requests.

<!-- GEOIP-STATISTICS:START -->
## Generated data status

Last run: `2026-10-20T12:44:34Z`  
Input rows: **30719101** · Published blocks: **1077985**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | Bytes | Details |
|---|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | fresh | 597286 | 0 | 0 | 11032548 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv4.csv |
| DB-IP via sapics IPv6 | fresh | 0 | 785668 | 0 | 20177287 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/dbip-country/dbip-country-ipv6.csv |
| DB-IP via tdulcet IPv4 | fresh | 26086 | 0 | 336022 | 7580021 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv4.tsv?inline=false |
| DB-IP via tdulcet IPv6 | fresh | 0 | 785668 | 0 | 24060852 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/dbip-country/ipv6.tsv?inline=false |
| GeoLite2 via sapics IPv4 | fresh | 562225 | 0 | 0 | 11070001 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv4.csv |
| GeoLite2 via sapics IPv6 | fresh | 0 | 500017 | 0 | 14970875 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/geolite2-country/geolite2-country-ipv6.csv |
| GeoLite2 via tdulcet IPv4 | fresh | 35561 | 0 | 529190 | 11825975 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv4.tsv?inline=false |
| GeoLite2 via tdulcet IPv6 | fresh | 0 | 500090 | 153 | 34516461 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geolite2-country/ipv6.tsv?inline=false |
| Server-Country via sapics IPv4 | fresh | 398706 | 0 | 0 | 8688499 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 471750 | 1 | 15531894 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 15644 | 0 | 267421 | 5930350 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false: HTTP 503 Service Unavailable |
| User-Country IPv4 | fresh | 415649 | 0 | 0 | 9066810 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 563998 | 1 | 16757331 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | cached | 28507 | 0 | 430722 | 9616617 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false: read tcp 10.0.0.92:62611->104.18.29.118:443: wsarecv: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond. |
| IP to ASN Country IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false: Get "https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false": dial tcp 104.18.29.118:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond. |
| IPINFO via tdulcet IPv4 | fresh | 31777 | 0 | 588831 | 12997637 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2707304 | 0 | 85118814 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPINFO Lite CSV | fresh | 1341028 | 2936383 | 0 | 208702865 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 19976 | 0 | 271288 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IPLocate | fresh | 785303 | 1368186 | 0 | 17235320 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177093 | 69680 | 0 | 3393343 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 287188 | 480983 | 97783 | 40299556 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| Clang CN | fresh | 4318 | 1675 | 0 | 97591 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3879 | 0 | 0 | 82305 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327605 | 0 | 0 | 537559 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 808852 | 422317 | 411004 | 18232363 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262213 | 70936 | 10 | 34563823 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4794144 | 0 | 127222 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 769175 | 23255 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| PikoNetwork IPv4 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | source produced no country/IP records |
| PikoNetwork IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | source produced no country/IP records |
| IP2Region | fresh | 802825 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| V2ray GeoIP | fresh | 597286 | 785617 | 24 | 23258601 | https://github.com/v2fly/geoip/releases/latest/download/geoip.dat |
| Meta Rules Dat | fresh | 512893 | 413486 | 10664 | 7647337 | https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/country.mmdb |
| Loyalsoldier GeoIP Enhanced | fresh | 557938 | 492270 | 10655 | 16587463 | https://raw.githubusercontent.com/Loyalsoldier/geoip/release/geoip.dat |

### Final country coverage

| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |
|---|---:|---:|---:|---:|---:|
| AD | 51 | 270.0000 | 117 | 43826350375.0000 | 89.79% |
| AE | 2306 | 25026.0000 | 1746 | 13975022325450.0000 | 89.07% |
| AF | 147 | 734.0000 | 124 | 121077236077.0000 | 96.69% |
| AG | 103 | 556.0000 | 141 | 39401932841.0000 | 65.59% |
| AI | 12 | 40.0000 | 48 | 4787472417.0000 | 69.40% |
| AL | 375 | 1775.0000 | 1027 | 2577747754826.0000 | 95.09% |
| AM | 271 | 2336.0000 | 292 | 1297430345543.0000 | 93.42% |
| AO | 220 | 5157.0000 | 151 | 177326297667.0000 | 90.50% |
| AQ | 5 | 7.0000 | 314 | 4403087382.0000 | 60.94% |
| AR | 2369 | 76407.0000 | 1857 | 23553538323731.0000 | 98.93% |
| AS | 6 | 26.0000 | 162 | 8787003444.0000 | 98.84% |
| AT | 2666 | 44352.0000 | 2599 | 10423699632333.0000 | 97.22% |
| AU | 9241 | 213098.0000 | 50151 | 2305941891674537063.0000 | 100.00% |
| AW | 19 | 436.0000 | 48 | 10809853988.0000 | 90.35% |
| AX | 24 | 90.0000 | 85 | 68865296434.0000 | 78.14% |
| AZ | 295 | 3082.0000 | 184 | 1933256049202.0000 | 97.23% |
| BA | 225 | 3205.0000 | 146 | 760822861428.0000 | 97.06% |
| BB | 104 | 555.0000 | 70 | 72703216167.0000 | 98.10% |
| BD | 2292 | 8117.0000 | 2075 | 3564500971931.0000 | 95.21% |
| BE | 1759 | 55610.0000 | 55684 | 8421861300947973.0000 | 57.18% |
| BF | 85 | 1445.0000 | 97 | 90772887852.0000 | 99.72% |
| BG | 1694 | 16782.0000 | 893 | 5262422388359.0000 | 94.81% |
| BH | 144 | 3856.0000 | 179 | 401303028539.0000 | 98.18% |
| BI | 26 | 162.0000 | 58 | 26129531428.0000 | 92.43% |
| BJ | 54 | 625.0000 | 90 | 56227768620.0000 | 96.45% |
| BL | 5 | 6.0000 | 46 | 324208676.0000 | 94.90% |
| BM | 76 | 482.0000 | 77 | 47319157040.0000 | 95.63% |
| BN | 50 | 862.0000 | 72 | 25878528043.0000 | 96.92% |
| BO | 162 | 4722.0000 | 162 | 167744242251.0000 | 97.39% |
| BQ | 11 | 102.0000 | 50 | 21664565292.0000 | 90.73% |
| BR | 7654 | 326144.0000 | 12878 | 37156741607821.0000 | 99.30% |
| BS | 58 | 583.0000 | 85 | 13263100465.0000 | 99.77% |
| BT | 67 | 222.0000 | 112 | 60709732931.0000 | 99.58% |
| BV | 0 |  | 5 | 17891344.0000 | 48.46% |
| BW | 81 | 691.0000 | 84 | 78124044584.0000 | 97.17% |
| BY | 235 | 7375.0000 | 234 | 699954611122.0000 | 99.29% |
| BZ | 90 | 482.0000 | 147 | 199069403705.0000 | 89.08% |
| CA | 9595 | 280879.0000 | 9237 | 6459390813268.0000 | 94.15% |
| CC | 0 |  | 19 | 87098386.0000 | 57.48% |
| CD | 133 | 898.0000 | 147 | 155169440076.0000 | 98.94% |
| CF | 13 | 37.0000 | 64 | 561795105.0000 | 55.30% |
| CG | 40 | 578.0000 | 69 | 36380700708.0000 | 97.66% |
| CH | 3821 | 96196.0000 | 53636 | 15878799865009072931.0000 | 75.02% |
| CI | 286 | 6711.0000 | 112 | 194255284520.0000 | 98.86% |
| CK | 5 | 40.0000 | 52 | 8780514343.0000 | 98.87% |
| CL | 1125 | 40318.0000 | 1072 | 1406015830344.0000 | 96.51% |
| CM | 113 | 2432.0000 | 120 | 68227459139.0000 | 98.12% |
| CN | 5971 | 1349594.0000 | 6954 | 294839046714244.0000 | 99.92% |
| CO | 1220 | 69911.0000 | 1452 | 1954414619528.0000 | 98.03% |
| CR | 311 | 9145.0000 | 245 | 461497309533.0000 | 97.42% |
| CU | 31 | 1019.0000 | 69 | 26269450533.0000 | 99.25% |
| CV | 41 | 236.0000 | 104 | 13461162026.0000 | 98.12% |
| CW | 64 | 738.0000 | 68 | 62334502946.0000 | 93.08% |
| CX | 0 |  | 19 | 36898834.0000 | 67.65% |
| CY | 517 | 4563.0000 | 686 | 7970150086980.0000 | 87.20% |
| CZ | 2206 | 34260.0000 | 2066 | 13205745827811.0000 | 97.99% |
| DE | 16885 | 534835.0000 | 47494 | 111144885393284.0000 | 96.96% |
| DJ | 21 | 233.0000 | 69 | 9084733473.0000 | 97.54% |
| DK | 1724 | 49408.0000 | 1346 | 9550892422988.0000 | 93.25% |
| DM | 31 | 63.0000 | 72 | 43409607204.0000 | 74.78% |
| DO | 465 | 7982.0000 | 559 | 923727826347.0000 | 99.01% |
| DZ | 100 | 18712.0000 | 116 | 43354884228.0000 | 93.73% |
| EC | 486 | 11182.0000 | 543 | 1154000898867.0000 | 98.70% |
| EE | 733 | 6381.0000 | 924 | 3018274084121.0000 | 92.43% |
| EG | 328 | 94691.0000 | 168 | 17730858991724.0000 | 99.98% |
| EH | 0 |  | 24 | 271452178.0000 | 45.37% |
| ER | 9 | 25.0000 | 55 | 376460325.0000 | 60.92% |
| ES | 5536 | 137746.0000 | 2879 | 25704006541533.0000 | 97.81% |
| ET | 34 | 1466.0000 | 90 | 1117288142891.0000 | 99.98% |
| FI | 2180 | 58122.0000 | 44830 | 6813102016021.0000 | 96.78% |
| FJ | 58 | 592.0000 | 78 | 43159062311.0000 | 96.44% |
| FK | 6 | 21.0000 | 49 | 56297504.0000 | 70.27% |
| FM | 16 | 46.0000 | 63 | 8783726640.0000 | 90.71% |
| FO | 16 | 179.0000 | 60 | 107514628391.0000 | 99.95% |
| FR | 10106 | 326550.0000 | 6662 | 66957740701363.0000 | 98.65% |
| GA | 48 | 1551.0000 | 62 | 35272726815.0000 | 93.65% |
| GB | 17386 | 504060.0000 | 21464 | 105712010613131.0000 | 92.35% |
| GD | 50 | 98.0000 | 68 | 77955598886.0000 | 99.81% |
| GE | 382 | 5075.0000 | 259 | 1834531525955.0000 | 95.09% |
| GF | 31 | 132.0000 | 56 | 27084980260.0000 | 90.99% |
| GG | 61 | 288.0000 | 63 | 146253089058.0000 | 88.76% |
| GH | 228 | 8892.0000 | 180 | 1228280999510.0000 | 99.95% |
| GI | 75 | 411.0000 | 82 | 197822711844.0000 | 89.87% |
| GL | 35 | 145.0000 | 115 | 43164893290.0000 | 99.76% |
| GM | 42 | 1098.0000 | 65 | 39063935268.0000 | 99.60% |
| GN | 40 | 171.0000 | 78 | 47754582052.0000 | 97.94% |
| GP | 112 | 641.0000 | 67 | 47453263921.0000 | 73.11% |
| GQ | 28 | 93.0000 | 58 | 8999491623.0000 | 98.19% |
| GR | 792 | 23128.0000 | 438 | 2558516134768.0000 | 96.14% |
| GS | 1 | 2.0000 | 19 | 203948050.0000 | 47.40% |
| GT | 295 | 2739.0000 | 204 | 326888446066.0000 | 99.51% |
| GU | 61 | 954.0000 | 81 | 34453192757.0000 | 99.91% |
| GW | 14 | 38.0000 | 62 | 9133643815.0000 | 97.39% |
| GY | 26 | 287.0000 | 67 | 34517878824.0000 | 96.26% |
| HK | 6880 | 80786.0000 | 5166 | 6512641575732.0000 | 76.97% |
| HM | 1 | 2.0000 | 8 | 35782672.0000 | 52.13% |
| HN | 267 | 1904.0000 | 228 | 421378427505.0000 | 92.13% |
| HR | 412 | 9608.0000 | 331 | 1617162313396.0000 | 95.75% |
| HT | 36 | 616.0000 | 88 | 25995643440.0000 | 94.05% |
| HU | 942 | 23230.0000 | 947 | 3305807195966.0000 | 97.23% |
| ID | 5868 | 81994.0000 | 4145 | 3779024092039.0000 | 97.82% |
| IE | 2136 | 70813.0000 | 5786 | 5011555202097.0000 | 88.37% |
| IL | 1485 | 34916.0000 | 1923 | 4415080156081.0000 | 95.93% |
| IM | 109 | 545.0000 | 152 | 399707736614.0000 | 90.73% |
| IN | 9599 | 190485.0000 | 7672 | 29376198162426.0000 | 96.98% |
| IO | 3 | 14.0000 | 49 | 4518856737.0000 | 97.40% |
| IQ | 404 | 3293.0000 | 335 | 3337977860255.0000 | 97.44% |
| IR | 1872 | 42858.0000 | 770 | 17091146875428.0000 | 98.11% |
| IS | 230 | 3634.0000 | 365 | 1570713460292.0000 | 96.56% |
| IT | 5338 | 220172.0000 | 3590 | 47116804000177.0000 | 98.89% |
| JE | 63 | 639.0000 | 84 | 219200227619.0000 | 96.99% |
| JM | 96 | 892.0000 | 103 | 1121397918287.0000 | 100.00% |
| JO | 183 | 2565.0000 | 158 | 1121446152267.0000 | 97.11% |
| JP | 6810 | 774234.0000 | 6008 | 51658320868284.0000 | 89.01% |
| KE | 550 | 24882.0000 | 462 | 664654612134.0000 | 98.94% |
| KG | 189 | 1184.0000 | 131 | 528725728849.0000 | 99.29% |
| KH | 377 | 1973.0000 | 206 | 292862444118.0000 | 95.62% |
| KI | 21 | 34.0000 | 87 | 13081315394.0000 | 94.05% |
| KM | 18 | 46.0000 | 51 | 9133490210.0000 | 97.35% |
| KN | 35 | 110.0000 | 59 | 11213801520.0000 | 67.78% |
| KP | 4 | 8.0000 | 265 | 497746178.0000 | 54.00% |
| KR | 2136 | 452544.0000 | 1420 | 22742752268565.0000 | 99.67% |
| KW | 179 | 7507.0000 | 145 | 872666230846.0000 | 99.83% |
| KY | 57 | 407.0000 | 62 | 17675520293.0000 | 99.36% |
| KZ | 1053 | 13861.0000 | 655 | 2501582131386.0000 | 96.17% |
| LA | 91 | 415.0000 | 105 | 103574929702.0000 | 87.72% |
| LB | 301 | 2317.0000 | 209 | 2981207884338.0000 | 98.80% |
| LC | 46 | 148.0000 | 58 | 81695082784.0000 | 96.53% |
| LI | 90 | 434.0000 | 118 | 425498581542.0000 | 87.67% |
| LK | 122 | 2292.0000 | 137 | 73418965306.0000 | 95.71% |
| LR | 37 | 175.0000 | 85 | 26215209255.0000 | 97.52% |
| LS | 33 | 504.0000 | 58 | 26129076260.0000 | 99.47% |
| LT | 1130 | 12345.0000 | 959 | 5335842905818.0000 | 79.72% |
| LU | 593 | 6980.0000 | 673 | 2156917595483.0000 | 96.40% |
| LV | 883 | 8152.0000 | 490 | 1508755323260.0000 | 94.00% |
| LY | 99 | 1909.0000 | 79 | 99378548791.0000 | 89.65% |
| MA | 149 | 48065.0000 | 172 | 1165701681535.0000 | 99.93% |
| MC | 46 | 276.0000 | 75 | 146308466723.0000 | 98.67% |
| MD | 532 | 4979.0000 | 973 | 6498590150170.0000 | 69.90% |
| ME | 81 | 799.0000 | 89 | 343924170030.0000 | 99.95% |
| MF | 11 | 29.0000 | 47 | 4652599332.0000 | 99.64% |
| MG | 47 | 2304.0000 | 73 | 22187233836.0000 | 98.83% |
| MH | 18 | 130.0000 | 45 | 4368498724.0000 | 99.51% |
| MK | 146 | 2766.0000 | 178 | 1108447448424.0000 | 96.79% |
| ML | 42 | 380.0000 | 57 | 21884305450.0000 | 99.27% |
| MM | 314 | 1090.0000 | 269 | 352664621399.0000 | 97.04% |
| MN | 164 | 740.0000 | 157 | 133697486910.0000 | 99.65% |
| MO | 61 | 1384.0000 | 355 | 38867910703.0000 | 97.36% |
| MP | 11 | 39.0000 | 51 | 4351000615.0000 | 99.62% |
| MQ | 118 | 867.0000 | 68 | 112950380576.0000 | 87.57% |
| MR | 28 | 216.0000 | 57 | 26145589285.0000 | 99.47% |
| MS | 3 | 7.0000 | 47 | 189859879.0000 | 47.64% |
| MT | 158 | 2662.0000 | 150 | 466226694441.0000 | 96.55% |
| MU | 1441 | 26444.0000 | 7846 | 4455230376759602.0000 | 100.00% |
| MV | 61 | 385.0000 | 96 | 17726177323.0000 | 98.63% |
| MW | 67 | 2194.0000 | 112 | 1241826365482.0000 | 99.86% |
| MX | 2184 | 118801.0000 | 1712 | 3532711922082.0000 | 97.53% |
| MY | 1555 | 30241.0000 | 1189 | 902297808766.0000 | 94.79% |
| MZ | 91 | 1854.0000 | 87 | 69163028535.0000 | 97.09% |
| NA | 105 | 1877.0000 | 65 | 29198455081.0000 | 98.58% |
| NC | 49 | 582.0000 | 77 | 47438497072.0000 | 99.79% |
| NE | 30 | 211.0000 | 55 | 21616984097.0000 | 99.74% |
| NF | 2 | 3.0000 | 44 | 173148196.0000 | 45.63% |
| NG | 532 | 12807.0000 | 469 | 633226069840.0000 | 96.06% |
| NI | 88 | 1687.0000 | 79 | 94699915318.0000 | 98.37% |
| NL | 11867 | 212274.0000 | 145984 | 8407898916064919.0000 | 99.94% |
| NO | 1741 | 63962.0000 | 1436 | 10877302293162.0000 | 95.71% |
| NP | 328 | 2370.0000 | 283 | 425873989716.0000 | 92.28% |
| NR | 15 | 49.0000 | 46 | 8780842020.0000 | 98.87% |
| NU | 4 | 11.0000 | 48 | 4485088292.0000 | 97.80% |
| NZ | 1678 | 29090.0000 | 1463 | 1338531493931.0000 | 95.75% |
| OM | 191 | 4312.0000 | 136 | 586525917239.0000 | 96.32% |
| PA | 274 | 7068.0000 | 379 | 461179775332.0000 | 90.17% |
| PE | 599 | 14096.0000 | 735 | 1064037170631.0000 | 96.45% |
| PF | 27 | 308.0000 | 108 | 21705001041.0000 | 99.47% |
| PG | 104 | 330.0000 | 112 | 73263087935.0000 | 90.70% |
| PH | 1474 | 25893.0000 | 1020 | 1037855178132.0000 | 96.47% |
| PK | 1007 | 23337.0000 | 644 | 898562986148.0000 | 97.05% |
| PL | 5532 | 82070.0000 | 3409 | 28452952930129.0000 | 96.64% |
| PM | 5 | 23.0000 | 49 | 4485088548.0000 | 97.80% |
| PN | 1 | 2.0000 | 15 | 52757522.0000 | 69.89% |
| PR | 385 | 4454.0000 | 221 | 223350012025.0000 | 97.85% |
| PS | 265 | 3318.0000 | 230 | 983827437883.0000 | 94.32% |
| PT | 691 | 26828.0000 | 809 | 3296762935594.0000 | 92.53% |
| PW | 8 | 31.0000 | 45 | 8646885412.0000 | 99.79% |
| PY | 265 | 5028.0000 | 214 | 451189555836.0000 | 97.16% |
| QA | 173 | 5059.0000 | 281 | 521252404144.0000 | 94.21% |
| RE | 68 | 1720.0000 | 59 | 129275248678.0000 | 84.57% |
| RO | 3043 | 29302.0000 | 1615 | 5748391576626.0000 | 93.83% |
| RS | 699 | 9102.0000 | 578 | 1856997797703.0000 | 98.35% |
| RU | 11456 | 180338.0000 | 9255 | 49448590027609.0000 | 93.08% |
| RW | 67 | 1224.0000 | 82 | 1147183386919.0000 | 99.88% |
| SA | 911 | 45955.0000 | 587 | 2741264339582.0000 | 97.56% |
| SB | 22 | 68.0000 | 59 | 17270703396.0000 | 97.19% |
| SC | 2472 | 13909.0000 | 3908 | 8496212918819.0000 | 56.17% |
| SD | 53 | 7419.0000 | 70 | 69198807339.0000 | 97.10% |
| SE | 4404 | 125563.0000 | 4796 | 30577974224236.0000 | 85.13% |
| SG | 5222 | 112093.0000 | 4541 | 149417616999348.0000 | 96.99% |
| SH | 1 | 2.0000 | 38 | 102696987.0000 | 59.00% |
| SI | 665 | 10398.0000 | 624 | 2254909755522.0000 | 92.64% |
| SJ | 1 | 2.0000 | 20 | 204867602.0000 | 49.26% |
| SK | 588 | 10970.0000 | 601 | 2905236107008.0000 | 97.17% |
| SL | 49 | 365.0000 | 69 | 26212960298.0000 | 98.43% |
| SM | 36 | 166.0000 | 73 | 197812946979.0000 | 92.82% |
| SN | 54 | 1618.0000 | 101 | 26969872445.0000 | 97.92% |
| SO | 44 | 219.0000 | 81 | 47705346342.0000 | 99.61% |
| SR | 26 | 339.0000 | 69 | 17358061868.0000 | 99.58% |
| SS | 50 | 132.0000 | 73 | 30592491560.0000 | 99.23% |
| ST | 12 | 56.0000 | 60 | 4553771045.0000 | 97.13% |
| SV | 169 | 2727.0000 | 139 | 210718640471.0000 | 97.22% |
| SX | 13 | 138.0000 | 54 | 30120806434.0000 | 95.19% |
| SY | 218 | 5143.0000 | 98 | 876758306859.0000 | 98.36% |
| SZ | 36 | 231.0000 | 65 | 29870149680.0000 | 99.91% |
| TC | 11 | 49.0000 | 51 | 1132661028.0000 | 64.77% |
| TD | 34 | 121.0000 | 61 | 13059688485.0000 | 99.48% |
| TF | 0 |  | 13 | 170196994.0000 | 42.99% |
| TG | 32 | 1402.0000 | 80 | 26162649381.0000 | 99.42% |
| TH | 1533 | 40401.0000 | 1101 | 593564379373.0000 | 93.73% |
| TJ | 96 | 382.0000 | 102 | 670792485421.0000 | 96.24% |
| TK | 4 | 14.0000 | 479 | 4351067970.0000 | 99.62% |
| TL | 43 | 118.0000 | 69 | 43426709543.0000 | 96.68% |
| TM | 28 | 125.0000 | 53 | 68810245416.0000 | 99.94% |
| TN | 61 | 30793.0000 | 94 | 112216064556.0000 | 99.78% |
| TO | 18 | 53.0000 | 51 | 17371103267.0000 | 99.43% |
| TR | 2513 | 67656.0000 | 2093 | 23614821145554.0000 | 89.36% |
| TT | 92 | 2211.0000 | 95 | 1177082579514.0000 | 70.55% |
| TV | 9 | 41.0000 | 60 | 12388271139.0000 | 81.68% |
| TW | 1403 | 144861.0000 | 1880 | 11502772294398.0000 | 98.21% |
| TZ | 246 | 4449.0000 | 174 | 266855793256.0000 | 98.53% |
| UA | 4197 | 35033.0000 | 2095 | 11622690497793.0000 | 92.81% |
| UG | 153 | 5714.0000 | 119 | 125051445547.0000 | 97.51% |
| UM | 1 | 2.0000 | 35 | 188090387.0000 | 46.76% |
| US | 63351 | 5980037.0000 | 168521 | 4837326565726392.0000 | 99.77% |
| UY | 120 | 10324.0000 | 123 | 177218584615.0000 | 99.87% |
| UZ | 348 | 1600.0000 | 191 | 1083194677824.0000 | 99.36% |
| VA | 5 | 43.0000 | 56 | 103155648799.0000 | 99.97% |
| VC | 39 | 84.0000 | 48 | 69044930593.0000 | 99.98% |
| VE | 536 | 29797.0000 | 476 | 2252950352260.0000 | 99.32% |
| VG | 181 | 2944.0000 | 209 | 499551282217.0000 | 68.28% |
| VI | 33 | 523.0000 | 113 | 30144620594.0000 | 94.63% |
| VN | 1681 | 65190.0000 | 1435 | 1902433616757.0000 | 97.83% |
| VU | 29 | 68.0000 | 247 | 51781372708.0000 | 85.56% |
| WF | 4 | 16.0000 | 44 | 4350609444.0000 | 99.60% |
| WS | 28 | 92.0000 | 61 | 21683044647.0000 | 91.03% |
| XK | 74 | 250.0000 | 11 | 73014575110.0000 | 52.88% |
| YE | 51 | 949.0000 | 85 | 206642741300.0000 | 99.91% |
| YT | 12 | 20.0000 | 48 | 147980324.0000 | 61.09% |
| ZA | 3293 | 114120.0000 | 1546 | 23429783044984.0000 | 99.39% |
| ZM | 103 | 6395.0000 | 88 | 43712008514.0000 | 95.02% |
| ZW | 119 | 700.0000 | 109 | 65324929372.0000 | 98.72% |

<!-- GEOIP-STATISTICS:END -->
