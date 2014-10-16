package clusterstatus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const node1StatusResposnse = `{
  "status" : 200,
    "name" : "Node1",
  "version" : {
      "number" : "1.3.4",
      "build_hash" : "a70f3ccb52200f8f2c87e9c370c6597448eb3e45",
      "build_timestamp" : "2014-09-30T09:07:17Z",
      "build_snapshot" : false,
      "lucene_version" : "4.9"
    },
  "tagline" : "You Know, for Search"
  }`
const node2StatusResposnse = `{
  "status" : 200,
    "name" : "Node2",
  "version" : {
      "number" : "1.3.4",
      "build_hash" : "a70f3ccb52200f8f2c87e9c370c6597448eb3e45",
      "build_timestamp" : "2014-09-30T09:07:17Z",
      "build_snapshot" : false,
      "lucene_version" : "4.9"
    },
  "tagline" : "You Know, for Search"
  }`
const node3StatusResposnse = `{
  "status" : 200,
    "name" : "Node3",
  "version" : {
      "number" : "1.3.4",
      "build_hash" : "a70f3ccb52200f8f2c87e9c370c6597448eb3e45",
      "build_timestamp" : "2014-09-30T09:07:17Z",
      "build_snapshot" : false,
      "lucene_version" : "4.9"
    },
  "tagline" : "You Know, for Search"
  }`
const nodeClusterResponse = `{
  "cluster_name" : "test-cluster",
  "master_node" : "Node1UniqeID",
  "nodes" : {
    "Node1UniqeID" : {
      "name" : "Node1",
      "transport_address" : "inet[/10.0.0.1:9300]",
      "attributes" : {
        "aws_zone" : "a"
      }
    },
    "Node2UniqeID" : {
      "name" : "Node2",
      "transport_address" : "inet[/10.0.0.2:9300]",
      "attributes" : {
        "aws_zone" : "b"
      }
    },
    "Node3UniqeID" : {
      "name" : "Node3",
      "transport_address" : "inet[/10.0.0.3:9300]",
      "attributes" : {
        "aws_zone" : "b"
      }
    }
  }
}`

var brokenCluster []ElasticsearchNode = []ElasticsearchNode{
	ElasticsearchNode{"Node1", 200, "Node1", 2, false},
	ElasticsearchNode{"Node2", 200, "Node1", 2, false},
	ElasticsearchNode{"Node3", 200, "Node3", 2, false},
}

var healthyCluster []ElasticsearchNode = []ElasticsearchNode{
	ElasticsearchNode{"Node1", 200, "Node1", 3, false},
	ElasticsearchNode{"Node2", 200, "Node1", 3, false},
	ElasticsearchNode{"Node3", 200, "Node1", 3, false},
}

func TestCheckForSplitBrainWhenSplitBrain(t *testing.T) {
	const expectedSplitBrainResult = false
	splitBrain := CheckForSplitBrain(healthyCluster)
	if expectedSplitBrainResult != splitBrain {
		t.Errorf("Split brain not detected. Expected %v. Got %v", expectedSplitBrainResult, splitBrain)
	}
}

func TestCheckForSplitBrainWhenNoSplitBrain(t *testing.T) {
	const expectedSplitBrainResult = true
	splitBrain := CheckForSplitBrain(brokenCluster)
	if expectedSplitBrainResult != splitBrain {
		t.Errorf("Split brain not detected. Expected %v. Got %v", expectedSplitBrainResult, splitBrain)
	}
}

func TestFetchNodes(t *testing.T) {
	const expectedFailedNodes = 0
	const expectedSuccessfullNodes = 3
	node1 := mockNodeServer(node1StatusResposnse, nodeClusterResponse)
	defer node1.Close()
	node2 := mockNodeServer(node2StatusResposnse, nodeClusterResponse)
	defer node2.Close()
	node3 := mockNodeServer(node3StatusResposnse, nodeClusterResponse)
	defer node3.Close()

	nodesSuccessfull, nodesFailed := FetchNodes([]string{node1.URL, node2.URL, node3.URL})
	if failedNodesAmount := len(nodesFailed); failedNodesAmount > expectedFailedNodes {
		t.Errorf("Failed nodes amount mismatch. Expected %d. Got %d", expectedFailedNodes, failedNodesAmount)
	}
	if successfullNodesAmount := len(nodesSuccessfull); successfullNodesAmount > expectedSuccessfullNodes {
		t.Errorf("Successfull nodes amount mismatch. Expected %d. Got %d", expectedSuccessfullNodes, successfullNodesAmount)
	}
}

func TestFetchNodesOneNodeMissing(t *testing.T) {
	const expectedFailedNodes = 1
	const expectedSuccessfullNodes = 2
	node1 := mockNodeServer(node1StatusResposnse, nodeClusterResponse)
	defer node1.Close()
	node2 := mockNodeServer(node2StatusResposnse, nodeClusterResponse)
	defer node2.Close()
	node3 := mockNodeServerFailing()
	defer node3.Close()

	nodesSuccessfull, nodesFailed := FetchNodes([]string{node1.URL, node2.URL, node3.URL})
	if failedNodesAmount := len(nodesFailed); failedNodesAmount > expectedFailedNodes {
		t.Errorf("Failed nodes amount mismatch. Expected %d. Got %d", expectedFailedNodes, failedNodesAmount)
	}
	if successfullNodesAmount := len(nodesSuccessfull); successfullNodesAmount > expectedSuccessfullNodes {
		t.Errorf("Successfull nodes amount mismatch. Expected %d. Got %d", expectedSuccessfullNodes, successfullNodesAmount)
	}
}

func TestGatherMasters(t *testing.T) {
	const expectedMasterNodesAmount = 2
	const expectedNodesCountGood = 3
	const expectedNodesCountBad = 1
	node1 := ElasticsearchNode{"Node1", 200, "Node1", 2, false}
	node2 := ElasticsearchNode{"Node2", 200, "Node1", 2, false}
	node3 := ElasticsearchNode{"Node3", 200, "Node1", 2, false}
	node4 := ElasticsearchNode{"Node4", 200, "Node4", 2, false}
	nodes := []ElasticsearchNode{node1, node2, node3, node4}
	masterNodes := GatherMasters(nodes)
	if masterNodesAmount := len(masterNodes); masterNodesAmount != expectedMasterNodesAmount {
		t.Errorf("Master nodes in cluster mismatch. Expected %d. Got %d", expectedMasterNodesAmount, masterNodesAmount)
	}
	if nodeCountGood := len(masterNodes["Node1"]); nodeCountGood != expectedNodesCountGood {
		t.Errorf("Nodes amount for good master mismatch. Expected %d. Got %d", expectedNodesCountGood, nodeCountGood)
	}
	if nodeCountBad := len(masterNodes["Node4"]); nodeCountBad != expectedNodesCountBad {
		t.Errorf("Nodes amount for bad master mismatch. Expected %d. Got %d", expectedNodesCountBad, nodeCountBad)
	}
}

func TestAmIMasterIfYes(t *testing.T) {
	me := mockNodeServer(node1StatusResposnse, nodeClusterResponse)
	defer me.Close()
	if amI := AmIMaster(me.URL); amI != true {
		t.Errorf("I am master but test said no")
	}
}

func TestAmIMasterIfNo(t *testing.T) {
	me := mockNodeServer(node2StatusResposnse, nodeClusterResponse)
	defer me.Close()
	if amI := AmIMaster(me.URL); amI != false {
		t.Errorf("I am not master but test said yes")
	}
}

func mockNodeServer(statusResponse, clusterResponse string) *httptest.Server {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		default:
			fmt.Fprintln(w, "`{}`")
		case "/_cluster/state/nodes,master_node":
			fmt.Fprintln(w, clusterResponse)
		case "/", "":
			fmt.Fprintln(w, statusResponse)
		}
	}))
	return node
}
func mockNodeServerFailing() *httptest.Server {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}))
	return node
}
