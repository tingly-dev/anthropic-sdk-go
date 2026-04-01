package anthropic

import (
	"reflect"

	"github.com/anthropics/anthropic-sdk-go/internal/apijson"
	"github.com/tidwall/gjson"
)

func init() {
	// Register custom decoder for []ContentBlockParamUnion to handle string content
	apijson.RegisterCustomDecoder[[]BetaContentBlockParamUnion](func(node gjson.Result, value reflect.Value, defaultDecoder func(gjson.Result, reflect.Value) error) error {
		// If it's a string, convert it to a TextBlock automatically
		if node.Type == gjson.String {
			textBlock := BetaTextBlockParam{
				Text: node.String(),
				Type: "text",
			}
			contentUnion := BetaContentBlockParamUnion{
				OfText: &textBlock,
			}
			arrayValue := reflect.MakeSlice(value.Type(), 1, 1)
			arrayValue.Index(0).Set(reflect.ValueOf(contentUnion))
			value.Set(arrayValue)
			return nil
		}

		return defaultDecoder(node, value)
	})

	// Register custom decoder for []BetaToolResultBlockParamContentUnion to handle string content
	apijson.RegisterCustomDecoder[[]BetaToolResultBlockParamContentUnion](func(node gjson.Result, value reflect.Value, defaultDecoder func(gjson.Result, reflect.Value) error) error {
		// If it's a string, convert it to a TextBlock automatically
		if node.Type == gjson.String {
			textBlock := BetaTextBlockParam{
				Text: node.String(),
				Type: "text",
			}
			contentUnion := BetaToolResultBlockParamContentUnion{
				OfText: &textBlock,
			}
			arrayValue := reflect.MakeSlice(value.Type(), 1, 1)
			arrayValue.Index(0).Set(reflect.ValueOf(contentUnion))
			value.Set(arrayValue)
			return nil
		}

		// If it's already an array, use the default decoder
		return defaultDecoder(node, value)
	})
}
