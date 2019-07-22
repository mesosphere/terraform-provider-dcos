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
