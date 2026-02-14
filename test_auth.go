package main

import (
	"fmt"
	"strings"
)

func isPublicPostPath(path string) bool {
	publicPostPaths := map[string]bool{
		"/onboarding/steam":  true,
		"/onboarding/google": true,
		"/onboarding/email":  true,
		"/auth/login":        true,
		"/auth/logout":       true,
		"/auth/guest":        true,
		"/webhooks/stripe":   true,
	}
	
	if strings.Contains(path, "/games/") && strings.Contains(path, "/replays") && !strings.Contains(path, "/replays/") {
		return true
	}
	
	return publicPostPaths[path]
}

func main() {
	testPaths := []string{
		"/games/cs2/replays",
		"/games/cs2/replays/123",
		"/auth/guest",
		"/onboarding/steam",
	}
	
	for _, path := range testPaths {
		result := isPublicPostPath(path)
		fmt.Printf("Path: %s -> %t\n", path, result)
	}
}
