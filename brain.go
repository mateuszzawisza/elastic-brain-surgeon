package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	connectToNode()
}

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

func connectToNode() ClusterState {
	resp, err := http.Get("http://localhost:9200/_cluster/state/nodes,master_node")
	if err != nil {
		log.Panic("could not connect to node")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var cs ClusterState
	json.Unmarshal(body, &cs)
	fmt.Printf("master node: %s\n", cs.MasterNode)
	fmt.Printf("cluster name: %s\n", cs.ClusterName)
	for key, value := range cs.Nodes {
		fmt.Printf("Node: %s => %s (%s) \n", key, value.Name, value.TransportAddress)
	}
	return cs
}
