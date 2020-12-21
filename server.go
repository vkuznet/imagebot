package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "expvar"         // to be used for monitoring, see https://github.com/divan/expvarmon
	_ "net/http/pprof" // profiler, see https://golang.org/pkg/net/http/pprof/
)

// helper function to hangle errors
func errorHandler(w http.ResponseWriter, r *http.Request, status int, msg string, err error) {
	log.Println(msg, err)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

// helper function to check auth token
func auth(r *http.Request) bool {
	if arr, ok := r.Header["Authorization"]; ok {
		token := strings.Replace(arr[0], "Bearer ", "", -1)
		token = strings.Replace(token, "bearer ", "", -1)
		if token == Config.Token {
			return true
		}
	}
	return false
}

// RequestHandler represents incoming request handler
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	start := time.Now()
	defer logRequest(w, r, start, &status)
	if !auth(r) {
		status = http.StatusUnauthorized
		errorHandler(w, r, status, "unauthorized access", nil)
		return
	}
	if r.Method == "GET" {
		info := clusterInfo(Config.Namespaces)
		log.Printf("cluster info: %+v\n", info)
		data, err := json.Marshal(info)
		if err != nil {
			msg := "unable to marshal server info"
			status = http.StatusInternalServerError
			errorHandler(w, r, status, msg, err)
			return
		}
		w.WriteHeader(status)
		w.Write(data)
		return
	}
	defer r.Body.Close()
	var imgRequest = Request{}
	err := json.NewDecoder(r.Body).Decode(&imgRequest)
	if err != nil {
		msg := "unable to marshal server settings"
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}

	// check that our request is allowed to be processed
	if err := checkRequest(imgRequest); err != nil {
		msg := fmt.Sprintf("provided request is not allowed")
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}

	// execute request
	err = exeRequest(imgRequest)
	if err != nil {
		msg := "unable to execute request"
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}
	w.WriteHeader(status)
}

// StatusHandler represents incoming request handler
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	start := time.Now()
	defer logRequest(w, r, start, &status)
	if r.Method != "GET" {
		status = http.StatusInternalServerError
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
	return
}

// http server implementation
func server(serverCrt, serverKey string) {
	// the request handler
	http.HandleFunc(fmt.Sprintf("%s/status", Config.Base), StatusHandler)
	http.HandleFunc(fmt.Sprintf("%s/", Config.Base), RequestHandler)

	// start HTTP or HTTPs server based on provided configuration
	addr := fmt.Sprintf(":%d", Config.Port)
	if serverCrt != "" && serverKey != "" {
		//start HTTPS server which require user certificates
		server := &http.Server{Addr: addr}
		log.Printf("Starting HTTPs server on %s", addr)
		log.Fatal(server.ListenAndServeTLS(serverCrt, serverKey))
	} else {
		// Start server without user certificates
		log.Printf("Starting HTTP server on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
