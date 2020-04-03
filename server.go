package main

import (
	"fmt"
	"net/http"
)

const (
	serverPort = "8000"
)

func main() {
	http.HandleFunc("/", myHandler)
	http.ListenAndServe(":"+serverPort, nil)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.URL.String()[1:]
	fmt.Fprintf(w, title)
}

// https://www.cnblogs.com/wolfred7464/p/4670864.html
