package connector_test

import (
	"errors"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/connector"
)

func TestNormalizeAndProjectConnector(t *testing.T) {
	spec, err := connector.Normalize(connector.Spec{ID: "Public-Web", Protocol: "MCP", Transport: "stdio"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.ID != "public-web" || spec.Name != "public-web" || spec.Protocol != connector.ProtocolMCP {
		t.Fatalf("spec=%#v", spec)
	}
	asset, err := connector.Project("host:connectors", spec)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Identity.Kind != "connector" || asset.Identity.ID != "public-web" || len(asset.Tags) != 2 {
		t.Fatalf("asset=%#v", asset)
	}
}

func TestNormalizeRejectsRuntimeConfiguration(t *testing.T) {
	_, err := connector.Normalize(connector.Spec{ID: "bad connector", Protocol: connector.ProtocolMCP, Transport: connector.TransportStdio})
	if !errors.Is(err, &connector.Error{Code: connector.ErrorCodeInvalidSpec}) {
		t.Fatalf("err=%v", err)
	}
}
