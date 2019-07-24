package util

import (
	"encoding/json"
	"testing"
)

/**
 * Test the JSON-to-JSON string normalization
 */
func TestNormalizeJSON(t *testing.T) {
	str, err := NormalizeJSON(
		`{"FOO": "bar",
			"nEstED": {"bob": "second", "alice":"first"},
			"ObjectInArrays": [
				1, "two", {"THREE": 3}
			]
		}`)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if str != `{"FOO":"bar","ObjectInArrays":[1,"two",{"THREE":3}],"nEstED":{"alice":"first","bob":"second"}}` {
		t.Errorf("The result does not match")
	}
}

/**
 * Test extracting defaults from schema
 */
func TestExtractDefaults(t *testing.T) {
	const JSONSCHEMA_STUB = `{
	  "type": "object",
	  "properties": {
	  	"foo": {
	  		"type": "object",
	  		"properties": {
	  			"bar": {
	          "default": 3,
	          "type": "number"
	  			},
	  			"baz": {
	          "default": "bax",
	          "type": "string"
	  			}
	  		}
	  	}
	  }
	}`

	var config map[string]interface{}
	err := json.Unmarshal([]byte(JSONSCHEMA_STUB), &config)
	if err != nil {
		t.Errorf("Unable to load stub config: %s", err.Error())
	}

	values, err := DefaultJSONFromSchema(config)
	if err != nil {
		t.Errorf("Unable to extract defaults from schema: %s", err.Error())
	}

	bytes, err := json.Marshal(values)
	str := string(bytes)
	if err != nil {
		t.Errorf("Unable to stringify the result: %s", err.Error())
	}

	if str != `{"foo":{"bar":3,"baz":"bax"}}` {
		t.Errorf("The result does not match")
	}
}

/**
 * Test cleaning-up maps
 */
func TestCleanupJSON(t *testing.T) {
	const JSON_WITH_EMPTIES_STUB = `{
		"a.remains": 1,
		"b.remains": "foo",
		"c.vanishes": "",
		"d.vanishes": null,
		"e.nested": {
			"a.remains": 1,
			"b.remains": "foo",
			"c.vanishes": "",
			"d.vanishes": null
		},
		"f.empties": {
			"a.vanishes": null,
			"b.vanishes": {},
			"c.vanishes": [],
			"d.vanishes": ""
		}
	}`

	srcMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON_WITH_EMPTIES_STUB), &srcMap)
	if err != nil {
		t.Errorf("Unable to load stub data: %s", err.Error())
	}

	dstMap := CleanupJSON(interface{}(srcMap))
	if err != nil {
		t.Errorf("Unable to clean map: %s", err.Error())
	}

	bytes, err := json.Marshal(dstMap)
	str := string(bytes)
	if err != nil {
		t.Errorf("Unable to stringify the result: %s", err.Error())
	}

	if str != `{"a.remains":1,"b.remains":"foo","e.nested":{"a.remains":1,"b.remains":"foo"}}` {
		t.Errorf("The result does not match")
	}
}

/**
 * Test dict diff
 */
func TestGetDictDiff(t *testing.T) {
	const JSON_BASE = `{
		"a.str1": "not included",
		"a.str2": "",
		"c.int1": 44,
		"c.int2": 0,
		"d.arr1": [ 1,2,3 ],
		"d.arr2": [ 1,2,4 ],
		"d.arr3": [ 1,2,3 ],
		"e.map1": {
			"a": 1,
			"b": "test",
			"c": true
		},
		"e.map2": {
			"a": 1,
			"b": "test",
			"c": true
		}
	}`

	const JSON_DIFF = `{
		"a.str1": "not included",
		"a.str2": "included",
		"c.int1": 44,
		"c.int2": 45,
		"d.arr1": [ 1,2,3 ],
		"d.arr2": [ 1,2,3 ],
		"d.arr3": [ 1,2,3,4 ],
		"e.map1": {
			"a": 1,
			"b": "test",
			"c": true
		},
		"e.map2": {
			"a": 1,
			"b": "test",
			"c": true,
			"e": "new"
		}
	}`

	baseMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON_BASE), &baseMap)
	if err != nil {
		t.Errorf("Unable to load base stub data: %s", err.Error())
		return
	}

	newMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(JSON_DIFF), &newMap)
	if err != nil {
		t.Errorf("Unable to load diff stub data: %s", err.Error())
		return
	}

	diffMap := GetDictDiff(baseMap, newMap)

	bytes, err := json.Marshal(diffMap)
	str := string(bytes)
	if err != nil {
		t.Errorf("Unable to stringify the result: %s", err.Error())
		return
	}

	if str != `{"a.str2":"included","c.int2":45,"d.arr2":[1,2,3],"d.arr3":[1,2,3,4],"e.map2":{"e":"new"}}` {
		t.Errorf("The result does not match: %s", str)
		return
	}
}
