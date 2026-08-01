package envfile

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Format specifies the export format
type Format string

const (
	FormatDotenv Format = "dotenv"
	FormatJSON   Format = "json"
	FormatDocker Format = "docker"
)

// Export formats secrets map into the specified format string
func Export(secrets map[string]string, format Format) (string, error) {
	switch format {
	case FormatDotenv:
		return exportDotenv(secrets), nil
	case FormatJSON:
		return exportJSON(secrets)
	case FormatDocker:
		return exportDocker(secrets), nil
	default:
		return "", fmt.Errorf("unknown export format: %q (supported: dotenv, json, docker)", format)
	}
}

func exportDotenv(secrets map[string]string) string {
	keys := sortedKeys(secrets)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, secrets[k]))
	}
	return sb.String()
}

func exportJSON(secrets map[string]string) (string, error) {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func exportDocker(secrets map[string]string) string {
	keys := sortedKeys(secrets)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("-e %s=%s\n", k, secrets[k]))
	}
	return sb.String()
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
