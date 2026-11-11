package apijson_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

func TestParamUnknownTopLevelKeysPreservedAsExtras(t *testing.T) {
	// Pins the fork patch in subfield.go: unknown top-level keys on a param
	// (APIObject-embedding) struct must land in ExtraFields on decode, so a
	// decode -> re-encode round trip does not drop them (v1 SDK behavior;
	// the proxy pipeline relies on it for temperature/top_p/top_k). Uses the
	// real MessageNewParams because extras re-emission on encode goes through
	// the generated MarshalJSON -> param.MarshalObject path.
	raw := `{"model":"m","max_tokens":64,
		"messages":[{"role":"user","content":"hi"}],
		"temperature":0.2,"top_p":0.9,"top_k":40,"custom":{"nested":[1,2]}}`
	var dst anthropic.MessageNewParams
	if err := json.Unmarshal([]byte(raw), &dst); err != nil {
		t.Fatalf("failed unmarshal: %v", err)
	}

	extras := dst.ExtraFields()
	want := map[string]any{
		"temperature": 0.2,
		"top_p":       0.9,
		"top_k":       float64(40),
		"custom":      map[string]any{"nested": []any{float64(1), float64(2)}},
	}
	if !reflect.DeepEqual(extras, want) {
		t.Fatalf("expected extras\n%#v\nbut got\n%#v", want, extras)
	}

	// Re-encode must faithfully re-emit the unknown keys.
	again, err := json.Marshal(&dst)
	if err != nil {
		t.Fatalf("failed marshal: %v", err)
	}
	var round map[string]any
	if err := json.Unmarshal(again, &round); err != nil {
		t.Fatalf("failed re-decode: %v", err)
	}
	for k, v := range want {
		got, ok := round[k]
		if !ok {
			t.Fatalf("key %q missing after round trip: %s", k, again)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("key %q: expected %#v but got %#v", k, v, got)
		}
	}
}
