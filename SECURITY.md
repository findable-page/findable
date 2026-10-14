# Security Policy

Report vulnerabilities privately through [GitHub security advisories](https://github.com/nadimtuhin/findable/security/advisories/new). Do not open a public issue.

Expect a first reply within 7 days. Only the latest release gets fixes.

`findable -serve` is meant for localhost. Do not expose it to the internet.

## SSRF policy

`POST /api/scan` on the edge API (https://findable.piply.xyz) fetches a URL chosen by a stranger, so it only fetches public pages:

- Blocks private and reserved addresses: loopback, link-local, RFC 1918, carrier-grade NAT, benchmarking and `0.0.0.0/8`, plus the IPv6 forms that can hide an IPv4 target (6to4 `2002::/16`, Teredo `2001::/32`, NAT64 `64:ff9b::/96`, IPv4-mapped).
- Checks hostname syntax only: IP literals, `localhost`, single-label names and the `.local`, `.internal`, `.localhost` and `.home.arpa` suffixes, including names written with a trailing dot. Workers cannot resolve DNS before `fetch`, so a public name that resolves to a private address (for example `127.0.0.1.nip.io`) or a DNS-rebinding host is not blocked at the edge. The Go servers and the CLI do block it, because they check the resolved address at dial time.
- Allows only ports 80 and 443.
- Follows at most 5 redirects and checks every hop against the same rules.
- Times out after 10 seconds.
- Rejects pages larger than 5 MB.
- Rate limits each client (10 scans a minute; IPv6 clients are limited by their /64).

The Go `-serve` API applies the same address, port, redirect, timeout and size rules and refuses local files. The CLI itself reads any URL or file you name, because that is its job.

To report a bypass, use the advisory link above.
