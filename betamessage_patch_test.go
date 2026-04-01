package anthropic

import (
	"encoding/json"
	"testing"
)

func TestBetaToolResultBlockParamStringContent(t *testing.T) {
	toolResultJSON := `{"type":"tool_result","content":"error message","tool_use_id":"123"}`
	var toolResult BetaToolResultBlockParam
	err := json.Unmarshal([]byte(toolResultJSON), &toolResult)
	if err != nil {
		t.Fatal(err)
	}
	if len(toolResult.Content) != 1 || toolResult.Content[0].OfText.Text != "error message" {
		t.Error("String content not converted to TextBlock")
	}
}
