package fetch

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

const (
	MaxPageBytes   = 5 << 20
	guardedTimeout = 10 * time.Second
	maxRedirects   = 5
)

var (
	ErrBlockedAddress = errors.New("blocked address: only public internet hosts can be scanned")
	ErrPageTooLarge   = fmt.Errorf("page is larger than %d bytes", MaxPageBytes)
	errTooManyHops    = errors.New("too many redirects")
	globalV6          = netip.MustParsePrefix("2000::/3")
	blockedPrefixes   = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("64:ff9b::/96"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("2001::/32"),
		netip.MustParsePrefix("3fff::/20"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("2001:db8::/32"),
	}
)

func PublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || (addr.Is6() && !globalV6.Contains(addr)) {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return false
		}
	}
	return true
}

func refuseNonPublic(_, address string, _ syscall.RawConn) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil || !PublicAddr(ap.Addr()) || (ap.Port() != 80 && ap.Port() != 443) {
		return fmt.Errorf("%s: %w", address, ErrBlockedAddress)
	}
	return nil
}

func NewGuardedClient() *http.Client {
	dialer := &net.Dialer{Timeout: guardedTimeout, Control: refuseNonPublic}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   guardedTimeout,
		ResponseHeaderTimeout: guardedTimeout,
	}
	return &http.Client{
		Timeout:   guardedTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errTooManyHops
			}
			if !IsHTTPURL(req.URL.String()) {
				return ErrBlockedAddress
			}
			return nil
		},
	}
}
