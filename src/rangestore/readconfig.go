// Given an YAML String and Key, it will return the value
// as an array or return error "No Such Key"

package rangestore

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
)

// We expect the YAML data to be in key value, where value is
// an array. Incase value is not an array, we will still return
// as an array
func yamlKeyLookup(content []byte, key string) (*[]string, error) {
	var u map[string]interface{}
	var err error
	err = yaml.Unmarshal(content, &u)
	// if unmarshal fails, return early with error
	if err != nil {
		return &[]string{}, err
	}

	// check whether the map has the key we are looking for
	value, ok := u[key]
	if !ok {
		return &[]string{}, errors.New(fmt.Sprintf("Cannot find Key [%s]", key))
	}

	// try to return result pointer to an array of strings
	switch value.(type) {
	// if it is an array
	case []interface{}:
		var results = make([]string, 0)
		for _, elem := range value.([]interface{}) {
			switch elem.(type) {
			case string:
				results = append(results, elem.(string))
			case int:
				results = append(results, fmt.Sprintf("%d", elem.(int)))
			case bool:
				results = append(results, fmt.Sprintf("%t", elem.(bool)))
			}
		}
		return &results, nil
		// if not an array, make it an array
	case string:
		return &[]string{value.(string)}, nil
	case int:
		return &[]string{fmt.Sprintf("%d", value.(int))}, nil
	case bool:
		return &[]string{fmt.Sprintf("%t", value.(bool))}, nil
	}

	return &[]string{}, nil
}
