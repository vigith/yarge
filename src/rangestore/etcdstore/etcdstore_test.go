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

	// we have tear down
	e.DisconnectEtcdStore()

	os.Exit(status)
}

// test KeyLookup
func TestKeyLookup(t *testing.T) {
	var cluster []string
	var err error
	var result *[]string
	var expected []string
	var key string

	cluster = []string{"ops-prod-vpc1-range"}
	key = "AUTHORS"
	expected = []string{"Vigith Maurice"}
	result, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-prod-vpc1-log"}
	key = "AUTHORS"
	expected = []string{"data@example.com"}
	result, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "QAFOR"
	expected = []string{"data"}
	result, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "KEYS"
	expected = []string{"AUTHORS", "NODES", "QAFOR"}
	result, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "FOOBAR"
	expected = []string{}
	result, err = e.KeyLookup(&cluster, key)
	if err == nil || !compare(*result, expected) {
		t.Errorf("Expected ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}
}

// test ClusterLookup
func TestClusterLookup(t *testing.T) {
	var cluster []string
	var err error
	var result *[]string
	var expected []string

	cluster = []string{"ops-prod-vpc1-range"}
	expected = []string{"range1001.ops.example.com", "range1002.ops.example.com", "range1003.ops.example.com"}
	result, err = e.ClusterLookup(&cluster)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, Cluster: %s, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}

	cluster = []string{"ops-prod-vpc1"}
	expected = []string{"ops-prod-vpc1-range", "ops-prod-vpc1-mon"}
	result, err = e.ClusterLookup(&cluster)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, Cluster: %s, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}

	cluster = []string{"ops-prod-vpc1-foobar"}
	expected = []string{}
	result, err = e.ClusterLookup(&cluster)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected ERROR, Cluster: %s IS NOT Present, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}
}

// test listClusters
func TestListClusters(t *testing.T) {
	var result []string
	var err error
	var expected []string
	var node string
	node = "ops-prod"
	result, err = e.listClusters(node)
	expected = []string{"ops-prod-vpc1", "ops-prod-vpc2"}
	if !compare(result, expected) || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS NOT a LeafNode, Got [result:%s, error:%s]", node, result, err)
	}

	node = "ops-foobar"
	result, err = e.listClusters(node)
	expected = []string{}
	if !compare(result, expected) || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] is not present, Got [result:%s, error:%s]", node, result, err)
	}
}

// checkIsLeafNode
func TestCheckIsLeafNode(t *testing.T) {
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
