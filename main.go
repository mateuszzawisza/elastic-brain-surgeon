package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mateuszzawisza/elastic-brain-surgeon/clusterstatus"
)

const Version = "0.0.3"

var esAddresses addresses
var strict bool
var printStatus bool
var version bool

var exitStatus int = 0

func init() {
	flag.Var(&esAddresses, "elasticsearch-list", "comma sperated list of elasticsearch instances addresses")
	flag.BoolVar(&strict, "strict", false, "Strict exit status")
	flag.BoolVar(&printStatus, "print", false, "Print cluster status")
	flag.BoolVar(&version, "version", false, "Print version")
}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("Version %s\n", Version)
		return
	}
	nodes, nodesFailed := clusterstatus.FetchNodes(esAddresses)
	split := clusterstatus.CheckForSplitBrain(nodes)
	if split {
		fmt.Println("The brain is split!")
		printStatus = true
		if strict {
			exitStatus = 1
		}
	} else {
		fmt.Println("Everything is ok")
	}
	if printStatus {
		masters := clusterstatus.GatherMasters(nodes)
		clusterstatus.PrintMasterNodes(masters)
	}
	if len(nodesFailed) > 0 {
		clusterstatus.PrintFailures(nodesFailed)
	}
	os.Exit(exitStatus)
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
