package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

var errPodmanUnavailable = errors.New("podman not available")

type podmanContainer struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Status    string            `json:"status"`
	CreatedAt string            `json:"createdAt,omitempty"`
	Ports     string            `json:"ports,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func registerPodmanRoutes(router *router.Router[*core.RequestEvent]) {
	router.GET("/podman/containers", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Unauthorized.",
			})
		}

		containers, err := listPodmanContainers()
		if err != nil {
			if errors.Is(err, errPodmanUnavailable) {
				return re.JSON(http.StatusServiceUnavailable, map[string]string{
					"message": "Podman is not available on the server.",
				})
			}

			return re.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to load Podman containers.",
			})
		}

		return re.JSON(http.StatusOK, containers)
	})
}

func listPodmanContainers() ([]podmanContainer, error) {
	if _, err := exec.LookPath("podman"); err != nil {
		return nil, errPodmanUnavailable
	}

	cmd := exec.Command("podman", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return []podmanContainer{}, nil
	}

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

		ports := getPorts(item)

		containers = append(containers, podmanContainer{
			ID:        getString(item, "Id"),
			Name:      name,
			Image:     image,
			Status:    getString(item, "Status"),
			CreatedAt: createdAt,
			Ports:     ports,
			Labels:    getStringMap(item, "Labels"),
		})
	}

	return containers, nil
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
