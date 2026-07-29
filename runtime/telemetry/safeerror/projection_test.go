package safeerror

import (
	"errors"
	"strings"
	"testing"
)

func TestProjectionObservationSurfacesExcludeRawMaterial(t *testing.T) {
	secret := "safeerror-public-secret-sentinel"
	projection := Project(errors.New(secret), " Runtime Error ", "UPSTREAM/FAILED")
	attrs := AppendAttrs(nil, "", projection)
	details := AppendDetails(nil, "", projection)
	observed := Summary(projection)
	for key, value := range attrs {
		observed += key + "=" + value.(string)
	}
	for key, value := range details {
		observed += key + "=" + value
	}
	if projection.Class != "runtime_error" || projection.Code != "upstream_failed" || len(projection.Identity) != 64 {
		t.Fatalf("projection=%#v", projection)
	}
	if strings.Contains(observed, secret) {
		t.Fatalf("safe observation leaked sentinel: %s", observed)
	}
}

func TestWrapPreservesCauseAndSafeIdentity(t *testing.T) {
	cause := errors.New("wrapped secret sentinel")
	err := Wrap(cause, "operation failed")
	if !errors.Is(err, cause) {
		t.Fatalf("cause not preserved: %v", err)
	}
	if err.Error() != "operation failed" || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe wrapper: %v", err)
	}
	if projection := Project(err, "runtime", "operation_failed"); len(projection.Identity) != 64 {
		t.Fatalf("projection=%#v", projection)
	}
}
