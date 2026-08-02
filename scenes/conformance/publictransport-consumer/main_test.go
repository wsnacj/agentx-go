package main

import "testing"

func TestFixedVersionPublicTransportConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-publictransport-ok:public-transport-readonly-pack:public_transport_ticket_lookup_v1:1:true"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
