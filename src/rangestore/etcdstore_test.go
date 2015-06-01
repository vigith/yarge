package rangestore

import (
	//	"log"
	"testing"
)

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
	if status || err == nil {
		t.Errorf("Expected ERROR, node [%s] IS NOT a LeafNode, NO such range, Got [bool:%v, error:%s]", node, status, err)
	}
}
