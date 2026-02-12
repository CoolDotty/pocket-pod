package main

import (
	"encoding/json"
	"hash/fnv"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func listPodmanContainers() ([]podmanContainer, error) {
	if _, err := exec.LookPath("podman"); err != nil {
		return nil, errPodmanUnavailable
	}

	output, err := runPodmanList(true)
	if err != nil {
		output, err = runPodmanList(false)
		if err != nil {
			return nil, err
		}
	}
	if len(output) == 0 {
		return []podmanContainer{}, nil
	}

	return parsePodmanContainers(output)
}

func runPodmanList(includeSize bool) ([]byte, error) {
	args := []string{"ps", "--all", "--format", "json"}
	if includeSize {
		args = append(args, "--size")
	}
	cmd := exec.Command("podman", args...)
	return cmd.Output()
}

func parsePodmanContainers(output []byte) ([]podmanContainer, error) {
	var raw []map[string]any
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, err
	}

	containers := make([]podmanContainer, 0, len(raw))
	for _, item := range raw {
		names := getStringSlice(item, "Names")
		name := ""
		if len(names) > 0 {
			name = names[0]
		}
		if name == "" {
			name = getString(item, "Name")
		}

		createdAt := getString(item, "CreatedAt")
		if createdAt == "" {
			createdAt = getString(item, "Created")
		}

		image := getString(item, "Image")
		if image == "" {
			image = getString(item, "ImageName")
		}

		status := getString(item, "Status")
		if status == "" {
			status = getString(item, "State")
		}

		ports := getPorts(item)
		storageSize := getStorageSize(item)

		containers = append(containers, podmanContainer{
			ID:          getString(item, "Id"),
			Name:        name,
			Image:       image,
			Status:      status,
			StorageSize: storageSize,
			CreatedAt:   createdAt,
			Ports:       ports,
			Labels:      getStringMap(item, "Labels"),
		})
	}

	return containers, nil
}

func normalizeContainers(containers []podmanContainer) {
	sort.SliceStable(containers, func(i, j int) bool {
		left := containers[i].Name
		right := containers[j].Name

		if left == "" {
			left = containers[i].ID
		}
		if right == "" {
			right = containers[j].ID
		}

		if left == right {
			return containers[i].ID < containers[j].ID
		}

		return left < right
	})
}

func hashContainers(containers []podmanContainer) uint64 {
	hasher := fnv.New64a()
	for _, container := range containers {
		writeHashField(hasher, container.ID)
		writeHashField(hasher, container.Name)
		writeHashField(hasher, container.Image)
		writeHashField(hasher, container.Status)
		writeHashField(hasher, container.StorageSize)
		writeHashField(hasher, container.CreatedAt)
		writeHashField(hasher, container.Ports)
		writeHashField(hasher, container.TunnelStatus)
		writeHashField(hasher, container.TunnelCode)
		writeHashField(hasher, container.TunnelMessage)

		if len(container.Labels) > 0 {
			keys := make([]string, 0, len(container.Labels))
			for key := range container.Labels {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				writeHashField(hasher, key)
				writeHashField(hasher, container.Labels[key])
			}
		}

		writeHashField(hasher, "|")
	}

	return hasher.Sum64()
}

func writeHashField(w io.Writer, value string) {
	_, _ = io.WriteString(w, value)
	_, _ = io.WriteString(w, "\x00")
}

func getString(data map[string]any, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return ""
	}
}

func getStringSlice(data map[string]any, key string) []string {
	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, entry := range typed {
			if entry == nil {
				continue
			}
			if str, ok := entry.(string); ok && strings.TrimSpace(str) != "" {
				out = append(out, strings.TrimSpace(str))
			}
		}
		return out
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return nil
	}
}

func getStringMap(data map[string]any, key string) map[string]string {
	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case map[string]string:
		if len(typed) == 0 {
			return nil
		}
		return typed
	case map[string]any:
		out := make(map[string]string, len(typed))
		for k, entry := range typed {
			if entry == nil {
				continue
			}
			if str, ok := entry.(string); ok {
				out[k] = str
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func getPorts(data map[string]any) string {
	value, ok := data["Ports"]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, entry := range typed {
			if entry == nil {
				continue
			}
			if str, ok := entry.(string); ok && strings.TrimSpace(str) != "" {
				out = append(out, strings.TrimSpace(str))
			}
		}
		return strings.Join(out, ", ")
	case []string:
		if len(typed) == 0 {
			return ""
		}
		return strings.Join(typed, ", ")
	default:
		return ""
	}
}

func getStorageSize(data map[string]any) string {
	size := getString(data, "Size")
	if size != "" {
		return size
	}
	size = getString(data, "SizeRw")
	if size != "" {
		return size
	}
	return getString(data, "SizeRootFs")
}
