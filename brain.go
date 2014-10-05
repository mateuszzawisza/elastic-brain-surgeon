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
	ErrorFetching  bool
}

var esAddresses addresses
var strict bool
var printStatus bool

func init() {
	flag.Var(&esAddresses, "elasticsearch-list", "comma sperated list of elasticsearch instances addresses")
	flag.BoolVar(&strict, "strict", false, "Strict exit status")
	flag.BoolVar(&printStatus, "print", false, "Print cluster status")
}

func main() {
	flag.Parse()
	nodes := fetchNodes(esAddresses)
	split := checkForSplitBrain(nodes)
	if split {
		fmt.Println("The brain is split!")
		printStatus = true
		if strict {
			os.Exit(1)
		}
	} else {
		fmt.Println("Everything is ok")
	}
	if printStatus {
		masters := gatherMasters(nodes)
		printMasterNodes(masters)
	}
	failures := gatherFailures(nodes)
	printFailures(failures)

}

func checkForSplitBrain(nodes []ElasticsearchNode) bool {
	for i := 1; i < len(nodes); i++ {
		if nodes[i].MasterNode == nodes[i-1].MasterNode {
			return false
		}
	}
	return true
}

func fetchNodes(esAddresses []string) []ElasticsearchNode {
	nodes := make([]ElasticsearchNode, len(esAddresses))
	nodesChan := make(chan ElasticsearchNode, len(esAddresses))
	for _, node := range esAddresses {
		go asyncFetchNode(node, nodesChan)
	}
	for i := 0; i < len(esAddresses); i++ {
		fetchedNode := <-nodesChan
		nodes[i] = fetchedNode
	}
	return nodes
}

func asyncFetchNode(node string, nodesChan chan ElasticsearchNode) {
	defer func() {
		if r := recover(); r != nil {
			esNode := ElasticsearchNode{node, 0, "", 0, true}
			nodesChan <- esNode
		}
	}()
	ns := getNodeStatus(node)
	cs := getClusterState(node)
	esNode := ElasticsearchNode{
		ns.Name,
		ns.Status,
		cs.Nodes[cs.MasterNode].Name,
		len(cs.Nodes),
		false,
	}
	nodesChan <- esNode
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

func printMasterNodes(ms map[string][]ElasticsearchNode) {
	for master, nodes := range ms {
		fmt.Printf("master: %s \n", master)
		for i, node := range nodes {
			fmt.Printf("  node %d: %s \n", i, node.Name)
		}
	}
}

func printFailures(failures []string) {
	fmt.Println("Failed connecting to:")
	for _, failure := range failures {
		fmt.Printf("  %s\n", failure)
	}
}

func gatherMasters(nodes []ElasticsearchNode) map[string][]ElasticsearchNode {
	mappedMasters := make(map[string][]ElasticsearchNode)
	for _, node := range nodes {
		if node.ErrorFetching == false {
			mappedMasters[node.MasterNode] = append(mappedMasters[node.MasterNode], node)
		}
	}
	return mappedMasters
}

func gatherFailures(nodes []ElasticsearchNode) []string {
	failedFetching := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.ErrorFetching {
			failedFetching = append(failedFetching, node.Name)
		}
	}
	return failedFetching
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
