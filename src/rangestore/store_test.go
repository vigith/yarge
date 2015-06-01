package rangestore

import (
	"log"
	"os"
	"testing"
)

var f *FileStore
var e *EtcdStore

// This is for setup and tear down.
// the Only setup we require is to make sure
// the store dir exists
func TestMain(m *testing.M) {
	var err error
	var status int

	// filestore
	var dir = "./t"
	var depth = 3
	var ffast = false
	f, err = ConnectFileStore(dir, depth, ffast)
	if err != nil {
		log.Fatal("ConnectFileStore ", err)
	}
	// we don't have tear down
	status = m.Run()
	if status != 0 {
		os.Exit(status)
	}

	// etcdstore
	var hosts = []string{"http://127.0.0.1:13824"}
	var roptimize = false
	var efast = false
	var node = ""
	e, err = ConnectEtcdStore(hosts, roptimize, efast, node)
	if err != nil {
		log.Fatal("ConnectEtcdStore ", err)
	}

	status = m.Run()
	if status != 0 {
		os.Exit(status)
	}

	// we have tear down
	e.DisconnectEtcdStore()

	os.Exit(status)
}

// Internal Functions

// Compare 2 Arrays, items need not be in correct order
func compare(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	if len(arr1) == 0 {
		return true
	}

	var flag bool
	for _, value1 := range arr1 {
		for _, value2 := range arr2 {
			if value1 == value2 {
				flag = true
			}
		}
		if !flag {
			return false
		}
	}

	return true
}
