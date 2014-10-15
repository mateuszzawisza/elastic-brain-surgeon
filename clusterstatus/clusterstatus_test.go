package clusterstatus

import "testing"

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
