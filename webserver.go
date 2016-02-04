package main

import (
	"flag"
    "fmt"
    "net/http"
)   

// We want to serve the current directory from the root path,
// which is not allowed by the default ServeMux :-(.
// Otherwise, this code would be a lot simpler
var chttp = http.NewServeMux()

func main() {
	var port = flag.String("port", "8080", "Define what TCP port to bind to")
	var root = flag.String("root", ".", "Define the root filesystem path")
	flag.Parse()
	
    chttp.Handle("/", http.FileServer(http.Dir(*root)))

    http.HandleFunc("/", FileHandler)

	fmt.Printf("Serving directory '%s' on port :%s\n", *root, *port)
    panic(http.ListenAndServe(":" + *port, nil))
}   

func FileHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	
	chttp.ServeHTTP(w, r)

	// Example of how to filter requests, if that's your thing
    // if (strings.Contains(r.URL.Path, ".")) {
    //     chttp.ServeHTTP(w, r)
    // } else {
    //     fmt.Fprintf(w, "HomeHandler")
    // }   
} 
