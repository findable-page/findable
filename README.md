# seoscore-cli

Scan a site from the terminal with [seoscore.tools](https://seoscore.tools) and save the same CSV the website offers. It also prints the scores and every failing check.

Unofficial: it calls the public `/api/scan` endpoint the website uses and is not affiliated with seoscore.tools. The endpoint is undocumented and may change or rate-limit you.

## Install

```
go install github.com/nadimtuhin/seoscore-cli@latest
```

## Use

```
seoscore-cli [-o dir] [-q] https://example.com
```

- `-o dir`: where to save `seoscore-report-<host>-<date>.csv` (default `.`, created if missing)
- `-q`: do not list failing checks

Output:

```
overall 89  SEO 90  AEO 90  GEO 84
FAIL SEO compression: No compression detected (gzip/brotli recommended)
saved ./seoscore-report-example.com-2026-10-10.csv
```

On a 429 it waits 10 seconds and retries, up to 3 times. No dependencies beyond the Go standard library.

## License

MIT
