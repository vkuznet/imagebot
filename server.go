package main

import (
	"crypto/tls"
	"crypto/x509"
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
		w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var imgRequest = Request{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(data, &imgRequest); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	imgRequest.Timestamp = time.Now().Unix() + Config.TokenInterval
	token, err := genToken(imgRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	if serverCrt == "" && serverKey == "" {
		// Start server without user certificates
		log.Printf("Starting HTTP server on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	} else {
		// start HTTP or HTTPs server based on provided configuration
		rootCAs := x509.NewCertPool()
		files, err := ioutil.ReadDir(Config.RootCAs)
		if err != nil {
			log.Fatalf("Unable to list files in '%s', error: %v\n", Config.RootCAs, err)
		}
		for _, finfo := range files {
			fname := fmt.Sprintf("%s/%s", Config.RootCAs, finfo.Name())
			caCert, err := ioutil.ReadFile(fname)
			if err != nil {
				if Config.Verbose > 1 {
					log.Printf("Unable to read %s\n", fname)
				}
			}
			if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
				if Config.Verbose > 1 {
					log.Printf("invalid PEM format while importing trust-chain: %q", fname)
				}
			}
			log.Println("Load CA file", fname)
		}
		cert, err := tls.LoadX509KeyPair(serverCrt, serverKey)
		if err != nil {
			log.Fatalf("server loadkeys: %s", err)

		}
		tlsConfig := &tls.Config{
			RootCAs:      rootCAs,
			Certificates: []tls.Certificate{cert},
		}
		addr := fmt.Sprintf(":%d", Config.Port)
		server := &http.Server{
			Addr:           addr,
			TLSConfig:      tlsConfig,
			ReadTimeout:    time.Duration(Config.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(Config.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		//start HTTPS server which require user certificates
		log.Printf("Starting HTTPs server on %s", addr)
		log.Fatal(server.ListenAndServeTLS(serverCrt, serverKey))

	}
}
