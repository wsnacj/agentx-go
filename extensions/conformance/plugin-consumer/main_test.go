package main

import (
	"errors"
	"testing"

	plugincontract "github.com/wsnacj/agentx-go/extensions/plugin"
)

func TestFixedVersionPluginContract(t *testing.T) {
	manifest, err := plugincontract.Parse([]byte(`{"name":"sample","entrypoints":{"skills":"skills"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "sample" || manifest.SchemaVersion != plugincontract.SchemaVersionV1 {
		t.Fatalf("manifest=%#v", manifest)
	}
	_, err = plugincontract.Parse([]byte(`{"name":"sample","approval":"always"}`))
	if !errors.Is(err, &plugincontract.Error{Code: plugincontract.ErrorCodeForbiddenField}) {
		t.Fatalf("err=%v", err)
	}
}
