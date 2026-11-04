package claims

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"testing"
)

// cobra-viper/SKILL.md §6, go/SKILL.md (encoding/json), go-spec-reviewer/SKILL.md:
// on Go 1.27 the v1 encoding/json API keeps v1 behavior on the v2 implementation.
func TestJSONv1AcceptsDuplicateNamesAndInvalidUTF8(t *testing.T) {
	var m map[string]string
	if err := json.Unmarshal([]byte(`{"host":"localhost","host":"db.internal"}`), &m); err != nil {
		t.Fatalf("duplicate names: %v", err)
	}
	if m["host"] != "db.internal" {
		t.Errorf("duplicate names: got host %q, want the last value to win", m["host"])
	}
	var s string
	if err := json.Unmarshal([]byte("\"\xff\""), &s); err != nil {
		t.Fatalf("invalid UTF-8: %v", err)
	}
	if s != "�" {
		t.Errorf("invalid UTF-8: got %q, want U+FFFD", s)
	}
}

// Same sections: encoding/json/v2 is the strict one.
func TestJSONv2RejectsDuplicateNamesAndInvalidUTF8(t *testing.T) {
	var m map[string]string
	if err := jsonv2.Unmarshal([]byte(`{"host":"localhost","host":"db.internal"}`), &m); err == nil {
		t.Error("duplicate names: encoding/json/v2 accepted them")
	}
	var s string
	if err := jsonv2.Unmarshal([]byte("\"\xff\""), &s); err == nil {
		t.Error("invalid UTF-8: encoding/json/v2 accepted it")
	}
}
