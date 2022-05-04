package servicemanager

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

func (sm *ServiceManager) StartProxy() {

	routes := map[string]string{}
	rootServicePort := 9017
	proxyPort := 3000

	if sm.Commands.Port > 0 {
		proxyPort = sm.Commands.Port
	}

	// build routing table for services tagged as frontend
	for _, v := range sm.Services {
		if v.Location != "" && v.Location != "/" && v.Frontend {
			routes[v.Location] = fmt.Sprintf("localhost:%d", v.DefaultPort)
			sm.PrintVerbose("Setup: routing %s to %s on port %s\n", v.Location, v.Id, fmt.Sprint(v.DefaultPort))
		}
	}

	log.Printf("ReverseProxy: Loaded %d frontend routes\n", len(routes))
	log.Println("(only services with 'frontend: true' in services.json are addressable)")

	director := func(req *http.Request) {

		pathPrefix := "/" + strings.SplitN(req.URL.Path, "/", 3)[1]

		if proxyTo, ok := routes[pathPrefix]; ok {
			if sm.Commands.Verbose {
				log.Print(fmt.Sprintf("%s\t%s  ->  %s\n", req.Method, req.URL.Path, proxyTo))
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
