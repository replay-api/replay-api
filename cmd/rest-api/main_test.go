package main_test

// import (
// 	"context"
// 	"testing"

// 	"golang.org/x/vuln/client" // or osv if you don't want a vulnerability client
// )

// func TestModuleVulnerabilities(t *testing.T) {
// 	c := client.NewClient(nil)
// 	// or use this if you don't want to use a client: vulns, err := osv.GetByModule(context.Background(), modulePath)

// 	modulePath := "github.com/psavelis/team-pro/replay-api" // Adjust if your main module is different

// 	// Fetch vulnerabilities from OSV
// 	vulns, err := c.GetByModule(context.Background(), modulePath)
// 	if err != nil {
// 		t.Logf("Warning: Error fetching vulnerability data: %v", err)
// 		return // or t.Fail() to fail the test on errors
// 	}

// 	// Process vulnerabilities and log any matches
// 	if len(vulns) > 0 {
// 		for _, vuln := range vulns {
// 			mod := vuln.Module
// 			for _, affected := range vuln.Affected {
// 				t.Errorf("Vulnerability found in module %s: %s (ID: %s, Versions: %v)", mod.Path, vuln.Summary, vuln.ID, affected.Ranges)
// 			}
// 		}
// 		t.Fail() // Mark the test as failed after logging vulnerabilities
// 	}
// }
