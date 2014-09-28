package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type ClusterState struct {
	MasterNode  string          `json:"master_node"`
	ClusterName string          `json:"cluster_name"`
	Nodes       map[string]Node `json:"nodes"`
}
type Node struct {
	Name             string `json:"name"`
	TransportAddress string `json:"transport_address"`
	Attributes       struct {
		AwsZone string `json:"aws_zone"`
	} `json:"attributes"`
}

func main() {
	cs := getClusterState()
	printClusterState(cs)
}

func getClusterState() ClusterState {
	resp, err := http.Get("http://localhost:9200/_cluster/state/nodes,master_node")
	if err != nil {
		log.Panic("could not connect to node")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var cs ClusterState
	json.Unmarshal(body, &cs)
	return cs
}

func printClusterState(cs ClusterState) {
	fmt.Printf("master node: %s - %s (%s)\n", cs.MasterNode, cs.Nodes[cs.MasterNode].Name, cs.Nodes[cs.MasterNode].TransportAddress)
	fmt.Printf("cluster name: %s\n", cs.ClusterName)
	fmt.Printf("Nodes in the cluster: \n")
	for key, value := range cs.Nodes {
		fmt.Printf("%s => %s (%s) \n", key, value.Name, value.TransportAddress)
	}
}
