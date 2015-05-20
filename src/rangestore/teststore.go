// This file implmenets a TestStore used only for unit testing. We should be able
// to test the rangeexpr operations without really connecting to a real store

package rangestore

import (
	"errors"
	"fmt"
	"rangeops"
)

// Test Range Structure
/*
{
	"data": {
		"prod": {
			"vpc1": {
				"log": {
				"AUTHORS": [
           "data@example.com"
					],
					"NODES": [
						"data1001.data.example.com",
						"data1002.data.example.com",
            "data1003.data.example.com"
                    ]
				}
			},
			"vpc2": {
				"log": {
					"AUTHORS": [
            "data@example.com"
					],
					"NODES": [
						"data2001.data.example.com",
						"data2002.data.example.com",
            "data2003.data.example.com"
                    ]
				}
			},
			"vpc3": {
				"log": {
					"AUTHORS": [
            "data@example.com"
					],
					"NODES": [
						"data3001.data.example.com",
						"data3002.data.example.com",
            "data3003.data.example.com"
                    ]
				}
			}
		},
		"qa": {
			"vpc5": {
				"log": {
					"AUTHORS": [
            "qa@example.com"
					],
					"NODES": [
						"data5001.qa.example.com",
            "data5002.qa.example.com"
                    ]
				}
			}
		}
	},
	"ops": {
		"prod": {
			"vpc1": {
				"mon": {
					"AUTHORS": [
            "Ops"
					],
					"NODES": [
            "mon1001.ops.example.com"
                    ]
				},
				"range": {
					"AUTHORS": [
            "Vigith Maurice"
					],
					"NODES": [
						"range1001.ops.example.com",
						"range1002.ops.example.com",
            "range1003.ops.example.com"
                    ]
				}
			},
			"vpc2": {
				"mon": {
					"AUTHORS": [
            "Ops"
					],
					"NODES": [
            "mon2001.ops.example.com"
                    ]
				}
			}
		}
	}
}
*/

type TestStore struct {
	Test string // not for any good purpose
}

// function to connect to test store
func ConnectTestStore(test string) (t *TestStore, err error) {
	t = &TestStore{Test: test}
	return t, nil
}

// query map
func queryMap(cluster string, key string) (*[]string, error) {
	switch cluster {
	case "RANGE":
		return &[]string{"ops", "data"}, nil

	case "ops":
		return &[]string{"prod"}, nil
	case "ops-prod":
		return &[]string{"vpc1", "vpc2"}, nil
	case "ops-prod-vpc1":
		return &[]string{"range", "mon"}, nil
	case "ops-prod-vpc2":
		return &[]string{"mon"}, nil
	case "ops-prod-vpc1-range":
		if key == "NODES" {
			return &[]string{"range1001.ops.example.com", "range1002.ops.example.com", "range1003.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Vigith Maurice"}, nil
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc2-range":
		if key == "NODES" {
			return &[]string{}, errors.New("Cluster Not Found (ops-prod-vpc2-range)")
		} else if key == "AUTHORS" {
			return &[]string{}, errors.New("Cluster Not Found (ops-prod-vpc2-range)")
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc1-mon":
		if key == "NODES" {
			return &[]string{"mon1001.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Ops"}, nil
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc2-mon":
		if key == "NODES" {
			return &[]string{"mon2001.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Ops"}, nil
		} else {
			return &[]string{}, nil
		}

	case "data":
		return &[]string{"prod", "qa"}, nil
	case "data-prod":
		return &[]string{"vpc1", "vpc2", "vpc3"}, nil
	case "data-prod-vpc1":
		return &[]string{"log"}, nil
	case "data-prod-vpc2":
		return &[]string{"log"}, nil
	case "data-prod-vpc3":
		return &[]string{"log"}, nil
	case "data-prod-vpc1-log":
		if key == "NODES" {
			return &[]string{"data1001.data.example.com", "data1002.data.example.com", "data1003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else {
			return &[]string{}, nil
		}
	case "data-prod-vpc2-log":
		if key == "NODES" {
			return &[]string{"data2001.data.example.com", "data2002.data.example.com", "data2003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else {
			return &[]string{}, nil
		}
		return &[]string{"log"}, nil
	case "data-prod-vpc3-log":
		if key == "NODES" {
			return &[]string{"data3001.data.example.com", "data3002.data.example.com", "data3003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else {
			return &[]string{}, nil
		}

	case "data-qa":
		return &[]string{"vpc5"}, nil
	case "data-qa-vpc5":
		return &[]string{"log"}, nil
	case "data-qa-vpc5-log":
		if key == "NODES" {
			return &[]string{"data5001.qa.example.com", "data5002.qa.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"qa@example.com"}, nil
		} else {
			return &[]string{}, nil
		}
	}

	return nil, nil
}

// test clusterlookup
func (t *TestStore) ClusterLookup(cluster *[]string) (*[]string, error) {

	if (*cluster)[0] == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	var results = make([]string, 0)
	for _, elem := range *cluster {
		result, err := queryMap(elem, "")
		if err != nil {
			return &[]string{}, errors.New("Mock Map Failed!!")
		}
		results = append(results, *result...)
	}

	rangeops.ArrayToSet(&results)
	return &results, nil
}

// test KeyLookup
func (t *TestStore) KeyLookup(cluster *[]string, key string) (*[]string, error) {
	if (*cluster)[0] == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	var results = make([]string, 0)
	for _, elem := range *cluster {
		result, err := queryMap(elem, key)
		if err != nil {
			return &[]string{}, errors.New("Mock Map Failed!!")
		}
		fmt.Println(result)
		results = append(results, *result...)
	}

	rangeops.ArrayToSet(&results)

	return &results, nil
}

// test KeyReverseLookup
func (t *TestStore) KeyReverseLookup(key string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	return &[]string{"node1"}, nil
}

// test KeyReverseLookupAttr
func (t *TestStore) KeyReverseLookupAttr(key string, attr string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	return &[]string{"node2"}, nil
}

// test KeyReverseLookupHint
func (t *TestStore) KeyReverseLookupHint(key string, attr string, hint string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	return &[]string{"node3"}, nil
}
