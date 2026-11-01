package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestInfoGolden(t *testing.T) {
	fixture(t)
	var info map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "info", "--json")), &info); err != nil {
		t.Fatal(err)
	}
	info["root"], info["index_path"] = "<root>", "<index>"
	got, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../testdata/golden/info.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	expected, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(expected) {
		t.Fatalf("%s != %s", got, expected)
	}
}
