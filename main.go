package main

import (
	"fmt"
	"flag"
	"os"
	// "github.com/openrelayxyz/cardinal-rpc"
	"github.com/openrelayxyz/cardinal-rpc/transports"
	"github.com/openrelayxyz/cardinal-proxy/config"
	"github.com/openrelayxyz/cardinal-proxy/proxy"
	"github.com/openrelayxyz/cardinal-proxy/resolver"
	log "github.com/inconshreveable/log15"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)


func main() {
	flag.CommandLine.Parse(os.Args[1:])
	cfg, err := config.LoadConfig(flag.CommandLine.Args()[0])
	if err != nil {
		log.Error("Error parsing config", "err", err)
		os.Exit(1)
	}
	if cfg.PprofPort > 0 {
		p := &http.Server{
			Addr:              fmt.Sprintf(":%v", cfg.PprofPort),
			Handler:           http.DefaultServeMux,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}
		go p.ListenAndServe()
	}
	var wg sync.WaitGroup
	// var serviceError error
	for _, service := range cfg.Services {
		tm := transports.NewTransportManager(service.Concurrency)
		proxy.RegisterProxy(tm, service.BackendURLs, service.DefaultBackendURL, resolver.MethodResolver(service.Backends))
		if service.HTTPPort > 0 {
			tm.AddHTTPServer(service.HTTPPort)
		}
		if service.WSPort > 0 {
			tm.AddWSServer(service.WSPort)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, tm *transports.TransportManager, hcport int64) {
			if err := tm.Run(hcport); err != nil {
				log.Error("Error", "err", err)
			}
			wg.Done()
		}(&wg, tm, service.HCPort)

	}
	wg.Wait()
}
