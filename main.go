package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/go-logr/zapr"
	"github.com/supremind/pkg/log"
	"github.com/supremind/pkg/shutdown"
	"gopkg.in/yaml.v2"
)

func main() {
	address := flag.String("address", ":41357", "delivery server listen address")
	config := flag.String("config", "", "path to config file")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	log.SetLogger(zapr.NewLogger(log.ZapLogger(*verbose)))
	log := log.WithName("main")

	confFile, e := os.Open(*config)
	if e != nil {
		log.Error(e, "open config file")
		os.Exit(1)
	}
	defer confFile.Close()
	conf := &Config{}
	if e := yaml.NewDecoder(confFile).Decode(conf); e != nil {
		log.Error(e, "decode config file")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux, e := Serve(ctx, conf)
	if e != nil {
		log.Error(e, "config server")
		os.Exit(1)
	}

	l, e := net.Listen("tcp", *address)
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
