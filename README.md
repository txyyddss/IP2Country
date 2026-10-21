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

Last run: `2026-10-21T11:33:11Z`

Input rows: **38931386** · Published blocks: **1253792**

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
| Server-Country via sapics IPv4 | fresh | 398935 | 0 | 0 | 0 | 8692969 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv4.csv |
| Server-Country via sapics IPv6 | fresh | 0 | 471961 | 0 | 1 | 15533083 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/server-country/server-country-ipv6.csv |
| Geo-Whois-ASN-Country IPv4 | fresh | 398706 | 0 | 0 | 0 | 5934606 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv4.tsv?inline=false |
| Geo-Whois-ASN-Country IPv6 | fresh | 0 | 471750 | 0 | 1 | 18355794 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/geo-whois-asn-country/ipv6.tsv?inline=false |
| User-Country IPv4 | fresh | 415878 | 0 | 0 | 0 | 9071280 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv4.csv |
| User-Country IPv6 | fresh | 0 | 564209 | 0 | 1 | 16758520 | https://raw.githubusercontent.com/sapics/ip-location-db/refs/heads/main/user-country/user-country-ipv6.csv |
| IP to ASN Country IPv4 | fresh | 722178 | 0 | 0 | 2149 | 9617773 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv4.tsv?inline=false |
| IP to ASN Country IPv6 | fresh | 0 | 288468 | 0 | 3995 | 8457190 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/iptoasn-country/ipv6.tsv?inline=false |
| IPtoASN via sapics IPv4 CSV | fresh | 321845 | 0 | 0 | 0 | 7034204 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv4.csv |
| IPtoASN via sapics IPv6 CSV | fresh | 0 | 87699 | 0 | 0 | 3806985 | https://github.com/sapics/ip-location-db/releases/download/latest/iptoasn-country-ipv6.csv |
| IPINFO via tdulcet IPv4 | fresh | 982030 | 0 | 0 | 0 | 12997637 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv4.tsv?inline=false |
| IPINFO via tdulcet IPv6 | fresh | 0 | 2707304 | 0 | 0 | 85118814 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ipinfo-country/ipv6.tsv?inline=false |
| IPinfo via xream MMDB | fresh | 955329 | 2063526 | 0 | 17 | 18149674 | https://github.com/xream/geoip/releases/latest/download/ipinfo.country.mmdb |
| IPINFO Lite CSV | fresh | 1341028 | 2936383 | 0 | 0 | 208702865 | https://github.com/alice39s/ipinfo-csv-lite/releases/latest/download/ipinfo-lite.csv |
| IP2Location IPv4 | fresh | 390273 | 0 | 0 | 0 | 6100911 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv4.tsv?inline=false |
| IP2Location IPv6 | fresh | 0 | 1512762 | 0 | 0 | 24574212 | https://gitlab.com/tdulcet/ip-geolocation-dbs/-/raw/main/ip2location-country/ipv6.tsv?inline=false |
| IP2Location Lite via xream MMDB | fresh | 384815 | 1499498 | 0 | 314195 | 13193749 | https://github.com/xream/geoip/releases/latest/download/ip2location.country.mmdb |
| IPLocate | fresh | 785351 | 1366216 | 0 | 0 | 17219944 | https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb |
| IPGeo | fresh | 177102 | 69690 | 0 | 0 | 3393649 | https://github.com/analogic/ipgeo/raw/refs/heads/master/ipgeo-country.mmdb |
| IPRanges | fresh | 286757 | 487213 | 0 | 97794 | 40616299 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.json |
| IPRanges country CSV | fresh | 286757 | 487213 | 0 | 98381 | 18365449 | https://raw.githubusercontent.com/hexydec/ip-ranges/main/output/countries.csv |
| Clang CN | fresh | 4317 | 1676 | 0 | 0 | 97598 | https://ispip.clang.cn/all_cn_ipv46.txt |
| CHNRoutes2 | fresh | 3876 | 0 | 0 | 0 | 82317 | https://github.com/misakaio/chnroutes2/raw/refs/heads/master/chnroutes.mmdb |
| SYPEX Country | fresh | 327605 | 0 | 0 | 0 | 537559 | https://sypexgeo.net/files/SxGeoCountry.zip |
| IP66 | fresh | 806857 | 417982 | 0 | 410891 | 18173463 | https://downloads.ip66.dev/db/ip66.mmdb |
| IANA IP Blocks | fresh | 262237 | 70946 | 0 | 10 | 34567370 | https://raw.githubusercontent.com/HotCakeX/Official-IANA-IP-blocks/refs/heads/main/JSON/global.json |
| IP-Geolocation-DB | fresh | 923960 | 763194 | 0 | 1370 | 15398223 | https://github.com/HostByBelle/IP-Geolocation-DB/releases/latest/download/cc0-pddl-country-asn-both-variant-1.mmdb |
| IPAPI.IS IPv4 | fresh | 4921366 | 0 | 0 | 0 | 65954248 | https://ipapi.is/data/geolocationDatabaseIPv4.csv.zip |
| IPAPI.IS IPv6 | fresh | 0 | 792430 | 0 | 0 | 8460577 | https://ipapi.is/data/geolocationDatabaseIPv6.csv.zip |
| IP2Region | fresh | 802825 | 0 | 0 | 0 | 11122642 | https://github.com/adysec/IP_database/raw/refs/heads/main/ip2region/ip2region.xdb |
| gaoyifan China operator IPv4+IPv6 | fresh | 6525 | 3488 | 0 | 0 | 159289 | https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china46.txt |
| Iran CIDRs by parsamoh | fresh | 1748 | 574 | 0 | 0 | 36004 | https://github.com/parsamoh/geoip/releases/latest/download/iran-cidr.txt |
| GeoLite2 Country via P3TERX | fresh | 565511 | 500934 | 0 | 0 | 3408573 | https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb |

### Final country coverage

| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |
|---|---:|---:|---:|---:|---:|
| AD | 72 | 266 | 99 | 43887499570 | 96.15% |
| AE | 2637 | 22959 | 2447 | 16207570889695 | 85.76% |
| AF | 190 | 718 | 150 | 121229023871 | 96.16% |
| AG | 133 | 583 | 314 | 39553061962 | 66.27% |
| AI | 29 | 40 | 67 | 4939585330 | 85.14% |
| AL | 350 | 2001 | 990 | 2668046171981 | 95.88% |
| AM | 315 | 2320 | 244 | 1259042067013 | 94.94% |
| AO | 271 | 5149 | 134 | 177612333907 | 90.66% |
| AQ | 55 | 11 | 408 | 4651734659 | 67.96% |
| AR | 3014 | 76206 | 1783 | 23583364156075 | 98.88% |
| AS | 30 | 26 | 71 | 8932561720 | 98.28% |
| AT | 3630 | 44117 | 2606 | 10504655859298 | 96.97% |
| AU | 10891 | 206843 | 50613 | 2305942026002009360 | 100.00% |
| AW | 42 | 435 | 69 | 11096888117 | 90.25% |
| AX | 81 | 48 | 80 | 69067215663 | 80.84% |
| AZ | 347 | 3065 | 237 | 1933539105092 | 99.14% |
| BA | 330 | 3200 | 149 | 760895183999 | 97.07% |
| BB | 126 | 552 | 87 | 73661093432 | 97.77% |
| BD | 2399 | 8095 | 2055 | 3572064617637 | 94.66% |
| BE | 4737 | 55428 | 68627 | 8421683825552667 | 50.04% |
| BF | 105 | 1425 | 116 | 90892100670 | 99.03% |
| BG | 2032 | 16792 | 813 | 5244612509469 | 94.95% |
| BH | 204 | 3856 | 195 | 401557379148 | 98.85% |
| BI | 41 | 153 | 80 | 26415469622 | 90.93% |
| BJ | 67 | 611 | 97 | 56498238269 | 95.22% |
| BL | 16 | 5 | 65 | 611062565 | 74.87% |
| BM | 317 | 515 | 98 | 47605749313 | 95.43% |
| BN | 88 | 858 | 98 | 26132157244 | 97.99% |
| BO | 211 | 4708 | 183 | 167961564766 | 97.80% |
| BQ | 33 | 104 | 67 | 21816873773 | 90.21% |
| BR | 8598 | 323115 | 13482 | 37372265355132 | 99.20% |
| BS | 82 | 570 | 105 | 13565192520 | 98.68% |
| BT | 107 | 214 | 131 | 60827964500 | 99.55% |
| BV | 0 | 0 | 16 | 68419860 | 49.29% |
| BW | 113 | 672 | 104 | 78258463546 | 96.34% |
| BY | 289 | 7358 | 172 | 700202884017 | 99.19% |
| BZ | 128 | 499 | 194 | 197359867473 | 89.73% |
| CA | 13712 | 279119 | 15064 | 6360240575821 | 94.75% |
| CC | 2 | 0 | 37 | 323556116 | 56.60% |
| CD | 184 | 808 | 163 | 155387962206 | 98.15% |
| CF | 36 | 32 | 73 | 679699251 | 61.40% |
| CG | 55 | 571 | 90 | 36633674549 | 94.55% |
| CH | 5387 | 94197 | 61560 | 15878799548146917961 | 50.01% |
| CI | 337 | 6765 | 115 | 193972155177 | 98.59% |
| CK | 27 | 40 | 70 | 8932955944 | 98.28% |
| CL | 1358 | 39647 | 1083 | 1408148208850 | 96.46% |
| CM | 133 | 2402 | 121 | 69369885525 | 97.20% |
| CN | 6536 | 1349686 | 6073 | 294923763591147 | 99.90% |
| CO | 1817 | 69991 | 1474 | 1953382712402 | 98.11% |
| CR | 354 | 9067 | 269 | 461646029710 | 97.11% |
| CU | 76 | 1012 | 93 | 26403739193 | 99.08% |
| CV | 70 | 209 | 122 | 13596627771 | 97.94% |
| CW | 87 | 732 | 367 | 66916176658 | 90.34% |
| CX | 2 | 0 | 39 | 323687188 | 54.79% |
| CY | 696 | 4287 | 752 | 8416904988237 | 87.32% |
| CZ | 3110 | 34442 | 4613 | 13172874347653 | 98.04% |
| DE | 28554 | 530136 | 65298 | 111040713499385 | 96.89% |
| DJ | 36 | 195 | 83 | 43932390194 | 67.72% |
| DK | 3184 | 49051 | 2248 | 9536374239210 | 91.84% |
| DM | 52 | 59 | 73 | 43562178101 | 77.11% |
| DO | 494 | 7917 | 563 | 928173987027 | 98.89% |
| DZ | 163 | 18699 | 139 | 43640558485 | 93.70% |
| EC | 577 | 11149 | 501 | 1158509205297 | 98.57% |
| EE | 1151 | 6318 | 765 | 3280804047103 | 89.28% |
| EG | 413 | 94679 | 185 | 17730957219453 | 99.98% |
| EH | 2 | 0 | 37 | 322900755 | 56.69% |
| ER | 30 | 19 | 76 | 629351223 | 60.25% |
| ES | 8717 | 136562 | 3724 | 25704300621166 | 97.78% |
| ET | 55 | 1460 | 97 | 1117390906173 | 99.98% |
| FI | 3768 | 58261 | 36923 | 6831529383929 | 96.56% |
| FJ | 80 | 577 | 121 | 43328084536 | 98.19% |
| FK | 25 | 17 | 67 | 342954786 | 55.83% |
| FM | 41 | 39 | 85 | 8936297284 | 94.83% |
| FO | 39 | 178 | 75 | 107717006378 | 99.87% |
| FR | 19570 | 325923 | 7232 | 66941129090623 | 98.62% |
| GA | 66 | 1561 | 77 | 34989040434 | 92.30% |
| GB | 28841 | 503416 | 22527 | 103322525069104 | 92.87% |
| GD | 67 | 87 | 87 | 77973886519 | 99.79% |
| GE | 454 | 5083 | 305 | 1902294267211 | 94.32% |
| GF | 41 | 95 | 73 | 22892450597 | 95.68% |
| GG | 130 | 291 | 71 | 146371580963 | 87.38% |
| GH | 279 | 8883 | 202 | 1233883553895 | 99.67% |
| GI | 121 | 408 | 92 | 197910270759 | 92.56% |
| GL | 72 | 142 | 138 | 43317465966 | 99.63% |
| GM | 59 | 1092 | 83 | 39317234486 | 99.35% |
| GN | 57 | 164 | 96 | 47890441013 | 97.99% |
| GP | 67 | 349 | 106 | 49752400690 | 72.91% |
| GQ | 37 | 80 | 77 | 9219367736 | 97.43% |
| GR | 1331 | 23086 | 650 | 2592944299252 | 95.37% |
| GS | 3 | 2 | 39 | 339743508 | 55.74% |
| GT | 746 | 2725 | 221 | 327187523247 | 99.19% |
| GU | 142 | 975 | 115 | 34739786549 | 98.68% |
| GW | 33 | 32 | 77 | 9219171128 | 97.44% |
| GY | 42 | 279 | 87 | 34754597690 | 99.51% |
| HK | 9062 | 74284 | 6535 | 8779062632195 | 69.12% |
| HM | 0 | 0 | 20 | 244580628 | 41.60% |
| HN | 911 | 2117 | 228 | 400490682049 | 94.07% |
| HR | 809 | 9589 | 327 | 1677405043181 | 94.14% |
| HT | 56 | 606 | 92 | 26130781758 | 93.90% |
| HU | 1457 | 23208 | 664 | 3339569108376 | 96.90% |
| ID | 5978 | 76948 | 3712 | 3704008946499 | 98.63% |
| IE | 3682 | 67161 | 5762 | 5042851227416 | 87.67% |
| IL | 1929 | 34536 | 1801 | 4413998334676 | 96.25% |
| IM | 234 | 542 | 125 | 399791953698 | 89.64% |
| IN | 11268 | 186175 | 9656 | 29424427304667 | 96.87% |
| IO | 19 | 14 | 61 | 4637659938 | 96.74% |
| IQ | 646 | 3249 | 471 | 3471557502129 | 95.99% |
| IR | 1909 | 42877 | 794 | 17490227943718 | 97.29% |
| IS | 268 | 3623 | 322 | 1570806491985 | 96.56% |
| IT | 8338 | 218020 | 4486 | 47148963396577 | 98.90% |
| JE | 108 | 638 | 80 | 219403195429 | 97.24% |
| JM | 160 | 893 | 126 | 1121684297310 | 99.98% |
| JO | 242 | 2551 | 158 | 1091613120089 | 98.11% |
| JP | 8805 | 769594 | 5215 | 51529019895627 | 88.92% |
| KE | 650 | 24575 | 857 | 671364257978 | 98.49% |
| KG | 213 | 1182 | 140 | 528978876758 | 98.91% |
| KH | 430 | 1962 | 218 | 297306510951 | 94.40% |
| KI | 44 | 24 | 110 | 13233951574 | 96.50% |
| KM | 42 | 39 | 70 | 9235927859 | 97.33% |
| KN | 41 | 82 | 78 | 34989675330 | 57.63% |
| KP | 3 | 4 | 274 | 633340949 | 61.44% |
| KR | 2572 | 450349 | 1356 | 22722926865458 | 99.69% |
| KW | 210 | 7489 | 144 | 872665531213 | 99.96% |
| KY | 90 | 409 | 81 | 17827830312 | 99.04% |
| KZ | 1166 | 13864 | 466 | 2432565832849 | 98.02% |
| LA | 124 | 422 | 128 | 103744280119 | 86.10% |
| LB | 336 | 2324 | 247 | 2981477815875 | 98.90% |
| LC | 69 | 137 | 75 | 82250371889 | 96.88% |
| LI | 130 | 423 | 130 | 429875603765 | 88.33% |
| LK | 277 | 2267 | 163 | 78014944075 | 94.16% |
| LR | 59 | 170 | 89 | 26452124472 | 97.84% |
| LS | 52 | 493 | 79 | 26399220533 | 97.97% |
| LT | 1421 | 12590 | 1322 | 5396226191615 | 79.61% |
| LU | 1169 | 6926 | 1015 | 2187113733492 | 96.21% |
| LV | 1028 | 8184 | 417 | 1500758065293 | 94.34% |
| LY | 132 | 1900 | 97 | 99496649544 | 89.65% |
| MA | 261 | 48021 | 181 | 1164917020050 | 99.96% |
| MC | 77 | 274 | 76 | 146422764338 | 98.63% |
| MD | 589 | 4933 | 1132 | 6580341234200 | 68.64% |
| ME | 162 | 778 | 109 | 344009041471 | 99.95% |
| MF | 23 | 21 | 64 | 4939256613 | 96.89% |
| MG | 70 | 2296 | 95 | 22423888959 | 98.20% |
| MH | 39 | 126 | 71 | 39015158582 | 64.37% |
| MK | 203 | 2754 | 176 | 1099975431783 | 97.81% |
| ML | 66 | 374 | 79 | 22137738043 | 97.56% |
| MM | 341 | 1075 | 272 | 352936003172 | 96.75% |
| MN | 202 | 726 | 182 | 133848453968 | 99.58% |
| MO | 99 | 1383 | 149 | 38991452196 | 97.94% |
| MP | 61 | 16 | 64 | 4637463333 | 96.69% |
| MQ | 104 | 660 | 66 | 113086567201 | 89.03% |
| MR | 47 | 214 | 78 | 26432574262 | 98.99% |
| MS | 18 | 7 | 68 | 342496042 | 54.84% |
| MT | 216 | 2650 | 145 | 475051791932 | 95.69% |
| MU | 1427 | 27593 | 8145 | 4455226466919224 | 100.00% |
| MV | 78 | 378 | 95 | 17843163959 | 98.56% |
| MW | 94 | 2186 | 129 | 1241945629499 | 99.86% |
| MX | 2864 | 118257 | 2686 | 3494244457636 | 97.93% |
| MY | 1843 | 29085 | 1036 | 900810946571 | 94.50% |
| MZ | 107 | 1789 | 105 | 69399503688 | 95.67% |
| NA | 169 | 1894 | 87 | 29334463296 | 97.87% |
| NC | 67 | 597 | 115 | 47607911985 | 99.65% |
| NE | 45 | 204 | 78 | 21836927796 | 97.98% |
| NF | 23 | 3 | 64 | 325718823 | 55.26% |
| NG | 762 | 12799 | 450 | 632807938414 | 95.58% |
| NI | 105 | 1680 | 103 | 94868411975 | 97.99% |
| NL | 17750 | 204403 | 171133 | 8406467342116644 | 99.95% |
| NO | 2399 | 63049 | 1219 | 10893395716082 | 95.86% |
| NP | 405 | 2361 | 301 | 430305145701 | 90.74% |
| NR | 23 | 42 | 68 | 8933479222 | 98.28% |
| NU | 21 | 11 | 71 | 4637725479 | 96.71% |
| NZ | 1831 | 28652 | 1126 | 1337107662724 | 95.60% |
| OM | 257 | 4277 | 125 | 584832340800 | 96.51% |
| PA | 541 | 7639 | 452 | 463361143922 | 90.45% |
| PE | 942 | 14037 | 790 | 1073021719235 | 96.25% |
| PF | 50 | 304 | 129 | 21845250898 | 99.24% |
| PG | 144 | 314 | 140 | 73444759119 | 89.68% |
| PH | 1663 | 25632 | 1005 | 1033458240673 | 96.67% |
| PK | 1182 | 23303 | 600 | 902933734339 | 96.69% |
| PL | 7210 | 81183 | 2448 | 28404641641037 | 96.87% |
| PM | 32 | 22 | 68 | 4637464102 | 96.69% |
| PN | 3 | 2 | 36 | 339546900 | 52.24% |
| PR | 333 | 3061 | 296 | 212529398666 | 99.61% |
| PS | 283 | 3310 | 248 | 983945716542 | 96.53% |
| PT | 1403 | 26770 | 775 | 3262739077793 | 92.94% |
| PW | 32 | 28 | 68 | 8949404470 | 98.18% |
| PY | 316 | 5032 | 234 | 451439253134 | 97.45% |
| QA | 169 | 4464 | 246 | 520463349563 | 93.77% |
| RE | 69 | 1610 | 155 | 129631853882 | 84.06% |
| RO | 3658 | 29215 | 1633 | 5758988840288 | 94.62% |
| RS | 859 | 9090 | 905 | 1826112240241 | 99.10% |
| RU | 12339 | 180684 | 5264 | 47232742203773 | 94.75% |
| RW | 124 | 1178 | 102 | 1147420107576 | 99.79% |
| SA | 1067 | 45852 | 665 | 2729320387124 | 98.19% |
| SB | 39 | 60 | 85 | 17557820981 | 95.77% |
| SC | 3345 | 19456 | 4191 | 8437261604190 | 60.63% |
| SD | 85 | 7413 | 92 | 69484811068 | 96.49% |
| SE | 7501 | 123485 | 5698 | 30597173915480 | 88.96% |
| SG | 7288 | 113492 | 4651 | 149393525977288 | 96.91% |
| SH | 1 | 2 | 45 | 340071198 | 56.42% |
| SI | 978 | 10399 | 576 | 2229467818126 | 93.02% |
| SJ | 2 | 2 | 38 | 340661012 | 57.04% |
| SK | 828 | 10926 | 560 | 2905681634816 | 96.98% |
| SL | 82 | 359 | 88 | 26449679163 | 97.85% |
| SM | 51 | 159 | 87 | 202226799410 | 91.93% |
| SN | 80 | 1610 | 116 | 26517726030 | 98.88% |
| SO | 50 | 211 | 95 | 47940789303 | 99.43% |
| SR | 33 | 330 | 89 | 17578005565 | 99.02% |
| SS | 68 | 122 | 93 | 30715769658 | 99.12% |
| ST | 28 | 50 | 69 | 4638839606 | 96.89% |
| SV | 327 | 2708 | 178 | 210888246918 | 96.24% |
| SX | 40 | 138 | 71 | 30407529250 | 94.75% |
| SY | 269 | 5144 | 120 | 876859761471 | 99.05% |
| SZ | 50 | 220 | 84 | 30408794945 | 98.12% |
| TC | 17 | 45 | 69 | 613226022 | 74.83% |
| TD | 54 | 111 | 78 | 13295227703 | 98.68% |
| TF | 0 | 0 | 36 | 322900500 | 54.45% |
| TG | 53 | 1394 | 97 | 26416014134 | 97.77% |
| TH | 1686 | 40061 | 1134 | 587670289634 | 93.73% |
| TJ | 122 | 376 | 114 | 670984247614 | 96.54% |
| TK | 15 | 14 | 71 | 4637594410 | 96.69% |
| TL | 75 | 118 | 93 | 43579609912 | 98.25% |
| TM | 38 | 117 | 76 | 69046965818 | 99.79% |
| TN | 97 | 30784 | 111 | 112334423101 | 99.77% |
| TO | 32 | 47 | 73 | 17506898741 | 99.16% |
| TR | 2875 | 67413 | 1885 | 24439583848892 | 89.70% |
| TT | 108 | 2192 | 116 | 1177218602571 | 70.55% |
| TV | 22 | 35 | 80 | 11096101684 | 87.45% |
| TW | 1719 | 144377 | 1340 | 11524176714076 | 98.13% |
| TZ | 281 | 4351 | 190 | 266990637945 | 97.95% |
| UA | 4468 | 34883 | 1973 | 11318517491010 | 93.71% |
| UG | 242 | 5613 | 123 | 125268226878 | 97.01% |
| UM | 1 | 2 | 34 | 322704148 | 54.26% |
| US | 89253 | 6045259 | 149705 | 4837210524932875 | 99.78% |
| UY | 148 | 10309 | 141 | 177321104441 | 99.50% |
| UZ | 403 | 1591 | 175 | 1083275831621 | 99.63% |
| VA | 19 | 43 | 78 | 103426544688 | 99.84% |
| VC | 59 | 76 | 69 | 69331850035 | 99.78% |
| VE | 704 | 29696 | 504 | 2253233276063 | 99.35% |
| VG | 219 | 2913 | 398 | 542654935842 | 67.94% |
| VI | 59 | 521 | 133 | 30414336824 | 92.86% |
| VN | 1891 | 65084 | 1373 | 1906321079315 | 97.78% |
| VU | 45 | 62 | 270 | 51933812533 | 89.76% |
| WF | 15 | 16 | 69 | 4637332245 | 96.67% |
| WS | 43 | 86 | 85 | 21835617848 | 89.42% |
| XK | 4 | 3 | 5 | 134283267 | 35.71% |
| YE | 94 | 938 | 117 | 206895060791 | 99.86% |
| YT | 24 | 14 | 67 | 367858470 | 57.26% |
| ZA | 3600 | 112881 | 1319 | 23429954917809 | 99.30% |
| ZM | 108 | 6364 | 107 | 43630027860 | 93.77% |
| ZW | 109 | 612 | 145 | 65442496345 | 98.40% |

<!-- GEOIP-STATISTICS:END -->
