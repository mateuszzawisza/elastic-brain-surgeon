package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mateuszzawisza/elastic-brain-surgeon/clusterstatus"
)

// Version of package
const Version = "0.1.0"

var esAddresses addresses
var strict bool
var printJSON bool
var version bool

var exitStatus = 0

func init() {
	flag.Var(&esAddresses, "elasticsearch-list", "comma sperated list of elasticsearch instances addresses")
	flag.BoolVar(&strict, "strict", false, "Strict exit status")
	flag.BoolVar(&printJSON, "json", false, "Output in JSON")
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
		if strict {
			exitStatus = 1
		}
	}

	masters := clusterstatus.GatherMasters(nodes)
	if printJSON {
		jsonOutput, err := json.Marshal(masters)
		if err != nil {
			log.Panicf("Got error on json Marshal: %v", err)
		}
		os.Stdout.Write(jsonOutput)
	} else {
		if split {
			fmt.Println("The brain is split!")
		} else {
			fmt.Println("Everything is ok")
		}
		clusterstatus.PrintMasterNodes(masters)
		if len(nodesFailed) > 0 {
			clusterstatus.PrintFailures(nodesFailed)
		}
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
