package utils

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"

	"github.com/davecgh/go-spew/spew"

	"github.com/confluentinc/cli/v3/pkg/log"
)

func GetContext() context.Context {
	ctx := context.Background()
	if log.CliLogger.Level == log.TRACE {
		ctx = httpTracedContext(ctx)
	}
	return ctx
}

// httpTracedContext returns a context.Context that verbosely traces many HTTP events that occur during the request
func httpTracedContext(ctx context.Context) context.Context {
	trace := &httptrace.ClientTrace{
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			log.CliLogger.Tracef("DNS Start; Info: %+v\n", dnsInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			log.CliLogger.Tracef("DNS Done; Info: %+v\n", dnsInfo)
		},
		ConnectStart: func(network, addr string) {
			log.CliLogger.Tracef("Connect Start; Info: network=%s, addr=%s\n", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			log.CliLogger.Tracef("Connect Done; Info: network=%s, addr=%s\n", network, addr)
			if err != nil {
				log.CliLogger.Tracef("Connect Done; Error: %+v\n", err)
			} else {
				log.CliLogger.Tracef("(No error detected with network connection)\n")
			}
		},
		TLSHandshakeStart: func() {
			log.CliLogger.Tracef("TLSHandshakeStart\n")
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			log.CliLogger.Tracef("TLSHandShakeDone; Info:\n")
			spew.Dump(state)
			if err != nil {
				log.CliLogger.Tracef("TLSHandShakeDone; Error: %+v\n", err)
			} else {
				log.CliLogger.Tracef("(No error detected with TLS handshake)\n")
			}
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			log.CliLogger.Tracef("Got Conn; Info: %+v\n", connInfo)
		},
		GetConn: func(hostPort string) {
			log.CliLogger.Tracef("Get Conn; Info: %+v\n", hostPort)
		},
	}

	return httptrace.WithClientTrace(ctx, trace)
}
