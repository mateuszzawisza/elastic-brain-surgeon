package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

var esAddresses addresses

func init() {
	flag.Var(&esAddresses, "elasticsearch-list", "comma sperated list of elasticsearch instances addresses")
}

func main() {
	flag.Parse()
	for _, node := range esAddresses {
		cs := getClusterState(node)
		printClusterState(cs)
	}
}

func getClusterState(address string) ClusterState {
	statusEndpoint := fmt.Sprintf("http://%s/_cluster/state/nodes,master_node", address)
	resp, err := http.Get(statusEndpoint)
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

// address flag
type addresses []string

func (i *addresses) String() string {
	return fmt.Sprint(*i)
}

func (i *addresses) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("Addresses flag already set")
	}
	for _, address := range strings.Split(value, ",") {
		*i = append(*i, address)
	}
	return nil
}
