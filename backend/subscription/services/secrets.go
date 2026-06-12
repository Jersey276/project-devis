package services

import (
	"log"
	"os"
	"strings"
)

// ReadSecret returns the value of envKey, or reads it from the file pointed
// to by envKeyFile — same pattern as DB_PASSWORD / DB_PASSWORD_FILE.
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
