package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	w.Write([]byte(msg + "\n"))
}

// helper function to check auth token and compare it with given image request
func auth(r *http.Request, request Request) bool {
	if arr, ok := r.Header["Authorization"]; ok {
		token := strings.Replace(arr[0], "Bearer ", "", -1)
		token = strings.Replace(token, "bearer ", "", -1)
		req, err := decodeToken(token)
		if err != nil {
			log.Printf("unable to decode token, error %v\n", err)
			return false
		}
		// check that our request is allowed to be processed
		if err := checkRequest(req); err != nil {
			log.Printf("provided request is not allowed, error %v\n", err)
			return false
		}
		if compareRequests(req, request) {
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
	if r.Method == "GET" {
		status = http.StatusBadRequest
		errorHandler(w, r, status, "", nil)
		return
		//         info := clusterInfo(Config.Namespaces)
		//         log.Printf("cluster info: %+v\n", info)
		//         data, err := json.Marshal(info)
		//         if err != nil {
		//             msg := "unable to marshal server info"
		//             status = http.StatusInternalServerError
		//             errorHandler(w, r, status, msg, err)
		//             return
		//         }
		//         w.WriteHeader(status)
		//         w.Write(data)
		//         return
	}
	defer r.Body.Close()
	var imgRequest = Request{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := "unable to read request body"
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}
	err = json.Unmarshal(data, &imgRequest)
	if err != nil {
		msg := "unable to marshal server settings"
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}
	// check if given image request match the token
	if !auth(r, imgRequest) {
		status = http.StatusUnauthorized
		errorHandler(w, r, status, "unauthorized access", nil)
		return
	}

	// check that our request is allowed to be processed
	//     if err := checkRequest(imgRequest); err != nil {
	//         msg := fmt.Sprintf("provided request is not allowed")
	//         status = http.StatusInternalServerError
	//         errorHandler(w, r, status, msg, err)
	//         return
	//     }

	// execute request
	err = exeRequest(imgRequest)
	if err != nil {
		msg := "unable to process request"
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

// TokenHandler represents token API
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	start := time.Now()
	defer logRequest(w, r, start, &status)
	if r.Method != "POST" {
		status = http.StatusInternalServerError
		w.WriteHeader(status)
		return
	}
	defer r.Body.Close()
	var imgRequest = Request{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := "unable to read request body"
		status = http.StatusInternalServerError
		errorHandler(w, r, status, msg, err)
		return
	}
	err = json.Unmarshal(data, &imgRequest)
	imgRequest.Timestamp = time.Now().Unix() + Config.TokenInterval
	token, err := genToken(imgRequest)
	if err != nil {
		status = http.StatusInternalServerError
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
	w.Write([]byte(token))
	return
}

// http server implementation
func server(serverCrt, serverKey string) {
	// the request handler
	http.HandleFunc(fmt.Sprintf("%s/status", Config.Base), StatusHandler)
	http.HandleFunc(fmt.Sprintf("%s/token", Config.Base), TokenHandler)
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
