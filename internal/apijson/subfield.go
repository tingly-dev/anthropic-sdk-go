package apijson

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go/v2/packages/respjson"
	"reflect"
)

func getSubField(root reflect.Value, index []int, name string) reflect.Value {
	strct := root.FieldByIndex(index[:len(index)-1])
	if !strct.IsValid() {
		panic("couldn't find encapsulating struct for field " + name)
	}
	meta := strct.FieldByName("JSON")
	if !meta.IsValid() {
		return reflect.Value{}
	}
	field := meta.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}
	}
	return field
}

func setMetadataSubField(root reflect.Value, index []int, name string, meta Field) {
	target := getSubField(root, index, name)
	if !target.IsValid() {
		return
	}

	if target.Type() == reflect.TypeOf(meta) {
		target.Set(reflect.ValueOf(meta))
	} else if respMeta := meta.toRespField(); target.Type() == reflect.TypeOf(respMeta) {
		target.Set(reflect.ValueOf(respMeta))
	}
}

func setMetadataExtraFields(root reflect.Value, index []int, name string, metaExtras map[string]Field) {
	// Param structs (embedders of packages/param.APIObject) have no "JSON"
	// metadata struct; getSubField finds nothing for them. Instead, preserve
	// unknown keys in the param metadata extras so they survive a
	// decode -> re-encode round trip, matching the v1 SDK behavior.
	setParamMetadataExtraFields(root, metaExtras)

	target := getSubField(root, index, name)
	if !target.IsValid() {
		return
	}

	if target.Type() == reflect.TypeOf(metaExtras) {
		target.Set(reflect.ValueOf(metaExtras))
		return
	}

	newMap := make(map[string]respjson.Field, len(metaExtras))
	if target.Type() == reflect.TypeOf(newMap) {
		for k, v := range metaExtras {
			newMap[k] = v.toRespField()
		}
		target.Set(reflect.ValueOf(newMap))
	}
}

// setParamMetadataExtraFields stores unknown keys into a param struct's
// embedded metadata (via the promoted SetExtraFields method) so they are
// re-emitted by param.MarshalObject. Structs without param metadata are
// unaffected.
func setParamMetadataExtraFields(root reflect.Value, metaExtras map[string]Field) {
	if len(metaExtras) == 0 || !root.CanAddr() || !root.CanInterface() {
		return
	}
	setter, ok := root.Addr().Interface().(interface{ SetExtraFields(map[string]any) })
	if !ok {
		return
	}
	extras := make(map[string]any, len(metaExtras))
	for k, f := range metaExtras {
		var v any
		if err := json.Unmarshal(json.RawMessage(f.raw), &v); err != nil {
			// Keep the raw payload if it is not valid JSON (e.g. an
			// invalid field) so nothing is silently dropped.
			v = json.RawMessage(f.raw)
		}
		extras[k] = v
	}
	setter.SetExtraFields(extras)
}

func (f Field) toRespField() respjson.Field {
	if f.IsMissing() {
		return respjson.Field{}
	} else if f.IsNull() {
		return respjson.NewField("null")
	} else if f.IsInvalid() {
		return respjson.NewInvalidField(f.raw)
	} else {
		return respjson.NewField(f.raw)
	}
}
