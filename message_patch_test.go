package anthropic

import (
	"encoding/json"
	"testing"
)

func TestAccumulatePatch(t *testing.T) {
	for name, testCase := range map[string]struct {
		expected Message
		events   []string
	}{
		"interleaved content blocks (thinking start mid-text)": {
			events: []string{
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
			},
			expected: Message{Content: []ContentBlockUnion{
				{Type: "thinking", Thinking: "Let me think.", Signature: "sig123"},
				{Type: "text", Text: "Hello world!"},
				{Type: "thinking"},
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			message := Message{}
			for _, eventStr := range testCase.events {
				event := MessageStreamEventUnion{}
				err := (&event).UnmarshalJSON([]byte(eventStr))
				if err != nil {
					t.Fatal(err)
				}
				(&message).Accumulate(event)
			}
			marshaledMessage, err := json.Marshal(message)
			if err != nil {
				t.Fatal(err)
			}
			marshaledExpectedMessage, err := json.Marshal(testCase.expected)
			if err != nil {
				t.Fatal(err)
			}
			if string(marshaledMessage) != string(marshaledExpectedMessage) {
				t.Fatalf("Mismatched message: expected %s but got %s", marshaledExpectedMessage, marshaledMessage)
			}
		})
	}
}
