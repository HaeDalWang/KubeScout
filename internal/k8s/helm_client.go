package k8s

import (
	"log"
	"os"

	"github.com/haedalwang/kubescout/internal/model"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmClient struct {
	settings *cli.EnvSettings
}

func NewHelmClient() *HelmClient {
	return &HelmClient{
		settings: cli.New(),
	}
}

func (c *HelmClient) ListReleases() ([]model.Release, error) {
	// Request releases from all namespaces
	actionConfig := new(action.Configuration)
	
	// You can pass an empty string for the namespace to list across all namespaces,
	// but the driver initialization needs a specific namespace or empty for all with care.
	// For listing all, we often need to iterate or use a secret driver that covers all depending on permissions.
	// However, the standard way often involves iterating namespaces or using the "all-namespaces" flag logic.
	// For simplicity in this MVP, let's use the current namespace in settings, or loop if needed.
	// Actually, action.List has an AllNamespaces field.
	
	// To list releases across all namespaces, the action configuration must be initialized 
	// with an empty namespace string. This tells the underlying Kubernetes client 
	// (Secrets/ConfigMap driver) to query at the cluster scope or allow all-namespace listing.
	if err := actionConfig.Init(c.settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewList(actionConfig)
	client.AllNamespaces = true   // List across all namespaces
	client.Deployed = true        // Only deployed releases
	client.StateMask = action.ListDeployed

	releases, err := client.Run()
	if err != nil {
		return nil, err
	}

	var results []model.Release
	for _, r := range releases {
		results = append(results, model.Release{
			Name:         r.Name,
			Namespace:    r.Namespace,
			ChartName:    r.Chart.Metadata.Name,
			ChartVersion: r.Chart.Metadata.Version,
			AppVersion:   r.Chart.Metadata.AppVersion,
			Revision:     r.Version,
			Updated:      r.Info.LastDeployed.String(),
			Icon:         r.Chart.Metadata.Icon,
		})
	}

	return results, nil
}
