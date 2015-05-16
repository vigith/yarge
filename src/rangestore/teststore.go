// This file is only used for unit testing. We should be able to test
// the rangeexpr operations without really connecting to a real store

package rangestore

import (
	"errors"
)

type TestStore struct {
	Test string // not for any good purpose
}

// function to connect to test store
func ConnectTestStore(test string) (t TestStore, err error) {
	t = TestStore{Test: test}
	return t, nil
}

// test clusterlookup
func (t TestStore) ClusterLookup(cluster string) ([]string, error) {
	if cluster == "error" {
		return []string{}, errors.New("I am asked to return 'error'")
	}

	return []string{"node1", "node2", "node3"}, nil
}

// test KeyLookup
func (t TestStore) KeyLookup(cluster string, key string) ([]string, error) {
	if cluster == "error" {
		return []string{}, errors.New("I am asked to return 'error'")
	}

	return []string{"foo"}, nil
}

// test KeyReverseLookup
func (t TestStore) KeyReverseLookup(cluster string, key string) ([]string, error) {
	if cluster == "error" {
		return []string{}, errors.New("I am asked to return 'error'")
	}

	return []string{"node1"}, nil
}

// test KeyReverseLookupAttr
func (t TestStore) KeyReverseLookupAttr(cluster string, key string, attr string) ([]string, error) {
	if cluster == "error" {
		return []string{}, errors.New("I am asked to return 'error'")
	}

	return []string{"node2"}, nil
}
