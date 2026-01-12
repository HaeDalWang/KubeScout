package upstream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/haedalwang/kubescout/internal/model"
)

const artifactHubSearchAPI = "https://artifacthub.io/api/v1/packages/search"

type ArtifactHubClient struct {
	httpClient *http.Client
}

func NewArtifactHubClient() *ArtifactHubClient {
	return &ArtifactHubClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SearchResponse struct {
	Packages []PackageSummary `json:"packages"`
}

type PackageSummary struct {
	PackageId  string     `json:"package_id"`
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	AppVersion string     `json:"app_version"`
	Repository Repository `json:"repository"`
	Stars      int        `json:"stars"`
	Deprecated bool       `json:"deprecated"`
	Url        string     `json:"url"`
}

type Repository struct {
	Name             string `json:"name"`
	Url              string `json:"url"`
	Official         bool   `json:"official"`
	VerifiedPublisher bool  `json:"verified_publisher"`
}

// Preset Registry for "Zero-Config" accuracy on popular charts.
// This is a partial list of widely used charts to avoid search ambiguity.
var knownPackages = map[string]struct {
	Repo string
	Name string
}{
	"argo-cd":                      {Repo: "argo", Name: "argo-cd"},
	"aws-load-balancer-controller": {Repo: "aws", Name: "aws-load-balancer-controller"},
	"karpenter":                    {Repo: "aws-karpenter", Name: "karpenter"},
	"keda":                         {Repo: "kedacore", Name: "keda"},
	"cert-manager":                 {Repo: "cert-manager", Name: "cert-manager"},
	"ingress-nginx":                {Repo: "ingress-nginx", Name: "ingress-nginx"},
	"prometheus":                   {Repo: "prometheus-community", Name: "prometheus"},
	"kube-prometheus-stack":        {Repo: "prometheus-community", Name: "kube-prometheus-stack"},
	"external-dns":                 {Repo: "external-dns", Name: "external-dns"},
	"n8n":                          {Repo: "community-charts", Name: "n8n"},
}

// GetLatestVersion determines the upstream version using a Hybrid Strategy:
// 1. Preset Check: If the chart is in our 'knownPackages' list, query specific repo directly.
// 2. Search Fallback: If unknown, search Artifact Hub and prioritize Official > Stars.
func (c *ArtifactHubClient) GetLatestVersion(chartName string) (*model.ComparisonResult, error) {
	// 1. Preset Strategy
	if preset, ok := knownPackages[chartName]; ok {
		return c.getPackageDetail(preset.Repo, preset.Name)
	}

	// 2. Fallback Search Strategy
	return c.searchPackage(chartName)
}

func (c *ArtifactHubClient) getPackageDetail(repo, name string) (*model.ComparisonResult, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/helm/%s/%s", repo, name)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("artifact hub detail api error: %d", resp.StatusCode)
	}

	var pkg PackageSummary // The detail API returns similar fields
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, err
	}

	return &model.ComparisonResult{
		LatestVersion:    pkg.Version,
		LatestAppVersion: pkg.AppVersion,
		UpstreamUrl:      fmt.Sprintf("https://artifacthub.io/packages/helm/%s/%s", repo, name),
		CheckedAt:        time.Now(),
	}, nil
}

func (c *ArtifactHubClient) searchPackage(chartName string) (*model.ComparisonResult, error) {
	// Search with higher limit to find the exact match
	url := fmt.Sprintf("%s?ts_query_web=%s&kind=0&limit=20", artifactHubSearchAPI, chartName)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("artifact hub search api error: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	// Filter and Sort Strategy
	var bestMatch *PackageSummary

	for _, pkg := range searchResp.Packages {
		// 1. Exact Name Match Filter
		if pkg.Name != chartName {
			continue
		}

		// 2. Selection Strategy
		if bestMatch == nil {
			val := pkg
			bestMatch = &val
			continue
		}

		// Comparison Logic
		// 1. Official Repo wins
		if pkg.Repository.Official && !bestMatch.Repository.Official {
			val := pkg
			bestMatch = &val
			continue
		}
		if !pkg.Repository.Official && bestMatch.Repository.Official {
			continue
		}

		// 2. Non-Deprecated wins
		if !pkg.Deprecated && bestMatch.Deprecated {
			val := pkg
			bestMatch = &val
			continue
		}
		if pkg.Deprecated && !bestMatch.Deprecated {
			continue
		}

		// 3. Stars wins
		if pkg.Stars > bestMatch.Stars {
			val := pkg
			bestMatch = &val
			continue
		}
		if pkg.Stars < bestMatch.Stars {
			continue
		}

		// 4. Verified Publisher (Tie-breaker)
		if pkg.Repository.VerifiedPublisher && !bestMatch.Repository.VerifiedPublisher {
			val := pkg
			bestMatch = &val
			continue
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("package not found after filtering")
	}
	
	return &model.ComparisonResult{
		LatestVersion:    bestMatch.Version,
		LatestAppVersion: bestMatch.AppVersion,
		UpstreamUrl:      fmt.Sprintf("https://artifacthub.io/packages/helm/%s/%s", bestMatch.Repository.Name, bestMatch.Name),
		CheckedAt:        time.Now(),
	}, nil
}
