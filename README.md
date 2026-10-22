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

Last run: `2026-10-22T17:47:25Z`

Input rows: **31006976** · Published blocks: **2085463**

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
| IPinfo via xream MMDB | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPINFO Lite CSV | cached | 1341028 | 2936383 | 0 | 0 | 208702865 | local cache |
| IP2Location IPv4 | cached | 390273 | 0 | 0 | 0 | 6100911 | local cache |
| IP2Location IPv6 | cached | 0 | 1512762 | 0 | 0 | 24574212 | local cache |
| IP2Location Lite via xream MMDB | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| IPLocate | cached | 785303 | 1368186 | 0 | 0 | 17235320 | local cache |
| IPGeo | cached | 177093 | 69680 | 0 | 0 | 3393343 | local cache |
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
| gaoyifan China operator IPv4+IPv6 | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| Iran CIDRs by parsamoh | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |
| GeoLite2 Country via P3TERX | skipped-no-usable-copy | 0 | 0 | 0 | 0 | 0 | — |

### Assignment evidence

| Method | IPv4 blocks | IPv4 addresses | IPv6 blocks | IPv6 addresses |
|---|---:|---:|---:|---:|
| geolocation | 586251 | 3702257408 | 1498962 | 293421847207842436812546030166054273024 |
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
| AD | 209 | 307 | 173 | 43696077127 | 91.65% |
| AE | 2790 | 25437 | 2153 | 12757364277811 | 90.02% |
| AF | 229 | 877 | 478 | 120984778118 | 97.43% |
| AG | 116 | 174 | 171 | 40374450229 | 65.25% |
| AI | 44 | 46 | 87 | 5292833845 | 70.16% |
| AL | 559 | 1425 | 1223 | 2496064906649 | 95.29% |
| AM | 331 | 2315 | 502 | 1258811614671 | 94.73% |
| AO | 560 | 5115 | 256 | 143002042966 | 98.76% |
| AQ | 56 | 6 | 172 | 132462390 | 43.80% |
| AR | 3713 | 76922 | 2703 | 23557626049837 | 98.88% |
| AS | 39 | 28 | 319 | 8703129798 | 99.49% |
| AT | 4622 | 45074 | 3906 | 10433591217649 | 97.53% |
| AU | 12853 | 219020 | 7856 | 2305920503842975826 | 100.00% |
| AW | 47 | 441 | 82 | 10844698674 | 96.27% |
| AX | 108 | 136 | 116 | 68848072778 | 74.97% |
| AZ | 390 | 3065 | 322 | 1933582840439 | 97.50% |
| BA | 331 | 3262 | 722 | 799398505753 | 94.87% |
| BB | 190 | 508 | 89 | 69113215274 | 99.94% |
| BD | 2623 | 8165 | 2352 | 3572993537320 | 96.16% |
| BE | 8311 | 58209 | 72624 | 8440515390332721 | 66.66% |
| BF | 122 | 1382 | 128 | 90656114996 | 99.80% |
| BG | 2286 | 16710 | 1326 | 5167782274363 | 95.76% |
| BH | 315 | 5242 | 237 | 401358418511 | 98.11% |
| BI | 56 | 167 | 94 | 26180137510 | 93.89% |
| BJ | 81 | 621 | 101 | 56278702382 | 97.14% |
| BL | 25 | 10 | 68 | 357775402 | 90.27% |
| BM | 317 | 497 | 101 | 43274975286 | 99.54% |
| BN | 105 | 877 | 104 | 25896153133 | 96.49% |
| BO | 334 | 4758 | 223 | 167763587434 | 97.39% |
| BQ | 36 | 103 | 62 | 17268488240 | 99.80% |
| BR | 11379 | 336004 | 25258 | 37192992749656 | 99.30% |
| BS | 135 | 632 | 145 | 13300271438 | 99.64% |
| BT | 121 | 225 | 149 | 60591845967 | 99.69% |
| BV | 4 | 1 | 3 | 1048577 | 50.00% |
| BW | 162 | 717 | 110 | 78258733358 | 97.75% |
| BY | 404 | 7402 | 406 | 699976659758 | 99.48% |
| BZ | 204 | 475 | 187 | 192427129410 | 89.09% |
| CA | 18895 | 285687 | 28509 | 6334462785768 | 95.47% |
| CC | 4 | 0 | 33 | 86915091 | 51.88% |
| CD | 263 | 933 | 171 | 155191165264 | 98.94% |
| CF | 47 | 40 | 83 | 445218856 | 59.51% |
| CG | 88 | 585 | 105 | 36399194159 | 98.06% |
| CH | 6738 | 98325 | 55642 | 15878801126182848386 | 99.98% |
| CI | 344 | 6837 | 145 | 194583681333 | 99.02% |
| CK | 33 | 41 | 73 | 8679668781 | 99.60% |
| CL | 1984 | 41384 | 1967 | 1406348987782 | 96.81% |
| CM | 188 | 2438 | 136 | 67562477642 | 98.44% |
| CN | 6898 | 1348026 | 12944 | 294793112425636 | 99.94% |
| CO | 3117 | 70141 | 1900 | 1972135150465 | 97.73% |
| CR | 590 | 9171 | 365 | 461401033574 | 97.39% |
| CU | 96 | 1021 | 115 | 26153123897 | 99.59% |
| CV | 79 | 240 | 124 | 13360838699 | 98.60% |
| CW | 140 | 720 | 94 | 62436227115 | 93.00% |
| CX | 5 | 0 | 63 | 54343716 | 56.39% |
| CY | 661 | 4143 | 669 | 5977213200480 | 94.02% |
| CZ | 3604 | 33868 | 5644 | 13281497553810 | 97.65% |
| DE | 36764 | 551177 | 69870 | 115055906446793 | 96.07% |
| DJ | 68 | 260 | 102 | 9069606957 | 97.36% |
| DK | 3628 | 50138 | 1843 | 9662936666510 | 93.33% |
| DM | 77 | 71 | 98 | 43846023466 | 73.59% |
| DO | 724 | 8062 | 628 | 923641087353 | 99.14% |
| DZ | 223 | 18725 | 224 | 43689261238 | 94.41% |
| EC | 1118 | 11253 | 661 | 1153873628758 | 98.71% |
| EE | 1167 | 6517 | 1596 | 2725232707354 | 96.04% |
| EG | 774 | 94833 | 272 | 17730854596757 | 99.98% |
| EH | 4 | 0 | 36 | 53426198 | 50.45% |
| ER | 44 | 29 | 78 | 393644071 | 58.57% |
| ES | 11083 | 139845 | 4999 | 26194219060496 | 97.13% |
| ET | 86 | 1478 | 108 | 1117176612916 | 99.98% |
| FI | 4003 | 58806 | 24936 | 6835600976573 | 96.84% |
| FJ | 110 | 594 | 105 | 43076631338 | 97.40% |
| FK | 38 | 26 | 67 | 106838052 | 57.54% |
| FM | 56 | 49 | 95 | 8683534393 | 89.69% |
| FO | 46 | 181 | 70 | 107480660266 | 99.96% |
| FR | 26944 | 325492 | 13882 | 67328880282630 | 98.50% |
| GA | 86 | 1559 | 88 | 35274197285 | 95.01% |
| GB | 39663 | 499072 | 49689 | 105969251371563 | 94.11% |
| GD | 100 | 128 | 88 | 82015828264 | 97.39% |
| GE | 491 | 5105 | 363 | 1797094278256 | 95.68% |
| GF | 193 | 287 | 66 | 22655610918 | 98.22% |
| GG | 140 | 264 | 81 | 146151913772 | 85.57% |
| GH | 361 | 8926 | 239 | 1227763932767 | 99.98% |
| GI | 134 | 327 | 97 | 197655476265 | 91.29% |
| GL | 106 | 145 | 158 | 43065946234 | 99.91% |
| GM | 74 | 1102 | 95 | 39081527593 | 99.58% |
| GN | 73 | 172 | 99 | 47654127655 | 97.86% |
| GP | 202 | 799 | 90 | 53853128757 | 73.30% |
| GQ | 54 | 95 | 74 | 8982923303 | 98.22% |
| GR | 1522 | 23126 | 836 | 2587659616597 | 96.52% |
| GS | 4 | 2 | 27 | 69283859 | 56.22% |
| GT | 660 | 2886 | 269 | 327213697398 | 99.50% |
| GU | 187 | 991 | 113 | 34505109817 | 99.85% |
| GW | 48 | 42 | 81 | 8999700521 | 98.22% |
| GY | 57 | 288 | 94 | 34518546476 | 95.66% |
| HK | 13112 | 95112 | 7493 | 5159591149454 | 85.91% |
| HM | 1 | 2 | 7 | 17965056 | 49.38% |
| HN | 827 | 1992 | 256 | 391890041209 | 94.75% |
| HR | 842 | 9695 | 411 | 1587087961808 | 96.47% |
| HT | 84 | 623 | 121 | 25913080126 | 94.26% |
| HU | 1917 | 23377 | 1185 | 3212512443884 | 98.26% |
| ID | 6854 | 85399 | 5024 | 3788769755256 | 98.47% |
| IE | 7961 | 72003 | 7709 | 4921893746438 | 88.90% |
| IL | 2366 | 39289 | 1419 | 4292366361556 | 97.51% |
| IM | 308 | 475 | 179 | 360901130812 | 96.22% |
| IN | 13294 | 197157 | 13773 | 28992080827638 | 97.85% |
| IO | 33 | 15 | 68 | 4384569384 | 99.10% |
| IQ | 675 | 3340 | 948 | 3415366255501 | 96.92% |
| IR | 1962 | 42732 | 2184 | 16790237825904 | 98.75% |
| IS | 337 | 3641 | 485 | 1605249929335 | 96.20% |
| IT | 9841 | 227815 | 6984 | 46956275740488 | 99.10% |
| JE | 197 | 639 | 104 | 219167521067 | 95.54% |
| JM | 245 | 1130 | 137 | 1121437104726 | 99.99% |
| JO | 440 | 2705 | 221 | 1121447786590 | 97.92% |
| JP | 12360 | 782267 | 10517 | 52707700385796 | 88.45% |
| KE | 1137 | 25121 | 553 | 666845280434 | 98.85% |
| KG | 229 | 1173 | 191 | 528753310054 | 99.29% |
| KH | 542 | 2007 | 290 | 434667663975 | 81.12% |
| KI | 77 | 35 | 119 | 12981254217 | 94.20% |
| KM | 59 | 51 | 71 | 8999743523 | 98.19% |
| KN | 114 | 164 | 92 | 5477579062 | 78.71% |
| KP | 9 | 11 | 317 | 429047839 | 56.24% |
| KR | 3911 | 455193 | 2603 | 22750774215248 | 99.64% |
| KW | 415 | 7566 | 218 | 872555832223 | 99.85% |
| KY | 125 | 442 | 95 | 17576311857 | 99.74% |
| KZ | 1318 | 13959 | 1122 | 3150237089541 | 86.55% |
| LA | 146 | 451 | 138 | 120672563501 | 88.36% |
| LB | 389 | 2386 | 276 | 3015572156995 | 98.09% |
| LC | 89 | 167 | 76 | 77702314274 | 97.97% |
| LI | 276 | 453 | 191 | 429629589799 | 90.76% |
| LK | 286 | 2327 | 249 | 73495245174 | 96.30% |
| LR | 79 | 438 | 94 | 26234620201 | 96.57% |
| LS | 69 | 507 | 83 | 26164084775 | 99.39% |
| LT | 1450 | 12207 | 896 | 4380847985354 | 84.57% |
| LU | 1548 | 6702 | 853 | 2187790107243 | 96.51% |
| LV | 1198 | 8173 | 715 | 3226865891543 | 65.69% |
| LY | 147 | 1923 | 100 | 99245719610 | 89.71% |
| MA | 342 | 48128 | 253 | 1166467875493 | 99.93% |
| MC | 281 | 284 | 99 | 141897250860 | 99.96% |
| MD | 758 | 4954 | 954 | 5927228702283 | 64.00% |
| ME | 148 | 823 | 118 | 343757130807 | 99.98% |
| MF | 57 | 106 | 72 | 4686936105 | 99.25% |
| MG | 85 | 2308 | 107 | 22190064179 | 99.00% |
| MH | 61 | 90 | 72 | 4386730024 | 99.24% |
| MK | 250 | 2795 | 181 | 1138535841131 | 96.26% |
| ML | 83 | 387 | 78 | 21902405675 | 99.20% |
| MM | 455 | 1161 | 324 | 352716403067 | 97.51% |
| MN | 241 | 738 | 266 | 133601259628 | 99.33% |
| MO | 103 | 1499 | 687 | 38813397303 | 99.80% |
| MP | 86 | 18 | 64 | 4367396906 | 99.40% |
| MQ | 242 | 876 | 61 | 104261957662 | 89.90% |
| MR | 70 | 232 | 82 | 26180073512 | 99.39% |
| MS | 30 | 15 | 61 | 72104234 | 63.51% |
| MT | 316 | 2677 | 214 | 492267714868 | 94.44% |
| MU | 1882 | 23875 | 8065 | 4455225967952187 | 100.00% |
| MV | 117 | 401 | 128 | 17576620085 | 99.11% |
| MW | 111 | 2199 | 136 | 1241710428206 | 99.87% |
| MX | 4200 | 122248 | 2605 | 3621241233527 | 96.81% |
| MY | 2811 | 34671 | 1468 | 911083590554 | 95.60% |
| MZ | 184 | 1911 | 120 | 69181407290 | 97.67% |
| NA | 389 | 1857 | 178 | 28832514347 | 99.83% |
| NC | 74 | 571 | 97 | 47606413618 | 99.74% |
| NE | 61 | 217 | 76 | 21634035747 | 99.71% |
| NF | 31 | 3 | 67 | 72431660 | 55.82% |
| NG | 1115 | 12938 | 795 | 599999707261 | 98.60% |
| NI | 175 | 1731 | 109 | 94884804670 | 98.33% |
| NL | 23358 | 215053 | 466332 | 8409476976353940 | 99.94% |
| NO | 3190 | 63876 | 2666 | 10917411554647 | 97.15% |
| NP | 469 | 2386 | 332 | 430193086562 | 93.62% |
| NR | 40 | 53 | 76 | 8648144939 | 99.78% |
| NU | 27 | 11 | 64 | 4385093671 | 99.20% |
| NZ | 2262 | 29281 | 1704 | 1355635949833 | 95.46% |
| OM | 302 | 4417 | 188 | 590674735188 | 96.31% |
| PA | 511 | 7098 | 442 | 428939568437 | 92.58% |
| PE | 1250 | 14834 | 907 | 1068014497048 | 96.40% |
| PF | 88 | 312 | 131 | 21587703897 | 99.83% |
| PG | 262 | 348 | 142 | 73179476297 | 91.89% |
| PH | 2225 | 26770 | 1278 | 1046299111723 | 97.34% |
| PK | 1417 | 23647 | 716 | 902713327862 | 97.61% |
| PL | 8509 | 83045 | 4980 | 28331568076827 | 97.18% |
| PM | 44 | 25 | 64 | 4384045097 | 99.21% |
| PN | 5 | 2 | 30 | 85997588 | 60.25% |
| PR | 633 | 4577 | 334 | 230250195342 | 99.12% |
| PS | 306 | 3297 | 165 | 979389139277 | 93.84% |
| PT | 1903 | 27025 | 1688 | 3328234641911 | 92.92% |
| PW | 48 | 43 | 70 | 8680517675 | 99.56% |
| PY | 443 | 5087 | 260 | 451262837113 | 97.01% |
| QA | 462 | 5170 | 306 | 521158782425 | 96.61% |
| RE | 239 | 1981 | 85 | 95150534704 | 96.67% |
| RO | 3962 | 28316 | 2340 | 5569915178061 | 95.29% |
| RS | 1040 | 9070 | 1203 | 1848638121896 | 98.42% |
| RU | 13453 | 180638 | 14983 | 54419090992151 | 90.25% |
| RW | 139 | 1265 | 110 | 1147205026090 | 99.92% |
| SA | 1637 | 46658 | 884 | 2810929095910 | 96.19% |
| SB | 54 | 71 | 75 | 17254136100 | 95.73% |
| SC | 923 | 2664 | 3300 | 3363307844135 | 58.94% |
| SD | 114 | 7412 | 95 | 69232636209 | 97.68% |
| SE | 15811 | 126261 | 47169 | 30350916706873 | 79.88% |
| SG | 10488 | 107653 | 51519 | 151201742679893 | 97.13% |
| SH | 6 | 2 | 112 | 119420929 | 55.94% |
| SI | 1075 | 10322 | 648 | 2109447955119 | 95.68% |
| SJ | 7 | 2 | 29 | 70465552 | 64.53% |
| SK | 1211 | 11162 | 813 | 2976077605380 | 97.28% |
| SL | 126 | 356 | 93 | 26213562412 | 97.96% |
| SM | 70 | 172 | 89 | 201991534627 | 92.53% |
| SN | 119 | 1644 | 127 | 26840369219 | 98.53% |
| SO | 73 | 224 | 135 | 47695102270 | 99.62% |
| SR | 51 | 341 | 91 | 17308332336 | 99.73% |
| SS | 86 | 134 | 95 | 30492233770 | 99.43% |
| ST | 47 | 60 | 74 | 4402657320 | 99.02% |
| SV | 345 | 2774 | 196 | 211144269273 | 97.15% |
| SX | 44 | 129 | 65 | 30153979938 | 95.13% |
| SY | 294 | 5166 | 114 | 842252095534 | 99.64% |
| SZ | 76 | 242 | 90 | 29904044085 | 99.85% |
| TC | 45 | 56 | 69 | 1166129191 | 73.92% |
| TD | 89 | 132 | 86 | 13043382315 | 99.54% |
| TF | 2 | 0 | 26 | 35991568 | 51.03% |
| TG | 71 | 1402 | 96 | 26179635492 | 99.37% |
| TH | 2236 | 43547 | 1199 | 591270493043 | 95.65% |
| TJ | 148 | 393 | 143 | 636421224763 | 98.15% |
| TK | 24 | 15 | 98 | 4385290286 | 99.19% |
| TL | 88 | 126 | 92 | 43326320684 | 96.66% |
| TM | 63 | 126 | 86 | 68778276917 | 99.96% |
| TN | 149 | 30817 | 132 | 112367071804 | 99.77% |
| TO | 45 | 55 | 76 | 17270648872 | 99.79% |
| TR | 3652 | 68019 | 2810 | 22452311283528 | 93.73% |
| TT | 167 | 2226 | 136 | 1176970295879 | 70.56% |
| TV | 41 | 46 | 79 | 10843928612 | 89.70% |
| TW | 2515 | 147477 | 2954 | 11501135107153 | 98.76% |
| TZ | 372 | 4485 | 227 | 266778178675 | 98.85% |
| UA | 4936 | 34857 | 3189 | 11792756366720 | 92.94% |
| UG | 258 | 5773 | 158 | 131500446011 | 95.50% |
| UM | 5 | 2 | 61 | 36517929 | 67.31% |
| US | 130338 | 5881521 | 380736 | 4836265365357929 | 99.75% |
| UY | 244 | 10382 | 179 | 177236963394 | 99.89% |
| UZ | 555 | 1660 | 252 | 1083081836906 | 99.37% |
| VA | 35 | 51 | 100 | 103192213544 | 99.95% |
| VC | 82 | 104 | 73 | 69078956069 | 99.95% |
| VE | 1000 | 30315 | 685 | 2248981881348 | 99.35% |
| VG | 329 | 554 | 371 | 255189948461 | 87.01% |
| VI | 83 | 532 | 170 | 30179888195 | 92.97% |
| VN | 2252 | 65537 | 1636 | 1851551343628 | 99.38% |
| VU | 65 | 86 | 271 | 51680983854 | 79.92% |
| WF | 22 | 16 | 56 | 4350228519 | 99.59% |
| WS | 60 | 97 | 81 | 21582524712 | 93.19% |
| XK | 177 | 620 | 73 | 111670788087 | 60.62% |
| YE | 115 | 958 | 108 | 206626305081 | 99.92% |
| YT | 43 | 44 | 60 | 131084326 | 62.63% |
| ZA | 5130 | 118014 | 2739 | 23426041685464 | 99.56% |
| ZM | 244 | 6508 | 130 | 43700027728 | 96.15% |
| ZW | 211 | 932 | 151 | 65246823779 | 98.51% |

<!-- GEOIP-STATISTICS:END -->
