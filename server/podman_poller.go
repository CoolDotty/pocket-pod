package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"
)

func (s *podmanService) runPoller(ctx context.Context) {
	s.poll()

	var timer *time.Timer
	var timerCh <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case delay := <-s.pollCh:
			if delay <= 0 {
				delay = podmanPollDebounce
			}
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
			timer = time.NewTimer(delay)
			timerCh = timer.C
		case <-timerCh:
			s.poll()
			timer = nil
			timerCh = nil
		}
	}
}

func (s *podmanService) runEventListener(ctx context.Context) {
	for {
		if err := s.streamEvents(ctx); err != nil {
			if errors.Is(err, errPodmanUnavailable) {
				s.mu.Lock()
				changed := s.errMessage != podmanUnavailableMessage
				s.errMessage = podmanUnavailableMessage
				s.initialized = true
				s.mu.Unlock()
				if changed {
					s.broadcast(podmanStreamMessage{
						Type:    "error",
						Data:    []podmanContainer{},
						Message: podmanUnavailableMessage,
					})
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(podmanEventRestartDelay):
		}
	}
}

func (s *podmanService) streamEvents(ctx context.Context) error {
	if _, err := exec.LookPath("podman"); err != nil {
		return errPodmanUnavailable
	}

	cmd := exec.CommandContext(ctx, "podman", "events", "--format", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if s.applyEvent(line) {
			continue
		}

		s.schedulePoll(podmanPollDebounce)
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return err
	}

	return cmd.Wait()
}

func (s *podmanService) poll() {
	containers, err := listPodmanContainers()
	if err != nil {
		message := podmanLoadFailedMessage
		if errors.Is(err, errPodmanUnavailable) {
			message = podmanUnavailableMessage
		}
		s.mu.Lock()
		changed := s.errMessage != message
		s.errMessage = message
		s.initialized = true
		s.mu.Unlock()
		if changed {
			s.broadcast(podmanStreamMessage{
				Type:    "error",
				Data:    []podmanContainer{},
				Message: message,
			})
		}
		return
	}

	normalizeContainers(containers)
	discoveredTunnelStates := discoverTunnelStatesForContainers(containers)

	s.mu.Lock()
	mergeTunnelStateMap(s.tunnelStateByContainerID, discoveredTunnelStates)
	pruneTunnelStateMap(s.tunnelStateByContainerID, containers)
	enrichContainersWithTunnelState(containers, s.tunnelStateByContainerID)
	hash := hashContainers(containers)
	changed := s.hash != hash || s.errMessage != ""
	stored := make([]podmanContainer, len(containers))
	copy(stored, containers)
	s.containers = stored
	s.hash = hash
	s.errMessage = ""
	s.initialized = true
	s.mu.Unlock()

	if changed {
		s.broadcast(podmanStreamMessage{
			Type: "containers",
			Data: containers,
		})
	}
}

func (s *podmanService) schedulePoll(delay time.Duration) {
	select {
	case s.pollCh <- delay:
	default:
	}
}

func (s *podmanService) applyEvent(raw string) bool {
	var event podmanEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return false
	}

	if strings.ToLower(event.Type) != "container" {
		return false
	}

	rawStatus := strings.TrimSpace(event.Status)
	if rawStatus == "" {
		return false
	}
	status := strings.ToLower(rawStatus)
	if !isPodmanStreamEventAllowed(status) {
		return false
	}

	isRemoval := status == "remove"
	shouldUpdateStatus := shouldUpdateContainerStatusFromEvent(status)

	s.mu.Lock()
	if !s.initialized || s.errMessage != "" {
		s.mu.Unlock()
		s.schedulePoll(podmanPollDebounce)
		return false
	}

	var changed bool
	if isRemoval {
		changed = s.removeContainerLocked(event)
	} else {
		changed = s.upsertContainerLocked(event, rawStatus, shouldUpdateStatus)
	}

	if !changed {
		s.mu.Unlock()
		return false
	}

	normalizeContainers(s.containers)
	enrichContainersWithTunnelState(s.containers, s.tunnelStateByContainerID)
	s.hash = hashContainers(s.containers)
	result := make([]podmanContainer, len(s.containers))
	copy(result, s.containers)
	s.mu.Unlock()

	s.broadcast(podmanStreamMessage{
		Type: "containers",
		Data: result,
	})

	if isRemoval {
		s.schedulePoll(podmanRemoveDebounce)
	} else {
		s.schedulePoll(podmanPollDebounce)
	}
	return true
}

func (s *podmanService) removeContainerLocked(event podmanEvent) bool {
	for i := len(s.containers) - 1; i >= 0; i-- {
		if matchesPodmanEvent(s.containers[i], event) {
			removedID := strings.TrimSpace(s.containers[i].ID)
			for key := range s.tunnelStateByContainerID {
				if isContainerIDMatch(key, removedID) {
					delete(s.tunnelStateByContainerID, key)
				}
			}
			s.containers = append(s.containers[:i], s.containers[i+1:]...)
			return true
		}
	}
	return false
}

func (s *podmanService) upsertContainerLocked(event podmanEvent, status string, shouldUpdateStatus bool) bool {
	for i := range s.containers {
		if matchesPodmanEvent(s.containers[i], event) {
			changed := false
			if event.Name != "" && s.containers[i].Name != event.Name {
				s.containers[i].Name = event.Name
				changed = true
			}
			if event.Image != "" && s.containers[i].Image != event.Image {
				s.containers[i].Image = event.Image
				changed = true
			}
			if event.ID != "" && s.containers[i].ID != event.ID {
				s.containers[i].ID = event.ID
				changed = true
			}
			if shouldUpdateStatus && s.containers[i].Status != status {
				s.containers[i].Status = status
				changed = true
			}
			return changed
		}
	}

	s.containers = append(s.containers, podmanContainer{
		ID:     event.ID,
		Name:   event.Name,
		Image:  event.Image,
		Status: status,
	})
	return true
}

func matchesPodmanEvent(container podmanContainer, event podmanEvent) bool {
	if event.ID != "" {
		left := strings.ToLower(strings.TrimSpace(container.ID))
		right := strings.ToLower(strings.TrimSpace(event.ID))
		if left != "" && right != "" {
			if left == right || strings.HasPrefix(left, right) || strings.HasPrefix(right, left) {
				return true
			}
		}
	}

	if event.Name != "" {
		left := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(container.Name), "/"))
		right := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(event.Name), "/"))
		if left != "" && right != "" && left == right {
			return true
		}
	}

	return false
}

func isPodmanStreamEventAllowed(status string) bool {
	_, ok := podmanStreamEventAllowlist[status]
	return ok
}

func shouldUpdateContainerStatusFromEvent(status string) bool {
	return status != "rename" && status != "update"
}
