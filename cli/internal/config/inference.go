package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InferenceResult contains the inferred project name
type InferenceResult struct {
	Project string
	Source  string // Description of where the value came from
}

// InferProject intelligently infers project name
func InferProject(providedProject string) (*InferenceResult, error) {
	result := &InferenceResult{}

	// Load global config
	cfg, _ := Load()

	// Load project-level config
	projectCfg, _ := LoadProjectConfig()

	// Infer Project
	switch {
	case providedProject != "":
		// User explicitly provided project via flag
		result.Project = providedProject
		result.Source = "command-line flag"

	case projectCfg != nil && projectCfg.Project != "":
		// Project-level config file exists
		result.Project = projectCfg.Project
		result.Source = "project config file"

	case cfg != nil && cfg.DefaultProject != "":
		// User has set a default project
		result.Project = cfg.DefaultProject
		result.Source = "default config"

	default:
		// Try to infer from environment
		inferred := inferFromEnvironment()
		if inferred != "" {
			result.Project = inferred
			result.Source = "auto-detected"
		} else if cfg != nil && cfg.LastProject != "" {
			// Use last used project
			result.Project = cfg.LastProject
			result.Source = "last used"
		} else {
			// Final fallback
			result.Project = "default"
			result.Source = "default fallback"
		}
	}

	return result, nil
}

// inferFromEnvironment tries to infer project name from environment
func inferFromEnvironment() string {
	// 1. Try Git repository name
	if gitRepo := getGitRepoName(); gitRepo != "" {
		return gitRepo
	}

	// 2. Try CI/CD environment variables
	if ciProject := getCIProjectName(); ciProject != "" {
		return ciProject
	}

	// 3. Use current directory name
	if dirName := getCurrentDirName(); dirName != "" {
		return dirName
	}

	return ""
}

// getGitRepoName gets the Git repository name
func getGitRepoName() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	repoPath := strings.TrimSpace(string(output))
	if repoPath == "" {
		return ""
	}

	return filepath.Base(repoPath)
}

// getCIProjectName gets project name from CI/CD environment
func getCIProjectName() string {
	// GitHub Actions
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		parts := strings.Split(repo, "/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	// GitLab CI
	if project := os.Getenv("CI_PROJECT_NAME"); project != "" {
		return project
	}

	// Jenkins
	if job := os.Getenv("JOB_NAME"); job != "" {
		return job
	}

	// CircleCI
	if project := os.Getenv("CIRCLE_PROJECT_REPONAME"); project != "" {
		return project
	}

	return ""
}

// getCurrentDirName gets the current directory name
func getCurrentDirName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	return filepath.Base(cwd)
}

// SanitizeGroupName sanitizes a command string for use as group identifier
func SanitizeGroupName(command string) string {
	// Limit length
	if len(command) > 100 {
		command = command[:100]
	}

	// Replace common separators with hyphens
	command = strings.ReplaceAll(command, "/", "-")
	command = strings.ReplaceAll(command, "\\", "-")
	command = strings.ReplaceAll(command, " ", "-")
	command = strings.ReplaceAll(command, "\t", "-")

	// Convert to lowercase
	command = strings.ToLower(command)

	// Remove any characters that aren't alphanumeric, hyphens, or dots
	var result strings.Builder
	for _, r := range command {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			result.WriteRune(r)
		}
	}

	name := result.String()

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// If empty after sanitization, use default
	if name == "" {
		return "default"
	}

	return name
}
