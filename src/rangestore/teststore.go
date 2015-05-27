// This file implmenets a TestStore used only for unit testing. We should be able
// to test the rangeexpr operations without really connecting to a real store

package rangestore

import (
	"errors"
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
        "QAFOR": [
            "data"
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
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"ops", "data"}, nil

	case "ops":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"ops-prod"}, nil
	case "ops-prod":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"ops-prod-vpc1", "ops-prod-vpc2"}, nil
	case "ops-prod-vpc1":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"ops-prod-vpc1-range", "ops-prod-vpc1-mon"}, nil
	case "ops-prod-vpc2":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"ops-prod-vpc2-mon"}, nil
	case "ops-prod-vpc1-range":
		if key == "NODES" {
			return &[]string{"range1001.ops.example.com", "range1002.ops.example.com", "range1003.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Vigith Maurice"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc2-range":
		if key == "NODES" {
			return &[]string{}, errors.New("Cluster Not Found (ops-prod-vpc2-range)")
		} else if key == "AUTHORS" {
			return &[]string{}, errors.New("Cluster Not Found (ops-prod-vpc2-range)")
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc1-mon":
		if key == "NODES" {
			return &[]string{"mon1001.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Ops"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}
	case "ops-prod-vpc2-mon":
		if key == "NODES" {
			return &[]string{"mon2001.ops.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"Ops"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}

	case "data":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-prod", "data-qa"}, nil
	case "data-prod":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-prod-vpc1", "data-prod-vpc2", "data-prod-vpc3"}, nil
	case "data-prod-vpc1":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-prod-vpc1-log"}, nil
	case "data-prod-vpc2":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-prod-vpc2-log"}, nil
	case "data-prod-vpc3":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-prod-vpc3-log"}, nil
	case "data-prod-vpc1-log":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		if key == "NODES" && key != "NODES" {
			return &[]string{"data1001.data.example.com", "data1002.data.example.com", "data1003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}
	case "data-prod-vpc2-log":
		if key == "NODES" {
			return &[]string{"data2001.data.example.com", "data2002.data.example.com", "data2003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}
		return &[]string{"log"}, nil
	case "data-prod-vpc3-log":
		if key == "NODES" {
			return &[]string{"data3001.data.example.com", "data3002.data.example.com", "data3003.data.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"data@example.com"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else {
			return &[]string{}, nil
		}

	case "data-qa":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-qa-vpc5"}, nil
	case "data-qa-vpc5":
		if key != "" && key != "NODES" {
			return &[]string{""}, errors.New("Not a leaf node")
		}
		return &[]string{"data-qa-vpc5-log"}, nil
	case "data-qa-vpc5-log":
		if key == "NODES" {
			return &[]string{"data5001.qa.example.com", "data5002.qa.example.com"}, nil
		} else if key == "AUTHORS" {
			return &[]string{"qa@example.com"}, nil
		} else if key == "KEYS" {
			return &[]string{"NODES", "AUTHORS"}, nil
		} else if key == "QAFOR" {
			return &[]string{"data"}, nil
		} else {
			return &[]string{}, nil
		}
	}

	return &[]string{}, errors.New("No entry found")
}

// test clusterlookup
func (t *TestStore) ClusterLookup(cluster *[]string) (*[]string, error) {

	if len(*cluster) > 0 && (*cluster)[0] == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	var results = make([]string, 0)
	for _, elem := range *cluster {
		result, err := queryMap(elem, "NODES")
		if err != nil {
			return &[]string{}, err
		}
		results = append(results, *result...)
	}
	return &results, nil
}

// test KeyLookup
func (t *TestStore) KeyLookup(cluster *[]string, key string) (*[]string, error) {
	if len(*cluster) > 0 && (*cluster)[0] == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	var results = make([]string, 0)
	for _, elem := range *cluster {
		result, err := queryMap(elem, key)
		if err != nil {
			return &[]string{}, err
		}
		results = append(results, *result...)
	}

	return &results, nil
}

func queryMapRev(key string, attr string, hint string) (*[]string, error) {
	switch key {
	case "range1001.ops.example.com", "range1002.ops.example.com", "range1003.ops.example.com":
		return &[]string{"ops-prod-vpc1-range"}, nil
	case "mon1001.ops.example.com":
		return &[]string{"ops-prod-vpc1-mon"}, nil
	case "mon1002.ops.example.com":
		return &[]string{"ops-prod-vpc2-mon"}, nil

	case "data1001.data.example.com", "data1002.data.example.com", "data1003.data.example.com":
		return &[]string{"data-prod-vpc1-log"}, nil
	case "data2001.data.example.com", "data2002.data.example.com", "data2003.data.example.com":
		return &[]string{"data-prod-vpc2-log"}, nil
	case "data3001.data.example.com", "data3002.data.example.com", "data3003.data.example.com":
		return &[]string{"data-prod-vpc3-log"}, nil

	case "data5001.qa.example.com", "data5002.qa.example.com":
		return &[]string{"data-qa-vpc5-log"}, nil

	case "Ops":
		if attr == "AUTHORS" {
			if hint == "ops-prod-vpc1-mon" || hint == "ops-prod-vpc2-mon" || hint == "" {
				return &[]string{"ops-prod-vpc1-mon", "ops-prod-vpc2-mon"}, nil
			}
			return &[]string{}, errors.New("Did not find any reverse lookup entry")
		}

	case "Vigith Maurice":
		if attr == "AUTHORS" {
			if hint == "ops-prod-vpc1-range1" || hint == "" {
				return &[]string{"ops-prod-vpc1-range"}, nil
			}
			return &[]string{}, errors.New("Did not find any reverse lookup entry")
		}
	}

	return &[]string{}, errors.New("Did not find any reverse lookup entry")
}

// test KeyReverseLookup
func (t *TestStore) KeyReverseLookup(key string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	result, err := queryMapRev(key, "", "")
	if err != nil {
		return &[]string{}, err
	}

	return result, nil
}

// test KeyReverseLookupAttr
func (t *TestStore) KeyReverseLookupAttr(key string, attr string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	result, err := queryMapRev(key, attr, "")
	if err != nil {
		return &[]string{}, err
	}

	return result, nil
}

// test KeyReverseLookupHint
func (t *TestStore) KeyReverseLookupHint(key string, attr string, hint string) (*[]string, error) {
	if key == "error" {
		return &[]string{}, errors.New("I am asked to return 'error'")
	}

	result, err := queryMapRev(key, attr, hint)
	if err != nil {
		return &[]string{}, err
	}

	return result, nil
}
