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

func TestRemoveContainerLockedClearsTunnelState(t *testing.T) {
	svc := newPodmanService()
	svc.containers = []podmanContainer{
		{ID: "abc123", Name: "ws-one"},
	}
	svc.tunnelStateByContainerID["abc123"] = podmanTunnelState{
		Status: tunnelStatusBlocked,
		Code:   "ABCD-EFGH",
	}

	removed := svc.removeContainerLocked(podmanEvent{
		ID:   "abc123",
		Name: "ws-one",
	})
	if !removed {
		t.Fatal("expected container removal")
	}

	if len(svc.containers) != 0 {
		t.Fatal("expected container list to be empty")
	}
	if _, exists := svc.tunnelStateByContainerID["abc123"]; exists {
		t.Fatal("expected tunnel state cleanup for removed container")
	}
}
