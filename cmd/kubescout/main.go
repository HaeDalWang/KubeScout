package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/Masterminds/semver/v3"
	"github.com/haedalwang/kubescout/internal/k8s"
	"github.com/haedalwang/kubescout/internal/model"
	"github.com/haedalwang/kubescout/internal/upstream"
)

func main() {
	log.Println("🔭 KubeScout - Starting Discovery...")

	// 1. Initialize Clients
	helmClient := k8s.NewHelmClient()
	ahClient := upstream.NewArtifactHubClient()

	// 2. List Releases
	log.Println("Listing Helm releases...")
	releases, err := helmClient.ListReleases()
	if err != nil {
		log.Fatalf("Failed to list releases: %v", err)
	}
	log.Printf("Found %d releases.\n", len(releases))

	// 3. Process each release (Concurrent)
	var wg sync.WaitGroup
	results := make([]model.ComparisonResult, len(releases))

	for i, r := range releases {
		wg.Add(1)
		go func(idx int, rel model.Release) {
			defer wg.Done()
			
			res := model.ComparisonResult{
				Release: rel,
				Status:  model.Unknown,
			}

			// Fetch Upstream
			latest, err := ahClient.GetLatestVersion(rel.ChartName)
			if err != nil {
				// log.Printf("Warning: Failed to find upstream for %s: %v", rel.ChartName, err)
			} else {
				res.LatestVersion = latest.LatestVersion
				res.LatestAppVersion = latest.LatestAppVersion
				res.UpstreamUrl = latest.UpstreamUrl
				
				// Compare Versions
				res.Status = compareVersions(rel.ChartVersion, latest.LatestVersion)
			}
			results[idx] = res

		}(i, r)
	}
	wg.Wait()

	// 4. Output Results
	printTable(results)
}

func compareVersions(current, latest string) model.DriftStatus {
	vCurrent, err := semver.NewVersion(current)
	if err != nil {
		return model.Unknown
	}
	vLatest, err := semver.NewVersion(latest)
	if err != nil {
		return model.Unknown
	}

	if vCurrent.Equal(vLatest) {
		return model.Sync
	}

	if vLatest.GreaterThan(vCurrent) {
		if vLatest.Major() > vCurrent.Major() {
			return model.MajorDrift
		}
		return model.MinorDrift
	}

	return model.Sync // Current is newer?
}

func printTable(results []model.ComparisonResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "STATUS\tRELEASE\tCHART\tCURRENT_VER\tLATEST_VER\tAPP_VER")
	
	for _, r := range results {
		statusIcon := "⚪"
		switch r.Status {
		case model.Sync:
			statusIcon = "🟢"
		case model.MinorDrift:
			statusIcon = "🟡"
		case model.MajorDrift:
			statusIcon = "🔴"
		case model.Unknown:
			statusIcon = "⚪"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			statusIcon,
			r.Release.Name,
			r.Release.ChartName,
			r.Release.ChartVersion,
			r.LatestVersion,
			r.Release.AppVersion,
		)
	}
	w.Flush()
}
