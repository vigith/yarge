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

// optimizedNodeReverseLookup
func TestOptimizedNodeReverseLookup(t *testing.T) {
	e.ROptimize = true
	key := "range1001.ops.example.com"
	results, err := e.optimizedNodeReverseLookup(key)
	expected := []string{"ops-prod-vpc1-range"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s) Expected: %s, Got: %s (Error: %s)", key, expected, *results, err)
	}
	e.ROptimize = false
}

// KeyReverseLookup (major testing is done in TestKeyReverseLookupHint)
func TestKeyReverseLookup(t *testing.T) {
	key := "range1001.ops.example.com"
	results, err := e.KeyReverseLookup(key)
	expected := []string{"ops-prod-vpc1-range"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s) Expected: %s, Got: %s (Error: %s)", key, expected, *results, err)
	}
}

// KeyReverseLookupAttr (major testing is done in TestKeyReverseLookupHint)
func TestKeyReverseLookupAttr(t *testing.T) {
	key := "data@example.com"
	attr := "AUTHORS"
	results, err := e.KeyReverseLookupAttr(key, attr)
	expected := []string{"data-prod-vpc1-log", "data-prod-vpc2-log", "data-prod-vpc3-log"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s) Expected: %s, Got: %s (Error: %s)", key, expected, *results, err)
	}
}

// KeyReverseLookupHint
func TestKeyReverseLookupHint(t *testing.T) {
	var err error
	var results *[]string
	var expected []string
	var key, attr, hint string

	key = "Ops"
	attr = "AUTHORS"
	hint = ""
	results, err = e.KeyReverseLookupHint(key, attr, hint)
	expected = []string{"ops-prod-vpc1-mon", "ops-prod-vpc2-mon"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s, Attr: %s, Hint: %s) Expected: %s, Got: %s (Error: %s)", key, attr, hint, expected, *results, err)
	}

	key = "Ops"
	attr = "AUTHORS"
	hint = ""
	// enable FastLookup
	e.FastLookup = true
	results, err = e.KeyReverseLookupHint(key, attr, hint)
	expected = []string{"ops-prod-vpc1-mon"}
	if err != nil || len(*results) != len(expected) {
		t.Errorf("Expected NO ERROR, (Key: %s, Attr: %s, Hint: %s, FastLookup: %t) Expected: %s, Got: %s (Error: %s)", key, attr, hint, e.FastLookup, expected, *results, err)
	}
	// toggle FastLookup back
	e.FastLookup = false

	// look for Ops in data, should return empty
	key = "Ops"
	attr = "AUTHORS"
	hint = "data"
	results, err = e.KeyReverseLookupHint(key, attr, hint)
	expected = []string{}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s, Attr: %s, Hint: %s, FastLookup: %t) Expected: %s, Got: %s (Error: %s)", key, attr, hint, e.FastLookup, expected, *results, err)
	}

	key = "data@example.com"
	attr = "AUTHORS"
	hint = "data-prod"
	results, err = e.KeyReverseLookupHint(key, attr, hint)
	expected = []string{"data-prod-vpc1-log", "data-prod-vpc2-log", "data-prod-vpc3-log"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Key: %s, Attr: %s, Hint: %s, FastLookup: %t) Expected: %s, Got: %s (Error: %s)", key, attr, hint, e.FastLookup, expected, *results, err)
	}
}

// test getAllLeafNodes
func TestGetAllLeafNodes(t *testing.T) {
	var results *[]string
	var err error
	var expected []string
	var root string

	root = ""
	results, err = e.getAllLeafNodes(root)
	expected = []string{"data-prod-vpc1-log", "data-prod-vpc2-log", "data-prod-vpc3-log", "data-qa-vpc5-log", "ops-prod-vpc1-mon", "ops-prod-vpc1-range", "ops-prod-vpc2-mon"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, Root: %s Expected: %s, Got: %s (Error: %s)", root, expected, *results, err)
	}

	root = "ops"
	results, err = e.getAllLeafNodes(root)
	expected = []string{"ops-prod-vpc1-mon", "ops-prod-vpc1-range", "ops-prod-vpc2-mon"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, Root: %s Expected: %s, Got: %s (Error: %s)", root, expected, *results, err)
	}

	root = "data-qa-vpc5-log"
	results, err = e.getAllLeafNodes(root)
	expected = []string{"data-qa-vpc5-log"}
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, Root: %s Expected: %s, Got: %s (Error: %s)", root, expected, *results, err)
	}
}

// test KeyLookup
func TestKeyLookup(t *testing.T) {
	var cluster []string
	var err error
	var results *[]string
	var expected []string
	var key string

	cluster = []string{"ops-prod-vpc1-range"}
	key = "AUTHORS"
	expected = []string{"Vigith Maurice"}
	results, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *results, err)
	}

	cluster = []string{"data-prod-vpc1-log"}
	key = "AUTHORS"
	expected = []string{"data@example.com"}
	results, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *results, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "QAFOR"
	expected = []string{"data"}
	results, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *results, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "KEYS"
	expected = []string{"AUTHORS", "NODES", "QAFOR"}
	results, err = e.KeyLookup(&cluster, key)
	if err != nil || !compare(*results, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *results, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "FOOBAR"
	expected = []string{}
	results, err = e.KeyLookup(&cluster, key)
	if err == nil || !compare(*results, expected) {
		t.Errorf("Expected ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *results, err)
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
	if err == nil || !compare(*result, expected) {
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
	if !compare(result, expected) || err == nil {
		t.Errorf("Expected ERROR, node [%s] is not present, Got [result:%s, error:%s]", node, result, err)
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
