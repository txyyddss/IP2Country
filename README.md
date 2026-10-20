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

Last run: `2026-10-20T15:34:07Z`

Input rows: **34366466** · Published blocks: **1281911**

### Source availability

| Source | Status | IPv4 rows | IPv6 rows | Rejected | No country | Bytes | Details |
|---|---:|---:|---:|---:|---:|---:|---|
| DB-IP via sapics IPv4 | cached | 597286 | 0 | 0 | 0 | 11032548 | local cache |
| DB-IP via sapics IPv6 | cached | 0 | 785668 | 0 | 0 | 20177287 | local cache |
| DB-IP via tdulcet IPv4 | cached | 597286 | 0 | 0 | 0 | 7580021 | local cache |
| DB-IP via tdulcet IPv6 | cached | 0 | 785668 | 0 | 0 | 24060852 | local cache |
| GeoLite2 via sapics IPv4 | cached | 562225 | 0 | 0 | 0 | 11070001 | local cache |
| GeoLite2 via sapics IPv6 | cached | 0 | 500017 | 0 | 0 | 14970875 | local cache |
| GeoLite2 via tdulcet IPv4 | cached | 564219 | 0 | 0 | 532 | 11825975 | local cache |
| GeoLite2 via tdulcet IPv6 | cached | 0 | 500090 | 0 | 153 | 34516461 | local cache |
| Server-Country via sapics IPv4 | cached | 398706 | 0 | 0 | 0 | 8688499 | local cache |
| Server-Country via sapics IPv6 | cached | 0 | 471750 | 0 | 1 | 15531894 | local cache |
| Geo-Whois-ASN-Country IPv4 | cached | 398438 | 0 | 0 | 0 | 5930350 | local cache |
| Geo-Whois-ASN-Country IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| User-Country IPv4 | cached | 415649 | 0 | 0 | 0 | 9066810 | local cache |
| User-Country IPv6 | cached | 0 | 563998 | 0 | 1 | 16757331 | local cache |
| IP to ASN Country IPv4 | cached | 722193 | 0 | 0 | 2152 | 9616617 | local cache |
| IP to ASN Country IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPtoASN via sapics IPv4 CSV | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPtoASN via sapics IPv6 CSV | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPINFO via tdulcet IPv4 | cached | 982030 | 0 | 0 | 0 | 12997637 | local cache |
| IPINFO via tdulcet IPv6 | cached | 0 | 2707304 | 0 | 0 | 85118814 | local cache |
| IPinfo via xream MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPINFO Lite CSV | cached | 1341028 | 2936383 | 0 | 0 | 208702865 | local cache |
| IP2Location IPv4 | cached | 390273 | 0 | 0 | 0 | 6100911 | local cache |
| IP2Location IPv6 | cached | 0 | 1512762 | 0 | 0 | 24574212 | local cache |
| IP2Location Lite via xream MMDB | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPLocate | cached | 785303 | 1368186 | 0 | 0 | 17235320 | local cache |
| IPGeo | cached | 177093 | 69680 | 0 | 0 | 3393343 | local cache |
| IPGeo IPv4 CSV | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPGeo IPv6 CSV | disabled | 0 | 0 | 0 | 0 | 0 | — |
| IPRanges | cached | 287188 | 480983 | 0 | 97783 | 40299556 | local cache |
| IPRanges country CSV | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| Clang CN | cached | 4318 | 1675 | 0 | 0 | 97591 | local cache |
| CHNRoutes2 | cached | 3879 | 0 | 0 | 0 | 82305 | local cache |
| SYPEX Country | cached | 327605 | 0 | 0 | 0 | 537559 | local cache |
| IP66 | cached | 808852 | 422317 | 0 | 411004 | 18232363 | local cache |
| IANA IP Blocks | cached | 262213 | 70936 | 0 | 10 | 34563823 | local cache |
| IP-Geolocation-DB | cached | 923960 | 763194 | 0 | 1370 | 15398223 | local cache |
| IPAPI.IS IPv4 | cached | 4921366 | 0 | 0 | 0 | 65954248 | local cache |
| IPAPI.IS IPv6 | cached | 0 | 792430 | 0 | 0 | 8460577 | local cache |
| IP2Region | cached | 802825 | 0 | 0 | 0 | 11122642 | local cache |
| V2ray GeoIP | cached | 597286 | 785617 | 0 | 24 | 23258601 | local cache |
| Meta Rules Dat | cached | 512893 | 413486 | 0 | 10664 | 7647337 | local cache |
| Loyalsoldier GeoIP Enhanced | cached | 557938 | 492270 | 0 | 10655 | 16587463 | local cache |
| gaoyifan China operator IPv4+IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| Iran CIDRs by parsamoh | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPLocate CSV ZIP alternative | disabled | 0 | 0 | 0 | 0 | 0 | — |
| GeoLite2 Country via P3TERX | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| GeoIP All-in-One consensus MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| laincat GeoIP MMDB | disabled | 0 | 0 | 0 | 0 | 0 | — |
| 17mon China CIDRs (stale backup) | disabled | 0 | 0 | 0 | 0 | 0 | — |

### Final country coverage

| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |
|---|---:|---:|---:|---:|---:|
| AD | 74 | 266 | 105 | 43822156322 | 90.95% |
| AE | 3009 | 25035 | 1729 | 14118750646214 | 89.48% |
| AF | 199 | 730 | 124 | 121110659437 | 96.88% |
| AG | 139 | 525 | 138 | 39401343016 | 64.63% |
| AI | 34 | 41 | 46 | 4786882593 | 75.31% |
| AL | 369 | 2004 | 982 | 2660173950270 | 94.74% |
| AM | 313 | 2327 | 286 | 1297463670580 | 94.18% |
| AO | 344 | 5150 | 225 | 177326248002 | 90.24% |
| AQ | 61 | 9 | 307 | 4407592980 | 66.35% |
| AR | 3192 | 76395 | 1802 | 23553553103246 | 99.19% |
| AS | 32 | 27 | 146 | 8786348071 | 98.84% |
| AT | 3863 | 44546 | 2677 | 10360386516161 | 97.57% |
| AU | 11127 | 212810 | 47685 | 2305941900050310086 | 100.00% |
| AW | 45 | 436 | 46 | 10809772068 | 91.05% |
| AX | 117 | 87 | 83 | 68882008112 | 82.76% |
| AZ | 380 | 3078 | 226 | 1933355926066 | 97.80% |
| BA | 292 | 3209 | 142 | 760843767662 | 97.70% |
| BB | 160 | 531 | 64 | 73508391207 | 97.71% |
| BD | 2454 | 8110 | 2071 | 3568828379540 | 95.50% |
| BE | 4858 | 55533 | 72058 | 8421816082124027 | 60.03% |
| BF | 110 | 1440 | 97 | 90806311212 | 99.70% |
| BG | 2092 | 16765 | 937 | 5275003924055 | 94.86% |
| BH | 253 | 3870 | 181 | 401335927611 | 98.53% |
| BI | 48 | 158 | 58 | 26163020068 | 92.83% |
| BJ | 71 | 612 | 75 | 56227964972 | 96.66% |
| BL | 20 | 6 | 46 | 324274212 | 94.31% |
| BM | 311 | 531 | 77 | 47319157040 | 95.95% |
| BN | 96 | 864 | 75 | 25912016939 | 97.41% |
| BO | 220 | 4717 | 154 | 167777141067 | 97.97% |
| BQ | 32 | 103 | 50 | 21664565292 | 91.61% |
| BR | 9233 | 325866 | 13571 | 37118307497418 | 99.47% |
| BS | 117 | 615 | 84 | 13262854710 | 99.74% |
| BT | 112 | 218 | 111 | 60743156035 | 99.54% |
| BV | 0 | 0 | 7 | 17956881 | 50.89% |
| BW | 152 | 692 | 82 | 78140166184 | 97.30% |
| BY | 300 | 7366 | 211 | 699985757087 | 99.36% |
| BZ | 242 | 482 | 166 | 197206083904 | 91.01% |
| CA | 14843 | 281074 | 11200 | 6424395454934 | 94.64% |
| CC | 2 | 0 | 20 | 120587282 | 56.02% |
| CD | 226 | 892 | 136 | 155168849996 | 99.17% |
| CF | 41 | 36 | 52 | 578179105 | 54.60% |
| CG | 63 | 573 | 67 | 36380110884 | 97.77% |
| CH | 5609 | 96086 | 54831 | 15878799854307456248 | 66.67% |
| CI | 334 | 6715 | 96 | 194272192552 | 98.90% |
| CK | 30 | 41 | 50 | 8780383271 | 98.87% |
| CL | 1477 | 40281 | 1063 | 1406096927110 | 97.18% |
| CM | 175 | 2428 | 105 | 68227131459 | 98.35% |
| CN | 6554 | 1349219 | 5603 | 294842677950218 | 99.93% |
| CO | 1894 | 69870 | 1479 | 1954366696579 | 98.36% |
| CR | 411 | 9132 | 241 | 461494949981 | 98.02% |
| CU | 78 | 1014 | 65 | 26252017957 | 99.29% |
| CV | 70 | 234 | 103 | 13477349418 | 98.02% |
| CW | 94 | 736 | 64 | 62334240802 | 94.80% |
| CX | 2 | 0 | 17 | 36767762 | 65.20% |
| CY | 651 | 4542 | 720 | 8214973413691 | 88.17% |
| CZ | 3058 | 34224 | 2355 | 13205706219440 | 98.14% |
| DE | 27700 | 529823 | 67423 | 110673184406486 | 97.16% |
| DJ | 39 | 229 | 68 | 9084602401 | 97.61% |
| DK | 3184 | 49403 | 1719 | 9503799484498 | 93.16% |
| DM | 59 | 64 | 71 | 43409606948 | 75.03% |
| DO | 542 | 7966 | 554 | 923710983592 | 99.20% |
| DZ | 187 | 18707 | 116 | 43387848836 | 94.30% |
| EC | 614 | 11183 | 516 | 1153796885022 | 99.03% |
| EE | 1106 | 6357 | 1132 | 3099965854981 | 91.37% |
| EG | 480 | 94707 | 165 | 17730890711404 | 99.98% |
| EH | 2 | 0 | 23 | 270927890 | 47.33% |
| ER | 34 | 21 | 55 | 409424933 | 59.00% |
| ES | 9364 | 137667 | 3483 | 25737254288435 | 97.98% |
| ET | 62 | 1462 | 77 | 1117305116715 | 99.97% |
| FI | 3759 | 58134 | 34819 | 6805619656109 | 97.11% |
| FJ | 99 | 588 | 77 | 43175249703 | 96.89% |
| FK | 27 | 21 | 50 | 89786400 | 60.58% |
| FM | 46 | 41 | 62 | 8783661104 | 92.34% |
| FO | 40 | 178 | 60 | 107547527464 | 99.93% |
| FR | 22845 | 327853 | 6728 | 66942336095263 | 98.69% |
| GA | 70 | 1546 | 60 | 35306035233 | 93.90% |
| GB | 30930 | 503849 | 24837 | 105652261005840 | 92.18% |
| GD | 91 | 113 | 65 | 77687097638 | 99.97% |
| GE | 448 | 5074 | 272 | 1863572723514 | 94.73% |
| GF | 45 | 98 | 57 | 27118469156 | 90.58% |
| GG | 112 | 291 | 63 | 146286577953 | 89.84% |
| GH | 310 | 8887 | 179 | 1233649642838 | 99.69% |
| GI | 126 | 412 | 76 | 197841520676 | 90.16% |
| GL | 97 | 144 | 114 | 43164827754 | 99.76% |
| GM | 63 | 1094 | 64 | 39097358372 | 99.54% |
| GN | 62 | 166 | 78 | 47788005412 | 98.20% |
| GP | 212 | 444 | 66 | 49599961137 | 71.65% |
| GQ | 46 | 88 | 59 | 9032980519 | 97.98% |
| GR | 1204 | 23131 | 508 | 2562843935588 | 96.02% |
| GS | 3 | 2 | 20 | 237436946 | 47.81% |
| GT | 775 | 2733 | 190 | 326869309810 | 99.62% |
| GU | 132 | 958 | 81 | 34436317236 | 99.93% |
| GW | 35 | 32 | 60 | 9133512743 | 97.37% |
| GY | 47 | 284 | 68 | 34551367720 | 97.27% |
| HK | 9439 | 82121 | 6808 | 6592555143004 | 76.05% |
| HM | 1 | 2 | 8 | 34734097 | 53.28% |
| HN | 927 | 1968 | 223 | 421646731630 | 93.53% |
| HR | 621 | 9607 | 387 | 1651588345566 | 94.75% |
| HT | 66 | 609 | 82 | 25978800429 | 95.45% |
| HU | 1475 | 23204 | 793 | 3305855693772 | 97.35% |
| ID | 6391 | 82117 | 4067 | 3773736268106 | 97.86% |
| IE | 3977 | 70745 | 5718 | 5002992610809 | 89.15% |
| IL | 1952 | 34802 | 1885 | 4410642512833 | 96.11% |
| IM | 218 | 544 | 161 | 399740963361 | 90.67% |
| IN | 12039 | 190667 | 9744 | 29372938530112 | 97.00% |
| IO | 20 | 14 | 49 | 4552329249 | 97.06% |
| IQ | 603 | 3283 | 335 | 3338011283615 | 97.55% |
| IR | 1943 | 42855 | 771 | 17194259644965 | 98.05% |
| IS | 302 | 3630 | 320 | 1566998093375 | 96.88% |
| IT | 8459 | 219173 | 5062 | 47165688951015 | 98.94% |
| JE | 113 | 639 | 154 | 219217594660 | 97.58% |
| JM | 224 | 1093 | 97 | 1121381009741 | 100.00% |
| JO | 282 | 2564 | 157 | 1121477593416 | 96.96% |
| JP | 9667 | 774756 | 6628 | 51625635697279 | 89.66% |
| KE | 757 | 24834 | 364 | 667087046056 | 98.83% |
| KG | 227 | 1186 | 116 | 528758868549 | 99.45% |
| KH | 464 | 1977 | 204 | 297190752596 | 95.20% |
| KI | 66 | 28 | 87 | 13081249859 | 95.12% |
| KM | 50 | 43 | 52 | 9166979106 | 97.15% |
| KN | 63 | 114 | 58 | 34837104688 | 57.22% |
| KP | 4 | 7 | 263 | 530972930 | 53.52% |
| KR | 2819 | 452510 | 1664 | 22740571629858 | 99.66% |
| KW | 253 | 7501 | 131 | 872664969269 | 99.88% |
| KY | 126 | 429 | 62 | 17675389223 | 99.35% |
| KZ | 1229 | 13881 | 568 | 2466446299761 | 96.96% |
| LA | 131 | 410 | 106 | 103609467173 | 87.35% |
| LB | 362 | 2318 | 208 | 2981241307442 | 99.05% |
| LC | 73 | 155 | 53 | 81963452448 | 97.18% |
| LI | 135 | 427 | 118 | 425531152932 | 88.08% |
| LK | 207 | 2292 | 135 | 73451339834 | 96.04% |
| LR | 63 | 170 | 71 | 26232182823 | 98.19% |
| LS | 56 | 496 | 58 | 26162499620 | 99.39% |
| LT | 1380 | 12346 | 1108 | 5417152263659 | 81.15% |
| LU | 1229 | 6948 | 920 | 2152080395619 | 96.63% |
| LV | 1087 | 8158 | 468 | 1504492222824 | 94.25% |
| LY | 130 | 1904 | 79 | 99411972151 | 92.15% |
| MA | 287 | 48059 | 169 | 1165734908287 | 99.92% |
| MC | 82 | 277 | 70 | 146307287073 | 98.62% |
| MD | 579 | 4968 | 947 | 6532981608206 | 69.89% |
| ME | 111 | 791 | 93 | 343957855534 | 99.95% |
| MF | 31 | 30 | 45 | 4652468260 | 99.61% |
| MG | 76 | 2299 | 72 | 22220656940 | 98.67% |
| MH | 43 | 128 | 44 | 4368433188 | 99.49% |
| MK | 231 | 2766 | 169 | 1134250904933 | 95.88% |
| ML | 72 | 378 | 57 | 21917728810 | 99.17% |
| MM | 354 | 1099 | 265 | 352698110290 | 97.26% |
| MN | 210 | 732 | 155 | 133729894462 | 99.60% |
| MO | 86 | 1403 | 341 | 38884803104 | 98.28% |
| MP | 41 | 39 | 47 | 4350935076 | 99.57% |
| MQ | 73 | 728 | 46 | 112933734432 | 87.75% |
| MR | 54 | 216 | 57 | 26162300965 | 99.41% |
| MS | 21 | 8 | 46 | 189794343 | 47.34% |
| MT | 241 | 2658 | 131 | 466260248872 | 96.91% |
| MU | 1765 | 26589 | 8067 | 4455234701053991 | 100.00% |
| MV | 115 | 385 | 92 | 17742872614 | 98.57% |
| MW | 97 | 2190 | 108 | 1241843060778 | 99.89% |
| MX | 3157 | 118799 | 1864 | 3528411947515 | 97.73% |
| MY | 2018 | 30434 | 1077 | 900803628596 | 95.26% |
| MZ | 190 | 1836 | 84 | 69179559991 | 97.25% |
| NA | 220 | 1875 | 65 | 29231943721 | 98.37% |
| NC | 67 | 571 | 77 | 47438497072 | 99.79% |
| NE | 50 | 208 | 55 | 21633695777 | 99.70% |
| NF | 26 | 3 | 45 | 189925412 | 46.24% |
| NG | 836 | 12794 | 458 | 632974083658 | 96.06% |
| NI | 137 | 1700 | 79 | 94699850038 | 98.75% |
| NL | 18264 | 213245 | 160887 | 8409242871444436 | 99.93% |
| NO | 2572 | 63921 | 1430 | 10872747525355 | 95.80% |
| NP | 405 | 2368 | 283 | 430168891476 | 92.26% |
| NR | 30 | 47 | 46 | 8780842020 | 98.87% |
| NU | 23 | 11 | 49 | 4501865508 | 97.61% |
| NZ | 1928 | 29011 | 1211 | 1338763052635 | 96.05% |
| OM | 278 | 4308 | 123 | 586559012911 | 96.43% |
| PA | 486 | 7162 | 443 | 463007246689 | 90.74% |
| PE | 877 | 14151 | 710 | 1064353037741 | 97.25% |
| PF | 76 | 309 | 107 | 21704935505 | 99.46% |
| PG | 164 | 327 | 109 | 73261973822 | 91.28% |
| PH | 1857 | 26024 | 1002 | 1033861823885 | 97.02% |
| PK | 1225 | 23365 | 630 | 898593198233 | 97.30% |
| PL | 7551 | 82033 | 3873 | 28353567616613 | 97.14% |
| PM | 35 | 23 | 46 | 4484891940 | 97.79% |
| PN | 3 | 2 | 15 | 52757522 | 66.54% |
| PR | 523 | 4437 | 229 | 223265028249 | 97.60% |
| PS | 288 | 3319 | 125 | 979536320571 | 95.42% |
| PT | 1468 | 26828 | 864 | 3262418376830 | 93.20% |
| PW | 36 | 31 | 44 | 8646819876 | 99.77% |
| PY | 331 | 5032 | 209 | 451204891005 | 97.89% |
| QA | 246 | 5065 | 253 | 521284496716 | 93.90% |
| RE | 94 | 1652 | 59 | 129543372843 | 84.06% |
| RO | 3745 | 29211 | 1771 | 5683748101002 | 94.84% |
| RS | 870 | 9094 | 822 | 1826935074856 | 99.14% |
| RU | 12323 | 180378 | 7357 | 49276768264785 | 93.38% |
| RW | 92 | 1188 | 81 | 1147200098343 | 99.86% |
| SA | 1218 | 46104 | 680 | 2731801104168 | 98.01% |
| SB | 43 | 62 | 57 | 17253860644 | 98.29% |
| SC | 3041 | 15005 | 3804 | 8716368052771 | 60.14% |
| SD | 88 | 7415 | 70 | 69232295979 | 97.24% |
| SE | 8395 | 125066 | 5692 | 30534759432578 | 86.42% |
| SG | 7815 | 112442 | 4975 | 149323835802090 | 97.29% |
| SH | 3 | 2 | 27 | 120326171 | 58.18% |
| SI | 885 | 10375 | 641 | 2220829691774 | 93.59% |
| SJ | 4 | 2 | 22 | 238422034 | 49.66% |
| SK | 879 | 10948 | 580 | 2909546950124 | 97.30% |
| SL | 74 | 360 | 70 | 26246449194 | 98.71% |
| SM | 55 | 161 | 69 | 197812815905 | 92.88% |
| SN | 89 | 1614 | 98 | 26448648253 | 98.84% |
| SO | 54 | 213 | 75 | 47738605862 | 99.57% |
| SR | 43 | 336 | 70 | 17391550764 | 99.48% |
| SS | 75 | 128 | 72 | 30625849384 | 99.17% |
| ST | 32 | 50 | 58 | 4587128868 | 96.84% |
| SV | 308 | 2750 | 140 | 210718525798 | 97.90% |
| SX | 45 | 146 | 53 | 30120740898 | 96.37% |
| SY | 271 | 5143 | 99 | 876791795755 | 98.55% |
| SZ | 58 | 227 | 63 | 30172074032 | 99.44% |
| TC | 35 | 50 | 47 | 326240548 | 94.18% |
| TD | 61 | 117 | 62 | 13093177381 | 99.35% |
| TF | 2 | 0 | 14 | 170197010 | 43.55% |
| TG | 56 | 1396 | 79 | 26196072485 | 99.34% |
| TH | 1844 | 40454 | 1196 | 589301540809 | 94.62% |
| TJ | 130 | 380 | 104 | 670826039853 | 96.61% |
| TK | 21 | 14 | 51 | 4350871591 | 99.58% |
| TL | 80 | 119 | 67 | 43426578471 | 97.26% |
| TM | 52 | 120 | 54 | 68843668777 | 99.92% |
| TN | 125 | 30790 | 94 | 112249553196 | 99.76% |
| TO | 37 | 48 | 50 | 17371037731 | 99.43% |
| TR | 2972 | 67614 | 2368 | 23572077025981 | 89.85% |
| TT | 133 | 2196 | 91 | 1177065703738 | 77.91% |
| TV | 29 | 38 | 59 | 12388205603 | 81.28% |
| TW | 1904 | 144859 | 1845 | 11515624277827 | 98.42% |
| TZ | 315 | 4425 | 170 | 266872406120 | 98.61% |
| UA | 4500 | 34881 | 2386 | 11588388669696 | 93.13% |
| UG | 202 | 5666 | 103 | 125050314796 | 97.77% |
| UM | 3 | 2 | 33 | 188024850 | 46.40% |
| US | 98871 | 5982403 | 168590 | 4836249701322909 | 99.78% |
| UY | 186 | 10325 | 122 | 177250959399 | 99.85% |
| UZ | 421 | 1597 | 176 | 1083226921524 | 99.47% |
| VA | 23 | 43 | 54 | 103155566879 | 99.97% |
| VC | 63 | 86 | 47 | 69044865057 | 99.97% |
| VE | 714 | 29812 | 475 | 2252982727044 | 99.46% |
| VG | 237 | 2945 | 386 | 542501676071 | 68.88% |
| VI | 72 | 537 | 107 | 30127712303 | 94.55% |
| VN | 1967 | 65220 | 1417 | 1902198653792 | 97.87% |
| VU | 52 | 64 | 245 | 51781306916 | 87.14% |
| WF | 19 | 16 | 43 | 4350543908 | 99.55% |
| WS | 49 | 88 | 60 | 21682979111 | 91.59% |
| XK | 7 | 6 | 104 | 3489071108 | 46.67% |
| YE | 97 | 940 | 83 | 206625898548 | 99.92% |
| YT | 30 | 15 | 49 | 181469220 | 59.04% |
| ZA | 3793 | 114023 | 1544 | 23426303469187 | 99.40% |
| ZM | 144 | 6393 | 87 | 43460284738 | 95.55% |
| ZW | 188 | 693 | 104 | 65341640792 | 98.53% |

<!-- GEOIP-STATISTICS:END -->
