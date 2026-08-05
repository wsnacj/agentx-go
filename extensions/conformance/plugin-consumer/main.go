package main

import (
	"fmt"

	plugincontract "github.com/wsnacj/agentx-go/extensions/plugin"
)

func main() {
	manifest, err := plugincontract.Parse([]byte(`{
		"name":"research-kit",
		"schema_version":"v1",
		"trust_boundary":"reviewed",
		"entrypoints":{"skills":"skills","tools":"tools"},
		"requested_permissions":[{"capability":"network-read","reason":"public sources"}]
	}`))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s %s permissions=%d\n", manifest.Name, manifest.TrustBoundary, len(manifest.RequestedPermissions))
}
