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
	"os"
	"path/filepath"

	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/versioninfo"
	"tailscale.com/tsnet"
)

var (
	addr     = flag.String("addr", ":443", "address to listen on")
	hostname = flag.String("hostname", "tsrproxy", "hostname for the reverse proxy")
	authkey  = flag.String("authkey", os.Getenv("TS_AUTHKEY"), "`key` for proxy server")
	verbose  = flag.Bool("verbose", false, "log Tailscale output")
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
	flagx.ParseEnv(nil, "tsrproxy")
	if rpURL == nil {
		flag.Usage()
		log.Fatal("-proxy required")
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir = filepath.Join(dir, "tsnet-tsrproxy-"+*hostname)
	_ = os.MkdirAll(dir, 0o700)

	logf := func(format string, args ...any) {}
	if *verbose {
		logf = nil
	}

	s := &tsnet.Server{
		Dir:      dir,
		Hostname: *hostname,
		AuthKey:  *authkey,
		Logf:     logf,
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
	log.Printf("starting %s%s proxing to %v",
		*hostname, *addr, rpURL)

	log.Fatal(http.Serve(ln, httputil.NewSingleHostReverseProxy(rpURL)))
}
