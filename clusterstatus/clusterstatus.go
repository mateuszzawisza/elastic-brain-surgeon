package clusterstatus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const clusterStatusEndpoint = "/_cluster/state/nodes,master_node"
const nodeStatusEndpoint = "/"

var httpTimeout = time.Second * 5

// ClusterState is structure to keep response from /_cluster/state API call
type ClusterState struct {
	MasterNode  string          `json:"master_node"`
	ClusterName string          `json:"cluster_name"`
	Nodes       map[string]Node `json:"nodes"`
}

// Node is structure to keep response about nodes in ClusterState API call
type Node struct {
	Name             string `json:"name"`
	TransportAddress string `json:"transport_address"`
	Attributes       struct {
		AwsZone string `json:"aws_zone"`
	} `json:"attributes"`
}

// NodeStatus is structure to keep response from node status API call
type NodeStatus struct {
	Status int    `json:"status"`
	Name   string `json:"name"`
}

// ElasticsearchNode gathers information on Elasticsearch Node
type ElasticsearchNode struct {
	Name           string
	IPAddress      string
	Status         int
	MasterNode     string
	NodesInCluster int
	ErrorFetching  bool
}

// CheckForSplitBrain checks if there is more then one leader in the cluster
func CheckForSplitBrain(nodes []ElasticsearchNode) bool {
	for i := 1; i < len(nodes); i++ {
		if nodes[i].MasterNode != nodes[i-1].MasterNode {
			return true
		}
	}
	return false
}

// FetchNodes connects to group of nodes that it receives as a parameter and
// gathers information about cluster topology
func FetchNodes(esAddresses []string) ([]ElasticsearchNode, []ElasticsearchNode) {
	nodesSuccessfull := make([]ElasticsearchNode, 0, len(esAddresses))
	nodesFailed := make([]ElasticsearchNode, 0, len(esAddresses))
	nodesChan := make(chan ElasticsearchNode, len(esAddresses))
	for _, node := range esAddresses {
		go asyncFetchNode(node, nodesChan)
	}
	for i := 0; i < len(esAddresses); i++ {
		fetchedNode := <-nodesChan
		if fetchedNode.ErrorFetching {
			nodesFailed = append(nodesFailed, fetchedNode)
		} else {
			nodesSuccessfull = append(nodesSuccessfull, fetchedNode)
		}

	}
	return nodesSuccessfull, nodesFailed
}

func asyncFetchNode(node string, nodesChan chan ElasticsearchNode) {
	ns, nsErr := getNodeStatus(node)
	cs, csErr := getClusterState(node)
	esNode := ElasticsearchNode{node, "", 0, "", 0, true}

	if nsErr == nil && csErr == nil {
		esNode.Name = ns.Name
		esNode.IPAddress = getIPAddressForName(ns.Name, cs.Nodes)
		esNode.Status = ns.Status
		esNode.MasterNode = cs.Nodes[cs.MasterNode].Name
		esNode.NodesInCluster = len(cs.Nodes)
		esNode.ErrorFetching = false
	}
	nodesChan <- esNode
}

func getIPAddressForName(name string, nodes map[string]Node) string {
	for _, node := range nodes {
		if name == node.Name {
			return node.TransportAddress
		}
	}
	return ""
}

func getClusterState(address string) (ClusterState, error) {
	address = normalizeAddress(address)
	statusEndpoint := address + clusterStatusEndpoint
	resp, err := makeHTTPCall(statusEndpoint)
	if err != nil {
		log.Println("could not connect to node")
		return ClusterState{}, errors.New("could not connect to node")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusInternalServerError {
		log.Println("node has failed")
		return ClusterState{}, errors.New("node has failed")
	}
	var cs ClusterState
	json.Unmarshal(body, &cs)
	return cs, nil
}

func makeHTTPCall(endpoint string) (*http.Response, error) {
	var client = &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(endpoint)
	return resp, err
}

func getNodeStatus(address string) (NodeStatus, error) {
	address = normalizeAddress(address)
	statusEndpoint := address + nodeStatusEndpoint
	resp, err := makeHTTPCall(statusEndpoint)
	if err != nil {
		log.Println("could not connect to node")
		return NodeStatus{}, errors.New("could not connect to node")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusInternalServerError {
		log.Println("node has failed")
		return NodeStatus{}, errors.New("node has failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading body: %v\n", err)
		return NodeStatus{}, fmt.Errorf("error reading body: %v", err)
	}
	var ns NodeStatus
	json.Unmarshal(body, &ns)
	return ns, nil
}

// PrintMasterNodes prints all master nodes along with their children
func PrintMasterNodes(ms map[string][]ElasticsearchNode) {
	for master, nodes := range ms {
		fmt.Printf("master: %s \n", master)
		for i, node := range nodes {
			nodeType := "node"
			if node.Name == master {
				nodeType = "master"
			}
			fmt.Printf("  node %d: %s - %s [%s]\n", i, node.Name, node.IPAddress,
				nodeType)
		}
	}
}

// PrintFailures prints nodes which we failed to connect to
func PrintFailures(failures []ElasticsearchNode) {
	fmt.Println("Failed connecting to:")
	for _, failure := range failures {
		fmt.Printf("  %s\n", failure.Name)
	}
}

// GatherMasters retruns a map with all master nodes along with their children
func GatherMasters(nodes []ElasticsearchNode) map[string][]ElasticsearchNode {
	mappedMasters := make(map[string][]ElasticsearchNode)
	for _, node := range nodes {
		if node.ErrorFetching == false {
			mappedMasters[node.MasterNode] = append(mappedMasters[node.MasterNode], node)
		}
	}
	return mappedMasters
}

// AmIMaster checks if given Elasticseach address is a leader of the cluster
func AmIMaster(myAddress string) (bool, error) {
	nodeStatus, gNerr := getNodeStatus(myAddress)
	clusterStatus, gCerr := getClusterState(myAddress)
	if gNerr != nil {
		return false, gNerr
	}
	if gCerr != nil {
		return false, gCerr
	}
	masterNode := clusterStatus.Nodes[clusterStatus.MasterNode].Name
	return masterNode == nodeStatus.Name, nil
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

func normalizeAddress(address string) string {
	httpPrefix := "http://"
	startsWithHTTP := strings.HasPrefix(address, httpPrefix)
	if startsWithHTTP {
		return address
	}
	return (httpPrefix + address)
}
