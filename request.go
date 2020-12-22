package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"time"
)

// Request represents image request to the server
type Request struct {
	Namespace  string `json:"namespace"`  // namespace to use
	Tag        string `json:"tag"`        // tag of the image
	Repository string `json:"repository"` // repository of the image
	Commit     string `json:"commit"`     // commit SHA of this tag
	Service    string `json:"service"`    // service name
	Expire     int64  `json:"expire"`     // expire timestamp of request
}

// helper function to change tag in provided string (yaml content)
func changeTag(s string, r Request) string {
	pat := fmt.Sprintf("image: %s.*", r.Repository)
	re := regexp.MustCompile(pat)
	img := fmt.Sprintf("image: %s:%s", r.Repository, r.Tag)
	return re.ReplaceAllString(s, img)
}

// helper function to execute request on k8s
func exeRequest(r Request) error {
	log.Printf("execute request %+v\n", r)
	var args []string

	// get yaml of our request image
	args = []string{"get", "deployment", r.Service, "-n", r.Namespace, "-o", "yaml"}
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
	fname := fmt.Sprintf("/tmp/%s-%s-%s.yaml", r.Service, r.Namespace, r.Tag)
	log.Println("fname", fname)
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
	log.Printf("deployed new image %s:%s to namespace %s, output %v\n", r.Repository, r.Tag, r.Namespace, string(out))
	return nil
}

// helper function to check incoming request
func checkRequest(r Request) error {
	if r.Namespace == "" || r.Tag == "" || r.Repository == "" || r.Commit == "" || r.Service == "" {
		log.Printf("ERROR, incomplete request %+v\n", r)
		return fmt.Errorf("incomplete request")
	}
	if r.Expire < time.Now().Unix() {
		log.Printf("ERROR, request expired %+v\n", r)
		return fmt.Errorf("expired request")
	}
	if commit, err := getCommit(r); commit != r.Commit || err != nil {
		log.Printf("ERROR, unknown commit %s, request.Commit %v, error %v\n", commit, r.Commit, err)
		return fmt.Errorf("unknown commit %s", commit)
	}
	for idx, srv := range Config.Services {
		ns := Config.Namespaces[idx]
		repo := Config.Repositories[idx]
		if srv != r.Service {
			log.Printf("ERROR, unknown service %s, request.Service %v\n", srv, r.Service)
			return fmt.Errorf("unknown service %s", srv)
		}
		if ns != r.Namespace {
			log.Printf("ERROR, unknown namespace %s, request.Namespace %v\n", ns, r.Namespace)
			return fmt.Errorf("unknown namespace %s", ns)
		}
		if repo != r.Repository {
			log.Printf("ERROR, unknown repository %s, request.Repository %v\n", repo, r.Repository)
			return fmt.Errorf("unknown repository %s", repo)
		}
	}
	return nil
}

// helper function to compare requests
func compareRequests(r1, r2 Request) bool {
	if r1.Namespace == r2.Namespace || r1.Service == r2.Service || r1.Tag == r2.Tag || r1.Repository == r2.Repository || r1.Commit == r2.Commit {
		return true
	}
	log.Printf("requests do not match: %+v != %+v\n", r1, r2)
	return false
}

// helper function to generate token
func genToken(r Request) (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	data, err = encrypt(data, Config.Secret)
	hash := base64.StdEncoding.EncodeToString(data)
	return hash, err
}

// helper function to decode token
func decodeToken(t string) (Request, error) {
	var r Request
	data, err := base64.StdEncoding.DecodeString(t)
	if err != nil {
		return r, err
	}
	data, err = decrypt(data, Config.Secret)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(data, &r)
	return r, err
}
