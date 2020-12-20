package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	_ "expvar"         // to be used for monitoring, see https://github.com/divan/expvarmon
	_ "net/http/pprof" // profiler, see https://golang.org/pkg/net/http/pprof/
)

// Configuration stores server configuration parameters
type Configuration struct {
	Port       int      `json:"port"`       // server port number
	Base       string   `json:"base"`       // base URL
	Verbose    int      `json:"verbose"`    // verbose output
	ServerCrt  string   `json:"serverCrt"`  // path to server crt file
	ServerKey  string   `json:"serverKey"`  // path to server key file
	Namespaces []string `json:"namespaces"` // list of allowed namespaces
	UTC        bool     `json:"utc"`        // use UTC for logging or not
	Services   []string `json:services`     // list of allowed services
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

// helper function to hangle errors
func errorHandler(w http.ResponseWriter, r *http.Request, msg string, err error) {
	log.Println(msg, err)
	w.Write([]byte(msg))
}

// Request represents image request to the server
type Request struct {
	Namespace string `json:"namespace"` // namespace to use
	Name      string `json:"name"`      // name of the image
	Tag       string `json:"tag"`       // tag of the image
	Repo      string `json:"repo"`      // repository of the image
	Token     string `json:"token"`     // authentication token
}

// helper function to change tag in provided string (yaml content)
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
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	yaml := string(out)
	log.Println("YAML", yaml)
	// change image tag
	content := changeTag(yaml, r)
	log.Println("NEW YAML", content)

	// write new yml file
	fname := fmt.Sprintf("/tmp/%s-%s-%s-%s.yaml", r.Repo, r.Name, r.Namespace, r.Tag)
	err = ioutil.WriteFile(fname, []byte(content), 0777)
	if err != nil {
		return err
	}

	// kubectl apply -f file.yml
	args = []string{"apply", "-f", fname}
	cmd = exec.Command("kubectl", args...)
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	log.Printf("deployed new image %s/%s:%s to namespace %s, output %v\n", r.Repo, r.Name, r.Tag, r.Namespace, out)
	return nil
}

// RequestHandler represents incoming request handler
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	start := time.Now()
	defer logRequest(w, r, start, &status)
	if r.Method == "GET" {
		info := clusterInfo(Config.Namespaces)
		log.Printf("cluster info: %+v\n", info)
		data, err := json.Marshal(info)
		if err != nil {
			msg := "unable to marshal server info"
			errorHandler(w, r, msg, err)
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
		errorHandler(w, r, msg, err)
		return
	}
	// check if given request is allowed
	if !InList(imgRequest.Name, Config.Services) {
		msg := fmt.Sprintf("provided service %s is not allowed", imgRequest.Name)
		errorHandler(w, r, msg, nil)
		return
	}
	if !InList(imgRequest.Namespace, Config.Namespaces) {
		msg := fmt.Sprintf("provided namespace %s is not allowed", imgRequest.Namespace)
		errorHandler(w, r, msg, nil)
		return
	}

	// execute request
	err = exeRequest(imgRequest)
	if err != nil {
		msg := "unable to execute request"
		errorHandler(w, r, msg, err)
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
