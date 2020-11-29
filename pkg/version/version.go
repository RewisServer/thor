// Package to contain information like Version, GitCommit, etc.
// Will be filled with information at build time.
package version

import "fmt"

var (
	Version       string
	BuildTime     string
	BuildHost     string
	BuildUser     string
	GitBranch     string
	GitCommit     string
	GitCommitTime string
	GitSummary    string
)

func BuildContext() string {
	return fmt.Sprintf("(commit=%s, user=%s, time=%s)", GitCommit, BuildUser, BuildTime)
}
