package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/nytimes/gziphandler"
	"github.com/travissimon/remnant/client"
)

// We want to serve the current directory from the root path,
// which is not allowed by the default ServeMux :-(.
// Otherwise, this code would be a lot simpler
var chttp = http.NewServeMux()

func FileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Webserver: %s\n", time.Now().UTC().Format(time.RFC3339Nano), r.URL.Path)
	chttp.ServeHTTP(w, r)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	proxyPath := p.ByName("proxypath")[1:]

	cl, err := client.NewRemnantClient(remnantUrl, r)
	defer cl.EndSpan()
	if err != nil {
		fmt.Printf("Error creating remnant client: %s", err.Error())
	}
	// mark this as the originating span
	cl.Span.TraceId = cl.Span.Id
	cl.Span.ParentId = cl.Span.Id
	w.Header().Set("remnant-trace-id", cl.Span.TraceId)

	slashIdx := strings.IndexRune(proxyPath, '/')
	service := proxyPath[:slashIdx]
	path := proxyPath[slashIdx+1:]

	if service == "" {
		cl.LogError(fmt.Sprintf("Proxy request with no service path: %s\n", r.URL.Path))
		io.WriteString(w, "Could not proxy request :-(")
		return
	}

	// Ghetto DNS
	var url = "http://localhost"
	if service == "strlen" {
		url += ":8001"
	} else if service == "uniq" {
		url += ":8002"
	}

	destUrl := url + "/" + path
	fmt.Printf("Calling to %s\n", destUrl)
	resp, err := cl.Get(destUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling %s: %s\n", destUrl, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling %s: %s\n", destUrl, err.Error())
		return
	}

	w.Write(body)
}

var remnantUrl string

type staticFileServer struct {
	fileServer http.Handler
}

func (sfs *staticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Cache-Control", "public, no-cache")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}
	sfs.fileServer.ServeHTTP(w, r)
}

func main() {
	var port = flag.String("port", "8080", "Define what TCP port to bind to")
	var root = flag.String("root", ".", "Define the root filesystem path")
	var remnant = flag.String("remnant", "http://localhost:7777/", "")
	flag.Parse()

	remnantUrl = *remnant

	// create our own router so that we can add our own headers to static file responses
	router := httprouter.New()
	router.HandleMethodNotAllowed = false

	router.GET("/proxy/*proxypath", ProxyHandler)

	router.NotFound = &staticFileServer{
		gziphandler.GzipHandler(http.FileServer(http.Dir(*root))),
	}

	fmt.Printf("Proxying webserver: serving directory '%s' on port :%s\n", *root, *port)
	panic(http.ListenAndServe(":"+*port, router))
}
