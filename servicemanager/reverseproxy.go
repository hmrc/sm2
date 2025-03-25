package servicemanager

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sm2/ledger"
	"strings"
	"time"
)

func (sm *ServiceManager) StartProxy() {

	rootServicePort := 9017
	proxyPort := 3000

	if sm.Commands.Port > 0 {
		proxyPort = sm.Commands.Port
	}

	var routes map[string]string

	requestedServices := sm.requestedServicesAndProfiles()
	if len(requestedServices) > 0 {
		definedServices := map[string]Service{}
		for _, v := range requestedServices {
			if s, ok := sm.Services[v.service]; ok {
				definedServices[v.service] = s
			}
		}
		routes = buildRoutingTable(definedServices)
	} else {
		routes = buildRoutingTable(sm.Services)
	}

	log.Printf("ReverseProxy: Loaded %d frontend routes\n", len(routes))
	log.Println("(only services with 'frontend: true' in services.json are addressable)")

	state := ledger.ProxyState{Started: time.Now(), Pid: os.Getpid(), ProxyPaths: routes}
	sm.Ledger.SaveProxyState(sm.Config.TmpDir, state)

	director := func(req *http.Request) {

		pathPrefix := "/" + strings.SplitN(req.URL.Path, "/", 3)[1]

		if proxyTo, ok := routes[pathPrefix]; ok {
			if sm.Commands.Verbose {
				log.Printf("%s\t%s  ->  %s\n", req.Method, req.URL.Path, proxyTo)
			}
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", proxyTo)
			req.URL.Scheme = "http"
			req.URL.Host = proxyTo
		} else {
			// handle anything that doesn't match
			// this would be anything that hangs off '/' like catalogue frontend etc
			sm.PrintVerbose("%s %s\t-> No Proxy!\n", req.Method, req.URL.Path)
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("localhost:%d", rootServicePort)
		}
	}

	proxy := &httputil.ReverseProxy{Director: director}

	mux := http.NewServeMux()
	mux.Handle("/", proxy)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", proxyPort),
		Handler: mux,
	}

	log.Printf("ReverseProxy: listening on port %d...", proxyPort)
	log.Fatal(server.ListenAndServe())
}

func buildRoutingTable(services map[string]Service) map[string]string {
	routes := map[string]string{}
	for _, v := range services {
		for _, path := range v.ProxyPaths {
			routes[path] = fmt.Sprintf("localhost:%d", v.DefaultPort)
			log.Printf("Setup: routing %s to %s on port %s\n", path, v.Id, fmt.Sprint(v.DefaultPort))
		}
	}
	return routes
}
