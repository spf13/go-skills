// Command check enforces the rules in CONTRIBUTING.md against the working tree.
//
// It is hermetic: the manifest schemas are vendored and nothing is fetched from
// the network. Run it from anywhere inside the repository:
//
//	go run ./cmd/check
//	go run ./cmd/check -latest go1.27.3
//
// Exit status is 0 when every rule holds, 1 when one or more are broken, and 2
// when the checker itself cannot run.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/go-skills/internal/check"
)

func main() {
	root := flag.String("root", "", "repository root (default: nearest directory containing .claude-plugin/marketplace.json)")
	latest := flag.String("latest", "", "latest stable Go release, e.g. go1.27.3 (default: $GO_RELEASE_LATEST; empty skips the currency comparison)")
	flag.Parse()

	if *latest == "" {
		*latest = os.Getenv("GO_RELEASE_LATEST")
	}
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "check: unexpected argument %q; pass the Go release as -latest\n", flag.Arg(0))
		os.Exit(2)
	}

	report, err := check.Run(check.Options{Root: *root, Latest: *latest})
	if err != nil {
		fmt.Fprintf(os.Stderr, "check: %v\n", err)
		os.Exit(2)
	}

	if !report.OK() {
		fmt.Fprintf(os.Stderr, "check failed (%d problem(s)):\n", len(report.Issues))
		for _, issue := range report.Issues {
			fmt.Fprintf(os.Stderr, "  %s\n", issue)
		}
		os.Exit(1)
	}

	summary := fmt.Sprintf("check passed: %d skill(s), version %s, %d complete Go sample(s) built", len(report.Skills), report.Version, report.CompiledSamples)
	if report.SkippedSamples > 0 {
		summary += fmt.Sprintf(" (%d complete sample(s) skipped: they import another module)", report.SkippedSamples)
	}
	if report.Latest != "" {
		summary += fmt.Sprintf(", Go release claims match %s", report.Latest)
	} else {
		summary += ", Go release comparison skipped (no -latest supplied)"
	}
	fmt.Println(summary)
}
