package hostserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wsnacj/agentx-go/runtime/hosthttp/hostserver"
)

func TestExternalConsumerUsesRequestIdentityContract(t *testing.T) {
	config := hostserver.DefaultConfig("127.0.0.1:0")
	server, err := config.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := hostserver.RequestIDFromContext(r.Context()); got != "external-consumer" {
			t.Fatalf("context request identity = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/example", nil)
	request.Header.Set(hostserver.RequestIDHeader, "external-consumer")
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || recorder.Header().Get(hostserver.RequestIDHeader) != "external-consumer" {
		t.Fatalf("response status=%d request-id=%q", recorder.Code, recorder.Header().Get(hostserver.RequestIDHeader))
	}
}
