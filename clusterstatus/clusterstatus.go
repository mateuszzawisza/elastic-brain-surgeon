package clusterstatus

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

func CheckForSplitBrain(nodes []ElasticsearchNode) bool {
	for i := 1; i < len(nodes); i++ {
		if nodes[i].MasterNode != nodes[i-1].MasterNode {
			return true
		}
	}
	return false
}

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

func PrintMasterNodes(ms map[string][]ElasticsearchNode) {
	for master, nodes := range ms {
		fmt.Printf("master: %s \n", master)
		for i, node := range nodes {
			fmt.Printf("  node %d: %s \n", i, node.Name)
		}
	}
}

func PrintFailures(failures []ElasticsearchNode) {
	fmt.Println("Failed connecting to:")
	for _, failure := range failures {
		fmt.Printf("  %s\n", failure.Name)
	}
}

func GatherMasters(nodes []ElasticsearchNode) map[string][]ElasticsearchNode {
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
