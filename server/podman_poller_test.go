package main

import "testing"

func TestIsPodmanStreamEventAllowed(t *testing.T) {
	if !isPodmanStreamEventAllowed("start") {
		t.Fatal("expected start to be allowed")
	}

	if !isPodmanStreamEventAllowed("exited") {
		t.Fatal("expected exited to be allowed")
	}

	if isPodmanStreamEventAllowed("mount") {
		t.Fatal("expected mount to be filtered")
	}

	if isPodmanStreamEventAllowed("started") {
		t.Fatal("expected started to be filtered due to literal matching")
	}
}

func TestShouldUpdateContainerStatusFromEvent(t *testing.T) {
	if shouldUpdateContainerStatusFromEvent("rename") {
		t.Fatal("rename should not update status")
	}
	if shouldUpdateContainerStatusFromEvent("update") {
		t.Fatal("update should not update status")
	}
	if !shouldUpdateContainerStatusFromEvent("start") {
		t.Fatal("start should update status")
	}
}

func TestMapEventStatus(t *testing.T) {
	if got := mapEventStatus("kill"); got != "Exited" {
		t.Fatalf("mapEventStatus(kill) = %q, want Exited", got)
	}
	if got := mapEventStatus("exited"); got != "Exited" {
		t.Fatalf("mapEventStatus(exited) = %q, want Exited", got)
	}
}
