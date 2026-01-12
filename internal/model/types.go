package model

import "time"

// Release represents a lightweight Helm release info
type Release struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	AppVersion   string `json:"app_version"`
	Revision     int    `json:"revision"`
	Updated      string `json:"updated"`
	Icon         string `json:"icon,omitempty"`
}

// DriftStatus represents the drift severity
type DriftStatus string

const (
	Sync       DriftStatus = "SYNC"
	PatchDrift DriftStatus = "PATCH_DRIFT"
	MinorDrift DriftStatus = "MINOR_DRIFT"
	MajorDrift DriftStatus = "MAJOR_DRIFT"
	Unknown    DriftStatus = "UNKNOWN"
)

// ComparisonResult holds the drift analysis for a release
type ComparisonResult struct {
	Release         Release     `json:"release"`
	LatestVersion   string      `json:"latest_version"`
	LatestAppVersion string     `json:"latest_app_version"`
	Status          DriftStatus `json:"status"`
	UpstreamUrl     string      `json:"upstream_url"`
	CheckedAt       time.Time   `json:"checked_at"`
}
