// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/carlmjohnson/versioninfo"
	"tailscale.com/tsnet"
)

var (
	addr     = flag.String("addr", ":443", "address to listen on")
	hostname = flag.String("hostname", "tsrproxy", "hostname for the reverse proxy")
	rpURL    *url.URL
)

func init() {
	versioninfo.AddFlag(nil)
	flag.Func("proxy", "`url` to reverse proxy", func(s string) error {
		var err error
		rpURL, err = url.Parse(s)
		return err
	})
}

func main() {
	flag.Parse()
	s := &tsnet.Server{
		Hostname: *hostname,
		Logf:     func(format string, args ...any) {},
	}
	ln, err := s.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	lc, err := s.LocalClient()
	if err != nil {
		log.Fatal(err)
	}

	if *addr == ":443" {
		ln = tls.NewListener(ln, &tls.Config{
			GetCertificate: lc.GetCertificate,
		})
	}
	log.Printf("started %s%s proxing to %v",
		*hostname, *addr, rpURL)

	log.Fatal(http.Serve(ln, httputil.NewSingleHostReverseProxy(rpURL)))
}
