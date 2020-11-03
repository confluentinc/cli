package utils

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"

	"github.com/davecgh/go-spew/spew"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// HTTPTracedContext returns a context.Context that verbosely traces many HTTP events that occur during the request
func HTTPTracedContext(ctx context.Context, logger *log.Logger) context.Context {
	trace := &httptrace.ClientTrace{
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			logger.Tracef("DNS Start; Info: %+v\n", dnsInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			logger.Tracef("DNS Done; Info: %+v\n", dnsInfo)
		},
		ConnectStart: func(network, addr string) {
			logger.Tracef("Connect Start; Info: network=%s, addr=%s\n", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			logger.Tracef("Connect Done; Info: network=%s, addr=%s\n", network, addr)
			if err != nil {
				logger.Tracef("Connect Done; Error: %+v\n", err)
			} else {
				logger.Tracef("(No error detected with network connection)\n")
			}
		},
		TLSHandshakeStart: func() {
			logger.Tracef("TLSHandshakeStart\n")
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			logger.Tracef("TLSHandShakeDone; Info:\n")
			spew.Dump(state)
			if err != nil {
				logger.Tracef("TLSHandShakeDone; Error: %+v\n", err)
			} else {
				logger.Tracef("(No error detected with TLS handshake)\n")
			}
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			logger.Tracef("Got Conn; Info: %+v\n", connInfo)
		},
		GetConn: func(hostPort string) {
			logger.Tracef("Get Conn; Info: %+v\n", hostPort)
		},
	}

	return httptrace.WithClientTrace(ctx, trace)
}
