package mediaartifact

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDescriptorReflectContract(t *testing.T) {
	want := []struct {
		name string
		typ  reflect.Type
		tag  string
	}{
		{name: "Source", typ: reflect.TypeFor[string](), tag: "source"},
		{name: "Kind", typ: reflect.TypeFor[string](), tag: "kind"},
		{name: "Path", typ: reflect.TypeFor[string](), tag: "path,omitempty"},
		{name: "URL", typ: reflect.TypeFor[string](), tag: "url,omitempty"},
		{name: "MIMEType", typ: reflect.TypeFor[string](), tag: "mime_type,omitempty"},
		{name: "Format", typ: reflect.TypeFor[string](), tag: "format,omitempty"},
		{name: "Bytes", typ: reflect.TypeFor[int64](), tag: "bytes,omitempty"},
		{name: "Width", typ: reflect.TypeFor[int](), tag: "width,omitempty"},
		{name: "Height", typ: reflect.TypeFor[int](), tag: "height,omitempty"},
		{name: "DurationMs", typ: reflect.TypeFor[int64](), tag: "duration_ms,omitempty"},
		{name: "TimestampMs", typ: reflect.TypeFor[int64](), tag: "timestamp_ms,omitempty"},
		{name: "FPS", typ: reflect.TypeFor[int64](), tag: "fps,omitempty"},
		{name: "HasAudio", typ: reflect.TypeFor[*bool](), tag: "has_audio,omitempty"},
		{name: "ScreenIndex", typ: reflect.TypeFor[int64](), tag: "screen_index,omitempty"},
		{name: "CaptureScope", typ: reflect.TypeFor[string](), tag: "capture_scope,omitempty"},
		{name: "CaptureWidth", typ: reflect.TypeFor[int](), tag: "capture_width,omitempty"},
		{name: "CaptureHeight", typ: reflect.TypeFor[int](), tag: "capture_height,omitempty"},
		{name: "Facing", typ: reflect.TypeFor[string](), tag: "facing,omitempty"},
		{name: "Index", typ: reflect.TypeFor[int](), tag: "index,omitempty"},
		{name: "CreatedAt", typ: reflect.TypeFor[string](), tag: "created_at,omitempty"},
	}

	typ := reflect.TypeFor[Descriptor]()
	if typ.NumField() != len(want) {
		t.Fatalf("Descriptor field count = %d, want %d", typ.NumField(), len(want))
	}
	for i, fieldContract := range want {
		field := typ.Field(i)
		if field.Name != fieldContract.name ||
			field.Type != fieldContract.typ ||
			field.Tag.Get("json") != fieldContract.tag {
			t.Fatalf(
				"Descriptor field[%d] = %s %s json:%q, want %s %s json:%q",
				i,
				field.Name,
				field.Type,
				field.Tag.Get("json"),
				fieldContract.name,
				fieldContract.typ,
				fieldContract.tag,
			)
		}
	}
}

func TestDescriptorZeroValueJSONContract(t *testing.T) {
	payload, err := json.Marshal(Descriptor{})
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	if got, want := string(payload), `{"source":"","kind":""}`; got != want {
		t.Fatalf("zero-value JSON = %s, want %s", got, want)
	}
}

func TestDescriptorHasAudioTriState(t *testing.T) {
	falseValue := false
	trueValue := true
	cases := []struct {
		name string
		in   Descriptor
		want string
	}{
		{name: "unknown", in: Descriptor{}, want: `{"source":"","kind":""}`},
		{name: "false", in: Descriptor{HasAudio: &falseValue}, want: `{"source":"","kind":"","has_audio":false}`},
		{name: "true", in: Descriptor{HasAudio: &trueValue}, want: `{"source":"","kind":"","has_audio":true}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}
			if string(payload) != tc.want {
				t.Fatalf("JSON = %s, want %s", payload, tc.want)
			}
		})
	}
}
