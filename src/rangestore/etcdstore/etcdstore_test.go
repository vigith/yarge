package etcdstore

import (
	"log"
	"os"
	"testing"
)

var e *EtcdStore

// This is for setup and tear down.
func TestMain(m *testing.M) {
	var err error
	var status int

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

// checkIsLeafNode
func TestEtcdCheckIsLeafNode(t *testing.T) {
	var status bool
	var err error
	var node = "ops"
	status, err = e.checkIsLeafNode(node)
	if status || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS NOT a LeafNode, Got [bool:%v, error:%s]", node, status, err)
	}

	node = "ops-prod-vpc1-range"
	status, err = e.checkIsLeafNode(node)
	if !status || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS a LeafNode, Got [bool:%v, error:%s]", node, status, err)
	}

	node = "ops-prod-vpc1-foobar"
	status, err = e.checkIsLeafNode(node)
	if status != false || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS NOT a LeafNode, NO such range, Got [bool:%v, error:%s]", node, status, err)
	}
}
