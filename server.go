package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "expvar"         // to be used for monitoring, see https://github.com/divan/expvarmon
	_ "net/http/pprof" // profiler, see https://golang.org/pkg/net/http/pprof/
)

// Configuration stores server configuration parameters
type Configuration struct {
	Port             int    `json:"port"`      // server port number
	Base             string `json:"base"`      // base URL
	Verbose          int    `json:"verbose"`   // verbose output
	ServerCrt        string `json:"serverCrt"` // path to server crt file
	ServerKey        string `json:"serverKey"` // path to server key file
	LogFile          string `json:"logFile"`
	UTC              bool   `json:"utc"`
	PrintMonitRecord bool   `json:"printMonitRecord"`
}

// Config variable represents configuration object
var Config Configuration

// helper function to parse configuration
func parseConfig(configFile string) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read", err)
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Println("Unable to parse", err)
		return err
	}
	return nil
}

func errorHandler(w http.ResponseWriter, r *http.Request, msg string, err error) {
	log.Println(msg, err)
	w.Write([]byte(msg))
}

type Request struct {
	Namespace string
	Name      string
	Tag       string
	Repo      string
	Token     string
}

func changeTag(s string, r Request) string {
	pat := fmt.Sprintf("image: %s/%s:.*", r.Repo, r.Name)
	re := regexp.MustCompile(pat)
	img := fmt.Sprintf("image: %s/%s:%s", r.Repo, r.Name, r.Tag)
	return re.ReplaceAllString(s, img)
}

func exeRequest(r Request) error {
	log.Printf("execute request %+v\n", r)
	var args []string
	// get yaml of our request image
	args = []string{"get", "deployment", r.Name, "-n", r.Namespace, "-o", "yaml"}
	out, err := exe("kubectl", args...)
	if err != nil {
		return err
	}
	// change image tag
	content := changeTag(strings.Join(out, "\n"), r)

	// write new yml file
	fname := fmt.Sprintf("/tmp/%s-%s-%s-%s.yaml", r.Repo, r.Name, r.Namespace, r.Tag)
	err = ioutil.WriteFile(fname, []byte(content), 0777)
	if err != nil {
		return err
	}

	// kubectl apply -f file.yml
	args = []string{"apply", "-f", fname}
	out, err = exe("kubectl", args...)
	if err != nil {
		return err
	}
	log.Printf("deployed new image %s/%s:%s to namespace %s, output %v\n", r.Repo, r.Name, r.Tag, r.Namespace, out)
	return nil
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello"))
		return
	}
	defer r.Body.Close()
	var imgRequest = Request{}
	err := json.NewDecoder(r.Body).Decode(&imgRequest)
	if err != nil {
		msg := "unable to marshal server settings"
		errorHandler(w, r, msg, err)
		return
	}
	err = exeRequest(imgRequest)
	if err != nil {
		msg := "unable to execute request"
		errorHandler(w, r, msg, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// http server implementation
func server(serverCrt, serverKey string) {
	// the request handler
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
