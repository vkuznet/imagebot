package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
)

// Request represents image request to the server
type Request struct {
	Namespace  string `json:"namespace"`  // namespace to use
	Tag        string `json:"tag"`        // tag of the image
	Repository string `json:"repository"` // repository of the image
	Commit     string `json:"commit"`     // commit SHA of this tag
	Service    string `json:"service"`    // service name
}

// helper function to change tag in provided string (yaml content)
func changeTag(s string, r Request) string {
	pat := fmt.Sprintf("image: %s:.*", r.Repository)
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
	fname := fmt.Sprintf("/tmp/%s-%s-%s-%s.yaml", r.Repository, r.Service, r.Namespace, r.Tag)
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
	log.Printf("deployed new image %s:%s to namespace %s, output %v\n", r.Repository, r.Tag, r.Namespace, out)
	return nil
}

// helper function to check incoming request
func checkRequest(r Request) error {
	if r.Namespace == "" || r.Tag == "" || r.Repository == "" || r.Commit == "" || r.Service == "" {
		log.Printf("ERROR, incomplete request %+v\n", r)
		return errors.New(fmt.Sprintf("incomplete request"))
	}
	if commit, err := getCommit(r); commit != r.Commit || err != nil {
		log.Printf("ERROR, unknown commit %s, request.Commit %v, error %v\n", commit, r.Commit, err)
		return errors.New(fmt.Sprintf("unknown commit %s", commit))
	}
	for idx, srv := range Config.Services {
		ns := Config.Namespaces[idx]
		repo := Config.Repositories[idx]
		if srv != r.Service {
			log.Printf("ERROR, unknown service %s, request.Service %v\n", srv, r.Service)
			return errors.New(fmt.Sprintf("unknown service %s", srv))
		}
		if ns != r.Namespace {
			log.Printf("ERROR, unknown namespace %s, request.Namespace %v\n", ns, r.Namespace)
			return errors.New(fmt.Sprintf("unknown namespace %s", ns))
		}
		if repo != r.Repository {
			log.Printf("ERROR, unknown repository %s, request.Repository %v\n", repo, r.Repository)
			return errors.New(fmt.Sprintf("unknown repository %s", repo))
		}
	}
	return nil
}
