package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

/**
 * Parses the given JSON string, sorts the keys and re-encodes to JSON
 */
func NormalizeJSON(inputJson string) (string, error) {
	var anyJson map[string]interface{}
	json.Unmarshal([]byte(inputJson), &anyJson)

	// The unmarshal logic actually sorts the keys, so there is nothing
	// else required to do.
	bytes, err := json.Marshal(anyJson)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

/**
 * Normalizes and hashes
 */
func HashDict(input map[string]interface{}) (string, error) {
	// JSON serializer serializes the keys in alphabetical order, so we
	// are certain that every time the result will be the same
	bytes, err := json.Marshal(CleanupJSON(input))
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", sum), nil
}

/**
 * Processes the given JSON schema and extracts the default values into a
 * configuration JSON object
 */
func DefaultJSONFromSchema(inputSchema map[string]interface{}) (map[string]map[string]interface{}, error) {
	defaultValue, err := defaultFromSchemaObject(inputSchema)
	if err != nil {
		return nil, err
	}

	// Convert to nested map interface, as required
	result := make(map[string]map[string]interface{})
	for key, value := range defaultValue {
		if mapValue, ok := interface{}(value).(map[string]interface{}); ok {
			result[key] = mapValue
		} else {
			return nil, fmt.Errorf("%s: Expecting a map", key)
		}
	}

	return result, nil
}

/**
 * NestedToFlatMap converts a map of map of interfaces to a map of interfaces
 */
func NestedToFlatMap(input map[string]map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range input {
		ret[k] = interface{}(v)
	}
	return ret
}

/**
 * FlatToNestedMap converts a map of interfaces to a map of map of interfaces
 */
func FlatToNestedMap(input map[string]interface{}) (map[string]map[string]interface{}, error) {
	ret := make(map[string]map[string]interface{})
	for k, v := range input {
		if vMap, ok := v.(map[string]interface{}); ok {
			ret[k] = vMap
		} else {
			return nil, fmt.Errorf("Key '%s' is not a map", k)
		}
	}
	return ret, nil
}

/**
 * Remove empty strings, nulls, empty arrays and empty objects from the given dict
 */
func CleanupJSON(input interface{}) interface{} {
	if inputMap, ok := input.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for key, value := range inputMap {
			switch v := value.(type) {
			case string:
				// Strings must not be empty
				if v != "" {
					newMap[key] = value
				}
			case []interface{}:
				// Arrays must not be empty
				if len(v) > 0 {
					var newArray []interface{} = nil
					for _, value := range v {
						newArray = append(newArray, CleanupJSON(value))
					}
					newMap[key] = newArray
				}
			case map[string]interface{}:
				// Objects must not be empty
				if len(v) > 0 {
					newMapValue := CleanupJSON(value).(map[string]interface{})
					if len(newMapValue) > 0 {
						newMap[key] = newMapValue
					}
				}
			default:
				if value != nil {
					newMap[key] = value
				}
			}
		}
		return newMap
	}
	return input
}

/**
 * Best-effort auto-typing of strings that follow the given patterns:
 *
 * 1) Numeric values --> float64
 * 2) "true" / "false" --> bool
 * 3) "null" --> nil
 * 5) <anything else> --> string
 */
func AutotypeValue(input interface{}) interface{} {
	if strValue, ok := input.(string); ok {
		if intVal, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			return intVal
		} else if floatVal, err := strconv.ParseFloat(strValue, 64); err == nil {
			return floatVal
		} else if strValue == "true" {
			return true
		} else if strValue == "false" {
			return false
		} else if strValue == "null" {
			return nil
		}
	}

	return input
}

/**
 * Processes the values of the given map and tries some best-effort type-casting
 */
func AutotypeMap(input map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for key, value := range input {
		ret[key] = AutotypeValue(value)
	}

	return ret
}

/**
 * Processes the values of the given slice and tries some best-effort type-casting
 */
func AutotypeList(input []interface{}) []interface{} {
	var ret []interface{} = nil
	for _, value := range input {
		ret = append(ret, AutotypeValue(value))
	}

	return ret
}

/**
 * Walk a {type: "object"} schema entry and return a map with the default values
 */
func defaultFromSchemaObject(input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	varType, ok := input["type"]
	if !ok {
		return nil, fmt.Errorf("Missing `type` property")
	}
	if varType.(string) != "object" {
		return nil, fmt.Errorf("Trying to process a non-object as object")
	}

	props, ok := input["properties"]
	if !ok {
		return result, nil
	}

	for key, value := range props.(map[string]interface{}) {
		if valueMap, ok := value.(map[string]interface{}); ok {
			defaultValue, err := defaultFromSchemaValue(valueMap)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", key, err.Error())
			}
			if defaultValue != nil {
				result[key] = defaultValue
			}
		}
	}

	return result, nil
}

/**
 * Walk a {type: "*"} schema entry with a default value and return it.
 * Otherwise returns `nil` if a default value is missing
 */
func defaultFromSchemaValue(input map[string]interface{}) (interface{}, error) {
	varType, ok := input["type"]
	if !ok {
		return nil, fmt.Errorf("Missing `type` property")
	}

	// Objects require some nesting
	if varType == "object" {
		return defaultFromSchemaObject(input)
	}

	// Otherwise, get the "default" field, if any
	defaultValue, ok := input["default"]
	if !ok {
		return nil, nil
	}

	return defaultValue, nil
}

/**
 * Normalize the given abstract JSON
 */
func normalizeDict(dict map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for key, value := range dict {
		key = strings.ToLower(key)

		if valueDict, ok := value.(map[string]interface{}); ok {
			newMap[key] = normalizeDict(valueDict)
		} else if valueArray, ok := value.([]interface{}); ok {
			var newArray []interface{} = nil
			for _, value := range valueArray {
				if valueDict, ok := value.(map[string]interface{}); ok {
					newArray = append(newArray, normalizeDict(valueDict))
				} else {
					newArray = append(newArray, value)
				}
			}
			newMap[key] = newArray
		} else {
			newMap[key] = value
		}
	}
	return newMap
}
