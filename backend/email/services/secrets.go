package services

import (
	"log"
	"os"
	"strings"
)

func ReadSecret(envKey, envKeyFile EnvKey) string {
	if v := envKey.GetValue(); v != "" {
		return v
	}
	if path := envKeyFile.GetValue(); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read secret file %s: %v", path, err)
		}
		return strings.TrimSpace(string(data))
	}
	return ""
}
