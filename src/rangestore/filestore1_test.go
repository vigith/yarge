package rangestore

import (
	"log"
	"os"
	"testing"
)

var f *FileStore
var dir = "./t"
var depth = 3

// This is for setup and tear down.
// the Only setup we require is to make sure
// the store dir exists
func TestMain(m *testing.M) {
	var err error
	f, err = ConnectFileStore(dir, depth)
	if err != nil {
		log.Fatal(err)
	}
	// we don't have tear down
	os.Exit(m.Run())
}

func TestKeyLookup(t *testing.T) {
	var cluster []string
	var err error
	var result *[]string
	var expected []string
	var key string

	cluster = []string{"ops-prod-vpc1-range"}
	key = "AUTHORS"
	expected = []string{"Vigith Maurice"}
	result, err = f.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-prod-vpc1-log"}
	key = "AUTHORS"
	expected = []string{"data@example.com"}
	result, err = f.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, (Cluster: %s, Key: %s) Expected: %s, Got: %s (Error: %s)", cluster, key, expected, *result, err)
	}

	cluster = []string{"data-qa-vpc5-log"}
	key = "QAFOR"
	expected = []string{"data"}
	result, err = f.KeyLookup(&cluster, key)
	if err != nil || !compare(*result, expected) {
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
	result, err = f.ClusterLookup(&cluster)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, Cluster: %s, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}

	cluster = []string{"ops-prod-vpc1"}
	expected = []string{"range", "mon"}
	result, err = f.ClusterLookup(&cluster)
	if err != nil || !compare(*result, expected) {
		t.Errorf("Expected NO ERROR, Cluster: %s, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}

	cluster = []string{"ops-prod-vpc1-foobar"}
	expected = []string{}
	result, err = f.ClusterLookup(&cluster)
	if err == nil || !compare(*result, expected) {
		t.Errorf("Expected ERROR, Cluster: %s IS NOT Present, Expected: %s, Got: %s (Error: %s)", cluster, expected, *result, err)
	}
}

// test yamlKeyLookup
func TestYamlKeyLookup(t *testing.T) {
	var content string
	var key string
	var result *[]string
	var err error
	var expected []string

	content = `foo: true`
	key = "foo"
	result, err = yamlKeyLookup([]byte(content), key)
	expected = []string{"true"}
	if err != nil || !compare(expected, *result) {
		t.Errorf("Expected NO ERROR, Expected %s, Got %s (Error: %s)", expected, *result, err)
	}

	content = `foo: 1`
	key = "foo"
	result, err = yamlKeyLookup([]byte(content), key)
	expected = []string{"1"}
	if err != nil || !compare(expected, *result) {
		t.Errorf("Expected NO ERROR, Expected %s, Got %s (Error: %s)", expected, *result, err)
	}

	content = `foo: bar`
	key = "foo"
	result, err = yamlKeyLookup([]byte(content), key)
	expected = []string{"bar"}
	if err != nil || !compare(expected, *result) {
		t.Errorf("Expected NO ERROR, Expected %s, Got %s (Error: %s)", expected, *result, err)
	}

	content = `
foo: 
  - bar
  - 1
  - true
`
	key = "foo"
	result, err = yamlKeyLookup([]byte(content), key)
	expected = []string{"1", "bar", "true"}
	if err != nil || !compare(expected, *result) {
		t.Errorf("Expected NO ERROR, Expected %s, Got %s (Error: %s)", expected, *result, err)
	}

	content = `foo: bar`
	key = "bar"
	result, err = yamlKeyLookup([]byte(content), key)
	expected = []string{}
	if err == nil || !compare(expected, *result) {
		t.Errorf("Expected ERROR, No Key %s, Expected %s, Got %s (Error: %s)", key, expected, *result, err)
	}
}

// test listClusters
func TestListClusters(t *testing.T) {
	var result []string
	var err error
	var expected []string
	var node string
	node = "ops-prod"
	result, err = f.listClusters(node)
	expected = []string{"vpc1", "vpc2"}
	if !compare(result, expected) || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS NOT a LeafNode, Got [result:%s, error:%s]", node, result, err)
	}

	node = "ops-foobar"
	result, err = f.listClusters(node)
	expected = []string{""}
	if compare(result, expected) || err == nil {
		t.Errorf("Expected ERROR, node [%s] is not present, Got [result:%s, error:%s]", node, result, err)
	}
}

// test checkIsLeafNode
func TestCheckIsLeafNode(t *testing.T) {
	var status bool
	var err error
	var node = "ops"
	status, err = f.checkIsLeafNode(node)
	if status || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS NOT a LeafNode, Got [bool:%v, error:%s]", node, status, err)
	}

	node = "ops-prod-vpc1-range"
	status, err = f.checkIsLeafNode(node)
	if !status || err != nil {
		t.Errorf("Expected NO ERROR, node [%s] IS a LeafNode, Got [bool:%v, error:%s]", node, status, err)
	}

	node = "ops-prod-vpc1-foobar"
	status, err = f.checkIsLeafNode(node)
	if status || err == nil {
		t.Errorf("Expected ERROR, node [%s] IS NOT a LeafNode, NO such range, Got [bool:%v, error:%s]", node, status, err)
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
