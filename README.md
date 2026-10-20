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

Last run: `2026-10-20T13:07:13Z`  
Input rows: **31464105** · Published blocks: **1052903**

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
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 471707 | 1 | 18351723 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 415649 | 0 | 0 | 9066810 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 563998 | 1 | 16757331 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 28507 | 0 | 430722 | 9616617 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288944 | 4075 | 8466227 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPINFO via tdulcet IPv4 | fresh | 31777 | 0 | 588831 | 12997637 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2707304 | 0 | 85118814 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPINFO Lite CSV | fresh | 1341028 | 2936383 | 0 | 208702865 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 19976 | 0 | 271288 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IPLocate | fresh | 785303 | 1368186 | 0 | 17235320 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177093 | 69680 | 0 | 3393343 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 287188 | 480983 | 97783 | 40299556 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| Clang CN | fresh | 4318 | 1675 | 0 | 97591 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3878 | 0 | 0 | 82281 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
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
| AD | 51 | 270.0000 | 99 | 43786570274.0000 | 90.73% |
| AE | 2306 | 25026.0000 | 1703 | 16051387874132.0000 | 84.76% |
| AF | 147 | 734.0000 | 124 | 121077432685.0000 | 96.69% |
| AG | 103 | 556.0000 | 140 | 39401981992.0000 | 67.57% |
| AI | 12 | 40.0000 | 47 | 4787144737.0000 | 72.11% |
| AL | 375 | 1775.0000 | 996 | 2630096612930.0000 | 94.58% |
| AM | 271 | 2336.0000 | 254 | 1258775590957.0000 | 94.51% |
| AO | 220 | 5157.0000 | 114 | 177326592578.0000 | 91.26% |
| AQ | 5 | 7.0000 | 338 | 4411131666.0000 | 62.90% |
| AR | 2369 | 76407.0000 | 1784 | 23549192565457.0000 | 98.93% |
| AS | 6 | 26.0000 | 147 | 8786610215.0000 | 98.85% |
| AT | 2666 | 44352.0000 | 2409 | 10406504937586.0000 | 97.23% |
| AU | 9241 | 213098.0000 | 47520 | 2305941913496945964.0000 | 100.00% |
| AW | 19 | 436.0000 | 48 | 10810099748.0000 | 89.81% |
| AX | 24 | 90.0000 | 79 | 68848650286.0000 | 78.51% |
| AZ | 295 | 3082.0000 | 182 | 1933255721522.0000 | 97.18% |
| BA | 225 | 3205.0000 | 138 | 760822566765.0000 | 97.06% |
| BB | 104 | 555.0000 | 69 | 72702888487.0000 | 98.07% |
| BD | 2292 | 8117.0000 | 2071 | 3564500824468.0000 | 95.21% |
| BE | 1759 | 55610.0000 | 55527 | 8421856965097724.0000 | 57.18% |
| BF | 85 | 1445.0000 | 96 | 90772560172.0000 | 99.25% |
| BG | 1694 | 16782.0000 | 801 | 5223397142104.0000 | 95.26% |
| BH | 144 | 3856.0000 | 180 | 401301783356.0000 | 98.24% |
| BI | 26 | 162.0000 | 57 | 26129203748.0000 | 92.45% |
| BJ | 54 | 625.0000 | 75 | 56210925868.0000 | 96.11% |
| BL | 5 | 6.0000 | 46 | 324405284.0000 | 94.54% |
| BM | 76 | 482.0000 | 78 | 47318960432.0000 | 95.61% |
| BN | 50 | 862.0000 | 74 | 25878855723.0000 | 97.07% |
| BO | 162 | 4722.0000 | 156 | 167741932107.0000 | 97.45% |
| BQ | 11 | 102.0000 | 50 | 21664565292.0000 | 90.16% |
| BR | 7654 | 326144.0000 | 12651 | 37128799782879.0000 | 99.34% |
| BS | 58 | 583.0000 | 83 | 13262559794.0000 | 99.75% |
| BT | 67 | 222.0000 | 112 | 60709470787.0000 | 99.58% |
| BV | 0 |  | 7 | 17891346.0000 | 48.60% |
| BW | 81 | 691.0000 | 83 | 78123716904.0000 | 96.88% |
| BY | 235 | 7375.0000 | 193 | 699954103199.0000 | 99.26% |
| BZ | 90 | 482.0000 | 169 | 199085312825.0000 | 89.33% |
| CA | 9595 | 280879.0000 | 8587 | 6493107107076.0000 | 93.77% |
| CC | 0 |  | 19 | 86836242.0000 | 56.79% |
| CD | 133 | 898.0000 | 139 | 155152466252.0000 | 98.42% |
| CF | 13 | 37.0000 | 52 | 544952353.0000 | 56.49% |
| CG | 40 | 578.0000 | 68 | 36380373028.0000 | 95.81% |
| CH | 3821 | 96196.0000 | 52564 | 15878799889542740366.0000 | 75.00% |
| CI | 286 | 6711.0000 | 97 | 194238490920.0000 | 98.74% |
| CK | 5 | 40.0000 | 50 | 8780121127.0000 | 98.88% |
| CL | 1125 | 40318.0000 | 1034 | 1405997738751.0000 | 96.44% |
| CM | 113 | 2432.0000 | 105 | 68730775619.0000 | 97.59% |
| CN | 5971 | 1349594.0000 | 5067 | 294876497072156.0000 | 99.91% |
| CO | 1220 | 69911.0000 | 1420 | 1954414485889.0000 | 97.93% |
| CR | 311 | 9145.0000 | 242 | 461496850781.0000 | 97.13% |
| CU | 31 | 1019.0000 | 67 | 26269057317.0000 | 99.23% |
| CV | 41 | 236.0000 | 103 | 13460834346.0000 | 98.15% |
| CW | 64 | 738.0000 | 69 | 126759012386.0000 | 71.10% |
| CX | 0 |  | 18 | 36571154.0000 | 66.14% |
| CY | 517 | 4563.0000 | 727 | 8464054761787.0000 | 85.24% |
| CZ | 2206 | 34260.0000 | 1997 | 13211843305391.0000 | 97.95% |
| DE | 16885 | 534835.0000 | 45540 | 110597316794890.0000 | 97.06% |
| DJ | 21 | 233.0000 | 68 | 9084405793.0000 | 97.52% |
| DK | 1724 | 49408.0000 | 1193 | 9500423591680.0000 | 93.19% |
| DM | 31 | 63.0000 | 53 | 43409345060.0000 | 76.03% |
| DO | 465 | 7982.0000 | 554 | 923727563944.0000 | 98.98% |
| DZ | 100 | 18712.0000 | 119 | 43354818692.0000 | 93.82% |
| EC | 486 | 11182.0000 | 518 | 1154000456478.0000 | 98.63% |
| EE | 733 | 6381.0000 | 787 | 3018246493444.0000 | 92.56% |
| EG | 328 | 94691.0000 | 167 | 17730858598764.0000 | 99.98% |
| EH | 0 |  | 23 | 270927890.0000 | 45.83% |
| ER | 9 | 25.0000 | 54 | 376132645.0000 | 62.10% |
| ES | 5536 | 137746.0000 | 2900 | 25681366810054.0000 | 97.78% |
| ET | 34 | 1466.0000 | 76 | 1117271824427.0000 | 99.98% |
| FI | 2180 | 58122.0000 | 44866 | 6799644778193.0000 | 96.69% |
| FJ | 58 | 592.0000 | 79 | 43158865703.0000 | 96.44% |
| FK | 6 | 21.0000 | 48 | 56428576.0000 | 68.29% |
| FM | 16 | 46.0000 | 63 | 8783923248.0000 | 91.53% |
| FO | 16 | 179.0000 | 59 | 107514235176.0000 | 99.95% |
| FR | 10106 | 326550.0000 | 6421 | 66896954411949.0000 | 98.66% |
| GA | 48 | 1551.0000 | 59 | 34735921439.0000 | 93.67% |
| GB | 17386 | 504060.0000 | 18718 | 103273119251933.0000 | 92.99% |
| GD | 50 | 98.0000 | 68 | 77687425574.0000 | 99.97% |
| GE | 382 | 5075.0000 | 272 | 1864612992826.0000 | 94.59% |
| GF | 31 | 132.0000 | 56 | 27085176868.0000 | 89.21% |
| GG | 61 | 288.0000 | 61 | 146236508449.0000 | 88.66% |
| GH | 228 | 8892.0000 | 177 | 1229338226262.0000 | 99.86% |
| GI | 75 | 411.0000 | 77 | 197841782820.0000 | 90.54% |
| GL | 35 | 145.0000 | 115 | 43165089898.0000 | 99.76% |
| GM | 42 | 1098.0000 | 64 | 39064066340.0000 | 99.61% |
| GN | 40 | 171.0000 | 77 | 47754713124.0000 | 98.04% |
| GP | 112 | 641.0000 | 67 | 47452149809.0000 | 74.48% |
| GQ | 28 | 93.0000 | 59 | 8999753767.0000 | 98.24% |
| GR | 792 | 23128.0000 | 493 | 2588527854162.0000 | 95.35% |
| GS | 1 | 2.0000 | 19 | 204144658.0000 | 47.68% |
| GT | 295 | 2739.0000 | 202 | 326888446322.0000 | 99.32% |
| GU | 61 | 954.0000 | 84 | 34453487668.0000 | 99.31% |
| GW | 14 | 38.0000 | 61 | 9133774887.0000 | 97.45% |
| GY | 26 | 287.0000 | 68 | 34518140968.0000 | 96.25% |
| HK | 6880 | 80786.0000 | 4866 | 9366506255068.0000 | 69.78% |
| HM | 1 | 2.0000 | 9 | 34734098.0000 | 52.49% |
| HN | 267 | 1904.0000 | 221 | 399903853166.0000 | 94.10% |
| HR | 412 | 9608.0000 | 313 | 1647242833566.0000 | 94.98% |
| HT | 36 | 616.0000 | 85 | 25995840045.0000 | 94.05% |
| HU | 942 | 23230.0000 | 747 | 3305822991305.0000 | 97.27% |
| ID | 5868 | 81994.0000 | 4021 | 3709018974178.0000 | 98.62% |
| IE | 2136 | 70813.0000 | 5519 | 4951141889728.0000 | 88.41% |
| IL | 1485 | 34916.0000 | 1818 | 4428321374502.0000 | 95.86% |
| IM | 109 | 545.0000 | 125 | 399707867679.0000 | 90.86% |
| IN | 9599 | 190485.0000 | 7503 | 29225203447083.0000 | 97.28% |
| IO | 3 | 14.0000 | 47 | 4518971425.0000 | 97.42% |
| IQ | 404 | 3293.0000 | 338 | 3337978253471.0000 | 97.34% |
| IR | 1872 | 42858.0000 | 780 | 17125506220581.0000 | 97.97% |
| IS | 230 | 3634.0000 | 316 | 1570685149503.0000 | 96.45% |
| IT | 5338 | 220172.0000 | 3462 | 47075150367165.0000 | 98.91% |
| JE | 63 | 639.0000 | 83 | 219183712546.0000 | 97.08% |
| JM | 96 | 892.0000 | 101 | 1121398114893.0000 | 100.00% |
| JO | 183 | 2565.0000 | 176 | 1121446447432.0000 | 96.97% |
| JP | 6810 | 774234.0000 | 6243 | 51547426738006.0000 | 89.31% |
| KE | 550 | 24882.0000 | 360 | 662746400424.0000 | 98.93% |
| KG | 189 | 1184.0000 | 115 | 528725576261.0000 | 98.99% |
| KH | 377 | 1973.0000 | 203 | 292861510228.0000 | 95.58% |
| KI | 21 | 34.0000 | 87 | 13081446467.0000 | 94.31% |
| KM | 18 | 46.0000 | 51 | 9133686818.0000 | 97.41% |
| KN | 35 | 110.0000 | 58 | 34836318256.0000 | 56.50% |
| KP | 4 | 8.0000 | 258 | 497090820.0000 | 55.42% |
| KR | 2136 | 452544.0000 | 1309 | 22724698297079.0000 | 99.70% |
| KW | 179 | 7507.0000 | 133 | 872397910069.0000 | 99.85% |
| KY | 57 | 407.0000 | 63 | 17675651367.0000 | 99.36% |
| KZ | 1053 | 13861.0000 | 566 | 2466952974961.0000 | 96.71% |
| LA | 91 | 415.0000 | 106 | 103575257381.0000 | 87.10% |
| LB | 301 | 2317.0000 | 207 | 2981207884594.0000 | 98.82% |
| LC | 46 | 148.0000 | 55 | 81963714848.0000 | 96.53% |
| LI | 90 | 434.0000 | 102 | 356783495716.0000 | 95.62% |
| LK | 122 | 2292.0000 | 129 | 73417326906.0000 | 95.58% |
| LR | 37 | 175.0000 | 71 | 26198890791.0000 | 96.92% |
| LS | 33 | 504.0000 | 57 | 26129207332.0000 | 98.71% |
| LT | 1130 | 12345.0000 | 945 | 5379043713744.0000 | 78.75% |
| LU | 593 | 6980.0000 | 587 | 2156912037218.0000 | 96.34% |
| LV | 883 | 8152.0000 | 401 | 1500165650792.0000 | 94.21% |
| LY | 99 | 1909.0000 | 78 | 99378679863.0000 | 89.65% |
| MA | 149 | 48065.0000 | 170 | 1165701681536.0000 | 99.92% |
| MC | 46 | 276.0000 | 73 | 146308663329.0000 | 98.73% |
| MD | 532 | 4979.0000 | 1079 | 6545850584837.0000 | 72.50% |
| ME | 81 | 799.0000 | 91 | 343924497710.0000 | 99.95% |
| MF | 11 | 29.0000 | 46 | 4652730404.0000 | 99.62% |
| MG | 47 | 2304.0000 | 73 | 22187430444.0000 | 98.77% |
| MH | 18 | 130.0000 | 45 | 4368695332.0000 | 99.48% |
| MK | 146 | 2766.0000 | 153 | 1069776162149.0000 | 98.42% |
| ML | 42 | 380.0000 | 58 | 21884567594.0000 | 98.35% |
| MM | 314 | 1090.0000 | 267 | 352664932690.0000 | 97.06% |
| MN | 164 | 740.0000 | 156 | 133696733246.0000 | 99.67% |
| MO | 61 | 1384.0000 | 338 | 38867108128.0000 | 97.53% |
| MP | 11 | 39.0000 | 47 | 4351131684.0000 | 99.59% |
| MQ | 118 | 867.0000 | 68 | 112950577184.0000 | 88.53% |
| MR | 28 | 216.0000 | 57 | 26145785893.0000 | 99.48% |
| MS | 3 | 7.0000 | 47 | 190056487.0000 | 47.90% |
| MT | 158 | 2662.0000 | 128 | 466226825512.0000 | 96.58% |
| MU | 1441 | 26444.0000 | 7838 | 4455234668089640.0000 | 100.00% |
| MV | 61 | 385.0000 | 92 | 17726357542.0000 | 98.66% |
| MW | 67 | 2194.0000 | 93 | 1241826611242.0000 | 99.86% |
| MX | 2184 | 118801.0000 | 1675 | 3532426690778.0000 | 97.42% |
| MY | 1555 | 30241.0000 | 1129 | 895968511825.0000 | 95.09% |
| MZ | 91 | 1854.0000 | 83 | 69146333239.0000 | 96.50% |
| NA | 105 | 1877.0000 | 66 | 29198717225.0000 | 98.43% |
| NC | 49 | 582.0000 | 79 | 47438824752.0000 | 99.79% |
| NE | 30 | 211.0000 | 55 | 21617180705.0000 | 98.79% |
| NF | 2 | 3.0000 | 44 | 173344804.0000 | 46.09% |
| NG | 532 | 12807.0000 | 447 | 632655841098.0000 | 96.08% |
| NI | 88 | 1687.0000 | 80 | 94700112182.0000 | 98.14% |
| NL | 11867 | 212274.0000 | 148179 | 8408464770377529.0000 | 99.94% |
| NO | 1741 | 63962.0000 | 1193 | 10871843190477.0000 | 95.49% |
| NP | 328 | 2370.0000 | 283 | 425874186324.0000 | 92.25% |
| NR | 15 | 49.0000 | 47 | 8781104164.0000 | 98.87% |
| NU | 4 | 11.0000 | 48 | 4485284900.0000 | 97.81% |
| NZ | 1678 | 29090.0000 | 1176 | 1338465740949.0000 | 95.75% |
| OM | 191 | 4312.0000 | 127 | 586526048303.0000 | 96.19% |
| PA | 274 | 7068.0000 | 436 | 463293638753.0000 | 90.30% |
| PE | 599 | 14096.0000 | 696 | 1068382467757.0000 | 96.25% |
| PF | 27 | 308.0000 | 109 | 21705263185.0000 | 99.47% |
| PG | 104 | 330.0000 | 110 | 73263219006.0000 | 90.78% |
| PH | 1474 | 25893.0000 | 1008 | 1037584842125.0000 | 96.35% |
| PK | 1007 | 23337.0000 | 635 | 898563379353.0000 | 97.01% |
| PL | 5532 | 82070.0000 | 3178 | 28347448674818.0000 | 96.53% |
| PM | 5 | 23.0000 | 47 | 4485154084.0000 | 97.81% |
| PN | 1 | 2.0000 | 15 | 52954130.0000 | 67.93% |
| PR | 385 | 4454.0000 | 216 | 223081773209.0000 | 97.53% |
| PS | 265 | 3318.0000 | 231 | 983827749179.0000 | 94.64% |
| PT | 691 | 26828.0000 | 763 | 3262348097826.0000 | 92.76% |
| PW | 8 | 31.0000 | 45 | 8647082020.0000 | 99.78% |
| PY | 265 | 5028.0000 | 212 | 451189490301.0000 | 97.10% |
| QA | 173 | 5059.0000 | 315 | 521252663960.0000 | 93.69% |
| RE | 68 | 1720.0000 | 64 | 129275461675.0000 | 84.98% |
| RO | 3043 | 29302.0000 | 1553 | 5722854030682.0000 | 94.10% |
| RS | 699 | 9102.0000 | 423 | 1856931245616.0000 | 98.40% |
| RU | 11456 | 180338.0000 | 6996 | 46798317560402.0000 | 94.95% |
| RW | 67 | 1224.0000 | 82 | 1147166871847.0000 | 99.86% |
| SA | 911 | 45955.0000 | 524 | 2731758014760.0000 | 97.79% |
| SB | 22 | 68.0000 | 60 | 17270965540.0000 | 96.25% |
| SC | 2472 | 13909.0000 | 3364 | 8600377688611.0000 | 56.69% |
| SD | 53 | 7419.0000 | 71 | 69199069483.0000 | 97.10% |
| SE | 4404 | 125563.0000 | 4622 | 30534091423399.0000 | 85.87% |
| SG | 5222 | 112093.0000 | 4605 | 149592929848738.0000 | 97.01% |
| SH | 1 | 2.0000 | 37 | 102762523.0000 | 58.14% |
| SI | 665 | 10398.0000 | 445 | 2289248264581.0000 | 92.18% |
| SJ | 1 | 2.0000 | 21 | 205129746.0000 | 49.36% |
| SK | 588 | 10970.0000 | 531 | 2900891184876.0000 | 97.14% |
| SL | 49 | 365.0000 | 70 | 26213222442.0000 | 98.53% |
| SM | 36 | 166.0000 | 70 | 197813078049.0000 | 93.14% |
| SN | 54 | 1618.0000 | 99 | 26416470077.0000 | 98.94% |
| SO | 44 | 219.0000 | 75 | 47705379110.0000 | 99.62% |
| SR | 26 | 339.0000 | 70 | 17358324012.0000 | 99.57% |
| SS | 50 | 132.0000 | 70 | 30592491560.0000 | 99.21% |
| ST | 12 | 56.0000 | 58 | 4553902116.0000 | 97.14% |
| SV | 169 | 2727.0000 | 143 | 210718886246.0000 | 96.54% |
| SX | 13 | 138.0000 | 54 | 30121003042.0000 | 95.18% |
| SY | 218 | 5143.0000 | 97 | 876758437931.0000 | 98.47% |
| SZ | 36 | 231.0000 | 61 | 30138716208.0000 | 98.80% |
| TC | 11 | 49.0000 | 49 | 1131743524.0000 | 64.50% |
| TD | 34 | 121.0000 | 61 | 13059885093.0000 | 99.46% |
| TF | 0 |  | 16 | 170524690.0000 | 43.71% |
| TG | 32 | 1402.0000 | 79 | 26162780453.0000 | 98.57% |
| TH | 1533 | 40401.0000 | 1048 | 593262439881.0000 | 93.74% |
| TJ | 96 | 382.0000 | 92 | 670777084461.0000 | 96.37% |
| TK | 4 | 14.0000 | 52 | 4351133735.0000 | 99.59% |
| TL | 43 | 118.0000 | 69 | 43426906151.0000 | 96.97% |
| TM | 28 | 125.0000 | 53 | 68810376489.0000 | 99.94% |
| TN | 61 | 30793.0000 | 96 | 112216392236.0000 | 99.79% |
| TO | 18 | 53.0000 | 51 | 17371299875.0000 | 99.43% |
| TR | 2513 | 67656.0000 | 2095 | 23760838548157.0000 | 89.29% |
| TT | 92 | 2211.0000 | 95 | 1177082808890.0000 | 70.55% |
| TV | 9 | 41.0000 | 59 | 12388402211.0000 | 81.04% |
| TW | 1403 | 144861.0000 | 1804 | 11502680736959.0000 | 98.20% |
| TZ | 246 | 4449.0000 | 173 | 266856022632.0000 | 98.29% |
| UA | 4197 | 35033.0000 | 2125 | 11708457781752.0000 | 92.08% |
| UG | 153 | 5714.0000 | 104 | 125034848556.0000 | 97.09% |
| UM | 1 | 2.0000 | 34 | 188286994.0000 | 47.08% |
| US | 63351 | 5980037.0000 | 160336 | 4836998134325085.0000 | 99.77% |
| UY | 120 | 10324.0000 | 120 | 177218666535.0000 | 99.63% |
| UZ | 348 | 1600.0000 | 175 | 1083177884210.0000 | 99.38% |
| VA | 5 | 43.0000 | 53 | 103155697951.0000 | 99.97% |
| VC | 39 | 84.0000 | 48 | 69045127201.0000 | 99.97% |
| VE | 536 | 29797.0000 | 477 | 2252950614404.0000 | 99.31% |
| VG | 181 | 2944.0000 | 257 | 606926578739.0000 | 65.20% |
| VI | 33 | 523.0000 | 112 | 30144751663.0000 | 94.45% |
| VN | 1681 | 65190.0000 | 1415 | 1902165205856.0000 | 97.95% |
| VU | 29 | 68.0000 | 252 | 51797953316.0000 | 85.99% |
| WF | 4 | 16.0000 | 44 | 4350806052.0000 | 99.57% |
| WS | 28 | 92.0000 | 61 | 21683241255.0000 | 91.03% |
| XK | 74 | 250.0000 | 154 | 37848875009.0000 | 50.00% |
| YE | 51 | 949.0000 | 86 | 206643003444.0000 | 99.92% |
| YT | 12 | 20.0000 | 48 | 148176932.0000 | 59.75% |
| ZA | 3293 | 114120.0000 | 1463 | 23426274705438.0000 | 99.34% |
| ZM | 103 | 6395.0000 | 86 | 43426992450.0000 | 94.83% |
| ZW | 119 | 700.0000 | 105 | 65308414296.0000 | 98.79% |

<!-- GEOIP-STATISTICS:END -->
