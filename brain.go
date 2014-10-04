package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

type NodeStatus struct {
	Status int    `json:"status"`
	Name   string `json:"name"`
}

type ElasticsearchNode struct {
	Name           string
	Status         int
	MasterNode     string
	NodesInCluster int
}

var esAddresses addresses
var strict bool

func init() {
	flag.Var(&esAddresses, "elasticsearch-list", "comma sperated list of elasticsearch instances addresses")
	flag.BoolVar(&strict, "strict", false, "Strict exit status")
}

func main() {
	flag.Parse()
	nodes := fetchNodes(esAddresses)
	split := checkForSplitBrain(nodes)
	if split {
		fmt.Println("The brain is split!")
		for _, node := range nodes {
			printNodeStatus(node)
		}
		if strict {
			os.Exit(1)
		}
	} else {
		fmt.Println("Everything is ok")
	}

}

func checkForSplitBrain(nodes []ElasticsearchNode) bool {
	for i := 1; i < len(nodes); i++ {
		if nodes[i].MasterNode != nodes[i-1].MasterNode {
			return false
		}
	}
	return true
}

func fetchNodes(esAddresses []string) []ElasticsearchNode {
	nodes := make([]ElasticsearchNode, len(esAddresses))
	for i, node := range esAddresses {
		ns := getNodeStatus(node)
		cs := getClusterState(node)
		node := ElasticsearchNode{
			ns.Name,
			ns.Status,
			cs.Nodes[cs.MasterNode].Name,
			len(cs.Nodes),
		}
		nodes[i] = node
	}
	return nodes
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

func getNodeStatus(address string) NodeStatus {
	statusEndpoint := fmt.Sprintf("http://%s", address)
	resp, err := http.Get(statusEndpoint)
	if err != nil {
		log.Panic("could not connect to node")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var ns NodeStatus
	json.Unmarshal(body, &ns)
	return ns
}

func printClusterState(cs ClusterState) {
	fmt.Printf("master node: %s - %s (%s)\n", cs.MasterNode, cs.Nodes[cs.MasterNode].Name, cs.Nodes[cs.MasterNode].TransportAddress)
	//	fmt.Printf("cluster name: %s\n", cs.ClusterName)
	//	fmt.Printf("Nodes in the cluster: \n")
	//	for key, value := range cs.Nodes {
	//		fmt.Printf("%s => %s (%s) \n", key, value.Name, value.TransportAddress)
	//	}
}

func printNodeStatus(ns ElasticsearchNode) {
	fmt.Printf("node: %s (status: %d) => master: %s nodes: %d\n", ns.Name, ns.Status, ns.MasterNode, ns.NodesInCluster)
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
