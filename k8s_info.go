package main

// k8s_info - tool to provide k8s information about pods/images, etc.
//
// Copyright (c) 2020 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
)

type ContainerStatus struct {
	ContainerID  string                 `json:"ContainerID"`
	Image        string                 `json:"Image"`
	ImageID      string                 `json:"ImageID"`
	LastState    interface{}            `json:"LastState"`
	Name         string                 `json:"Name"`
	Ready        bool                   `json:"Ready"`
	RestartCount int                    `json:"RestartCount"`
	State        map[string]interface{} `json:"State"`
}

type Metadata struct {
	//     Annotations       map[string]string        `json:"Annotations"`
	//     CreationTimestamp string                   `json:"CreationTimestamp"`
	//     GenerateName      string                   `json:"GenerateName"`
	//     Labels            map[string]string        `json:"Labels"`
	Name      string `json:"Name"`
	Namespace string `json:"Namespace"`
	//     OwnerReferences   []map[string]interface{} `json:"OwnerReferences"`
}

type Spec struct {
	Containers []map[string]interface{} `json:"Containers"`
}

type Status struct {
	Conditions            []interface{}       `json:"Conditions"`
	ContainerStatuses     []ContainerStatus   `json:"ContainerStatuses"`
	HostIP                string              `json:"HostIP"`
	InitContainerStatuses []ContainerStatus   `json:"InitContainerStatuses"`
	Phase                 string              `json:"Phase"`
	PodIP                 string              `json:"PodIP"`
	PodIPs                []map[string]string `json:"PodIPs"`
	QosClass              string              `json:"QosClass"`
	StartTime             string              `json:"StartTime"`
}

type PodInfo struct {
	//     ApiVersion string   `json:"ApiVersion"`
	//     Kind       string   `json:"Kind"`
	Metadata Metadata `json:"Metadata"`
	//     Spec       Spec     `json:"Spec"`
	//     Status Status `json:Status`
}

// helper function to execute command
func exe(command string, args ...string) ([]string, error) {
	var out []string
	log.Println("command", command, args)
	cmd := exec.Command(command, args...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal("ERROR:", err, "while executing", command, args)
	}
	for _, v := range strings.Split(string(stdout), "\n") {
		if strings.HasPrefix(v, "NAME") {
			continue
		}
		arr := strings.Split(v, " ")
		if len(arr) > 0 {
			v := strings.Trim(arr[0], " ")
			if v != "" {
				out = append(out, arr[0])
			}
		}
	}
	return out, nil
}

// helper function to get namespaces
func namespaces() ([]string, error) {
	args := []string{"get", "namespaces", "-A"}
	out, err := exe("kubectl", args...)
	return out, err
}

// helper function to get deployments
func deployments(ns string) ([]string, error) {
	args := []string{"get", "deployments", "-n", ns}
	out, err := exe("kubectl", args...)
	return out, err
}

// helper function to get pods
func pods(ns string) ([]string, error) {
	args := []string{"get", "pods", "-n", ns}
	out, err := exe("kubectl", args...)
	return out, err
}

// helper function to get pod information
func podInfo(pod, ns string) (PodInfo, error) {
	var rec PodInfo
	args := []string{"get", "pod", "-n", ns, pod, "-o", "json"}
	cmd := exec.Command("kubectl", args...)
	stdout, err := cmd.Output()
	if err != nil {
		return rec, err
	}
	//     fmt.Println("output of pod info", string(stdout))
	err = json.Unmarshal(stdout, &rec)
	return rec, err
}

// InList helper function to check item in a list
func InList(a string, list []string) bool {
	check := 0
	for _, b := range list {
		if b == a {
			check += 1
		}
	}
	if check != 0 {
		return true
	}
	return false
}

// helper function to return cluster info
func clusterInfo(allowed []string) []PodInfo {
	log.Println("allowed namespaces", allowed)
	var info []PodInfo
	nss, err := namespaces()
	if err != nil {
		log.Println("ERROR", err)
		return info
	}
	log.Println("namespaces", nss)
	for _, ns := range nss {
		if !InList(ns, allowed) {
			log.Println("skip", ns)
			continue
		}
		pods, err := pods(ns)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}
		log.Println("pods", pods)
		for _, pod := range pods {
			p, err := podInfo(pod, ns)
			if err != nil {
				log.Println("ERROR", err)
				continue
			}
			info = append(info, p)
		}
	}
	return info
}
