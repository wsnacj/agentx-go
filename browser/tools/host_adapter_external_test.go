package tools_test

import (
	"errors"
	"testing"

	browserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	browsertools "github.com/wsnacj/agentx-go/browser/tools"
)

type externalRouteBackend struct{ externalBackend }

func (externalRouteBackend) BrowserRuntimeInfo() browsertools.BrowserRuntimeInfo {
	return browsertools.BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}
}

func (backend externalRouteBackend) ResolveBrowserExecutionRoute(requested browsertools.BrowserRuntimeInfo) (browsertools.BrowserResolvedExecutionRoute, error) {
	info := browsertools.MergeBrowserRuntimeInfo(backend.BrowserRuntimeInfo(), requested)
	return browsertools.BrowserResolvedExecutionRoute{
		Backend:      backend,
		RuntimeInfo:  info,
		Capabilities: browsertools.BrowserCapabilities{Open: true},
	}, nil
}

func (externalRouteBackend) BrowserDoctorRouteMetadata() browsertools.BrowserDoctorRouteMetadata {
	return browsertools.BrowserDoctorRouteMetadata{Source: "fixture", Endpoint: "http://127.0.0.1:43123"}
}

func TestExternalHostAdapterCanImplementRouteAndDoctorPorts(t *testing.T) {
	backend := externalRouteBackend{}
	var _ browserruntime.BrowserBackend = backend
	var _ browsertools.BrowserDoctorRouteMetadataProvider = backend

	route, err := backend.ResolveBrowserExecutionRoute(browsertools.BrowserRuntimeInfo{Profile: "isolated"})
	if err != nil || route.RuntimeInfo.Profile != "isolated" || route.Backend == nil {
		t.Fatalf("route=%#v err=%v", route, err)
	}
	typed := browsertools.NewBrowserManagedRouteUnavailableError("node", "fixture", errors.New("offline"))
	if typed == nil || typed.Error() == "" {
		t.Fatal("expected typed managed-route error")
	}
}
