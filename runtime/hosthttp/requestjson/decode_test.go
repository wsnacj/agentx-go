package requestjson

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeStrictBoundedContract(t *testing.T) {
	type request struct {
		Prompt string `json:"prompt,omitempty"`
	}
	for _, test := range []struct {
		name    string
		body    string
		limit   int64
		want    request
		wantErr error
	}{
		{name: "empty", body: "", limit: 32},
		{name: "whitespace", body: " \n\t", limit: 32},
		{name: "valid", body: `{"prompt":"ok"}`, limit: 32, want: request{Prompt: "ok"}},
		{name: "unknown field", body: `{"unknown":true}`, limit: 32},
		{name: "second value", body: `{ } { }`, limit: 32, wantErr: ErrTrailingData},
		{name: "trailing garbage", body: `{ } garbage`, limit: 32, wantErr: ErrTrailingData},
		{name: "null trailing garbage", body: `null garbage`, limit: 32, wantErr: ErrTrailingData},
		{name: "exact limit", body: `{"prompt":"ok"}`, limit: int64(len(`{"prompt":"ok"}`)), want: request{Prompt: "ok"}},
		{name: "over limit", body: `{"prompt":"ok"}`, limit: int64(len(`{"prompt":"ok"}`) - 1), wantErr: ErrBodyTooLarge},
		{name: "invalid limit", body: `{}`, limit: 0, wantErr: ErrInvalidLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got request
			err := Decode(strings.NewReader(test.body), test.limit, &got)
			if test.name == "unknown field" {
				if err == nil || !strings.Contains(err.Error(), "unknown field") {
					t.Fatalf("Decode() error=%v, want unknown-field rejection", err)
				}
				return
			}
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("Decode() error=%v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() error: %v", err)
			}
			if got != test.want {
				t.Fatalf("Decode()=%#v, want %#v", got, test.want)
			}
		})
	}
}
