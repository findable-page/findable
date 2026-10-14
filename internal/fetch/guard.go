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
	blockedPrefixes   = []netip.Prefix{
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("2001::/32"),
	}
)

func PublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() {
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
