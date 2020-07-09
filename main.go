package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/go-logr/zapr"
	"github.com/supremind/pkg/config"
	"github.com/supremind/pkg/log"
	"github.com/supremind/pkg/shutdown"
)

func main() {
	address := flag.String("address", ":41357", "delivery server listen address")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	log.SetLogger(zapr.NewLogger(log.ZapLogger(*verbose)))
	log := log.WithName("main")
	conf := &Config{}
	if e := config.LoadConfig(conf); e != nil {
		log.Error(e, "load config")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux, e := Serve(ctx, conf)
	if e != nil {
		log.Error(e, "config server")
		os.Exit(1)
	}

	l, e := net.Listen("TCP", *address)
	if e != nil {
		log.Error(e, "listen", "address", *address)
		os.Exit(1)
	}

	s := http.Server{
		Addr:    *address,
		Handler: mux,
	}

	go func() {
		shutdown.BornToDie(ctx)
		s.Shutdown(ctx)
	}()

	if e := s.Serve(l); e != nil {
		log.Error(e, "server down with error")
		os.Exit(1)
	}
}
