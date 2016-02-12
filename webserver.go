package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/travissimon/remnant/client"
)

// We want to serve the current directory from the root path,
// which is not allowed by the default ServeMux :-(.
// Otherwise, this code would be a lot simpler
var chttp = http.NewServeMux()

func FileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Webserver: %s\n", time.Now().UTC().Format(time.RFC3339), r.URL.Path)
	chttp.ServeHTTP(w, r)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Proxy: %s\n", time.Now().UTC().Format(time.RFC3339), r.URL.Path)
	cl, err := client.NewRemnantClient("localhost:8123", r)
	defer cl.EndSpan()
	if err != nil {
		fmt.Printf("Error creating remnant client: %s", err.Error())
		return
	}

	serviceAndPath := r.URL.Path[len("/proxy/"):]
	slashIdx := strings.IndexRune(serviceAndPath, '/')
	service := serviceAndPath[:slashIdx]
	path := serviceAndPath[slashIdx+1:]

	if service == "" {
		cl.LogError(fmt.Sprintf("Proxy request with no service path: %s\n", r.URL.Path))
		io.WriteString(w, "Could not proxy request :-(")
		return
	}

	msg := fmt.Sprintf("Proxy to %s/%s\n", service, path)
	io.WriteString(w, msg)
	fmt.Printf(msg)
	cl.LogDebug(msg)
}

func main() {
	var port = flag.String("port", "8080", "Define what TCP port to bind to")
	var root = flag.String("root", ".", "Define the root filesystem path")
	flag.Parse()

	http.HandleFunc("/proxy/", ProxyHandler)

	chttp.Handle("/", http.FileServer(http.Dir(*root)))
	http.HandleFunc("/", FileHandler)

	fmt.Printf("Proxying webserver: serving directory '%s' on port :%s\n\n", *root, *port)
	panic(http.ListenAndServe(":"+*port, nil))
}
