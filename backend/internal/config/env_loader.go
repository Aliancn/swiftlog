package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var loadEnvOnce sync.Once

func init() {
	loadEnvOnce.Do(func() {
		loadEnvFromFile()
	})
}

// loadEnvFromFile searches for a .env file (walking up from the current
// working directory) and loads variables that are not already set.
func loadEnvFromFile() {
	if os.Getenv("SKIP_DOTENV") == "1" {
		return
	}

	envPath, err := findEnvFile()
	if err != nil || envPath == "" {
		return
	}

	if err := parseEnvFile(envPath); err != nil {
		log.Printf("Warning: failed to parse %s: %v", envPath, err)
		return
	}

	log.Printf("Loaded environment variables from %s", envPath)
}

// findEnvFile walks up the directory tree to locate the first .env file.
func findEnvFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	visited := map[string]struct{}{}
	dir := cwd

	for {
		if _, seen := visited[dir]; seen {
			break
		}
		visited[dir] = struct{}{}

		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil
}

// parseEnvFile reads key=value pairs from the provided .env file.
func parseEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(line[7:])
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		value := trimInlineComment(strings.TrimSpace(parts[1]))
		value = unquote(value)

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// trimInlineComment removes inline comments that appear outside of quoted values.
func trimInlineComment(value string) string {
	var quote rune
	for i, r := range value {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
		case r == '#':
			return strings.TrimSpace(value[:i])
		}
	}
	return value
}

// unquote removes surrounding single or double quotes from a value.
func unquote(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
