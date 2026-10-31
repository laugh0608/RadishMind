package strictjson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStrictContractBoundary(t *testing.T) {
	type document struct {
		ID    string `json:"id"`
		Value int    `json:"value"`
	}
	valid := []byte(`{"id":"item","value":1}`)
	var dst document
	if err := Decode(valid, &dst); err != nil {
		t.Fatal(err)
	}
	if _, err := ObjectFields(valid, "id", "value"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`null`, `{"id":null,"value":1}`, `{"id":"first","id":"last","value":1}`, `{"id":"x","value":1,"extra":false}`, `{"id":"x","value":1} {}`, `{"id":"x","value":[null]}`, `{`, strings.Repeat(" ", 262145), string([]byte{0xff})} {
		if err := Decode([]byte(bad), &dst); err == nil {
			t.Fatalf("accepted malformed contract %q", bad[:min(len(bad), 80)])
		}
	}
	for _, bad := range []string{`{"ID":"item","value":1}`, `{"id":"item"}`, `{"id":"item","value":1,"extra":true}`} {
		if _, err := ObjectFields([]byte(bad), "id", "value"); err == nil {
			t.Fatal("field presence / casing accepted")
		}
	}
	var nested map[string]json.RawMessage
	if err := Decode([]byte(`{"data":{"items":[1,"x",true]}}`), &nested); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(`{"data":{"id":1,"id":2}}`), &nested); err == nil {
		t.Fatal("nested duplicate accepted")
	}
}
