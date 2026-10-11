package anthropic

// Tingly regression tests: pin behaviors that tingly-dev once had to patch.
// The patches were dropped once upstream fixed them, but these tests are
// carried forward on every tingly-dev-* branch so an upstream regression is
// caught at fork time instead of in production.

import (
	"encoding/json"
	"strings"
	"testing"
)

// Interleaved stream events (a new content block starts before the previous
// one stops) must accumulate deltas into the block named by their index, not
// the most recently started block. Was patched in tingly-dev-v1.45.0.
var interleavedContentBlockEvents = []string{
	`{"type": "message_start", "message": {}}`,
	`{"type": "content_block_start", "index": 0, "content_block": {"type": "thinking", "thinking": ""}}`,
	`{"type": "content_block_delta", "index": 0, "delta": {"type": "thinking_delta", "thinking": "Let me think."}}`,
	`{"type": "content_block_delta", "index": 0, "delta": {"type": "signature_delta", "signature": "sig123"}}`,
	`{"type": "content_block_stop", "index": 0}`,
	`{"type": "content_block_start", "index": 1, "content_block": {"type": "text", "text": ""}}`,
	`{"type": "content_block_delta", "index": 1, "delta": {"type": "text_delta", "text": "Hello"}}`,
	`{"type": "content_block_start", "index": 2, "content_block": {"type": "thinking", "thinking": ""}}`,
	`{"type": "content_block_delta", "index": 1, "delta": {"type": "text_delta", "text": " world"}}`,
	`{"type": "content_block_delta", "index": 1, "delta": {"type": "text_delta", "text": "!"}}`,
	`{"type": "content_block_stop", "index": 1}`,
	`{"type": "content_block_stop", "index": 2}`,
	`{"type": "message_stop"}`,
}

func assertSameJSON(t *testing.T, expected, actual any) {
	t.Helper()
	e, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	a, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	if string(e) != string(a) {
		t.Fatalf("Mismatched message: expected %s but got %s", e, a)
	}
}

func TestTinglyAccumulateInterleavedContentBlocks(t *testing.T) {
	message := Message{}
	for _, eventStr := range interleavedContentBlockEvents {
		event := MessageStreamEventUnion{}
		if err := (&event).UnmarshalJSON([]byte(eventStr)); err != nil {
			t.Fatal(err)
		}
		if err := (&message).Accumulate(event); err != nil {
			t.Fatal(err)
		}
	}
	assertSameJSON(t, Message{Content: []ContentBlockUnion{
		{Type: "thinking", Thinking: "Let me think.", Signature: "sig123"},
		{Type: "text", Text: "Hello world!"},
		{Type: "thinking"},
	}}, message)
}

func TestTinglyBetaAccumulateInterleavedContentBlocks(t *testing.T) {
	message := BetaMessage{}
	for _, eventStr := range interleavedContentBlockEvents {
		event := BetaRawMessageStreamEventUnion{}
		if err := (&event).UnmarshalJSON([]byte(eventStr)); err != nil {
			t.Fatal(err)
		}
		if err := (&message).Accumulate(event); err != nil {
			t.Fatal(err)
		}
	}
	assertSameJSON(t, BetaMessage{Content: []BetaContentBlockUnion{
		{Type: "thinking", Thinking: "Let me think.", Signature: "sig123"},
		{Type: "text", Text: "Hello world!"},
		{Type: "thinking"},
	}}, message)
}

// thinking {"type":"adaptive"} must decode into the OfAdaptive variant and
// survive a re-marshal. Was patched in tingly-dev-v1.27.1 (missing union
// registration for the adaptive variant).
const adaptiveThinkingRequest = `{"model":"claude-opus-5-5","max_tokens":1024,"messages":[],"thinking":{"type":"adaptive"}}`

func TestTinglyThinkingAdaptiveDecode(t *testing.T) {
	var params MessageNewParams
	if err := json.Unmarshal([]byte(adaptiveThinkingRequest), &params); err != nil {
		t.Fatal(err)
	}
	if params.Thinking.OfAdaptive == nil {
		t.Fatal("thinking.type=adaptive did not decode into OfAdaptive")
	}
	out, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"thinking":{"type":"adaptive"`) {
		t.Fatalf("adaptive thinking lost on re-marshal: %s", out)
	}
}

func TestTinglyBetaThinkingAdaptiveDecode(t *testing.T) {
	var params BetaMessageNewParams
	if err := json.Unmarshal([]byte(adaptiveThinkingRequest), &params); err != nil {
		t.Fatal(err)
	}
	if params.Thinking.OfAdaptive == nil {
		t.Fatal("thinking.type=adaptive did not decode into OfAdaptive")
	}
	out, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"thinking":{"type":"adaptive"`) {
		t.Fatalf("adaptive thinking lost on re-marshal: %s", out)
	}
}
