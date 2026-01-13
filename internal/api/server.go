package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/haedalwang/kubescout/internal/k8s"
	"github.com/haedalwang/kubescout/internal/model"
	"github.com/haedalwang/kubescout/internal/ui"
	"github.com/haedalwang/kubescout/internal/upstream"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/Masterminds/semver/v3"
)

type Server struct {
	helmClient *k8s.HelmClient
	ahClient   *upstream.ArtifactHubClient
	echo       *echo.Echo
}

// getLogLevel 환경변수에서 로그 레벨 결정
func getLogLevel() middleware.LoggerConfig {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	
	config := middleware.DefaultLoggerConfig
	
	switch level {
	case "DEBUG":
		// 모든 로그 출력
		config.Skipper = func(c echo.Context) bool { return false }
	case "WARN", "ERROR":
		// 4xx, 5xx 에러만 출력
		config.Skipper = func(c echo.Context) bool {
			return c.Response().Status < 400
		}
	case "INFO":
		fallthrough
	default:
		// 기본: health 체크 제외
		config.Skipper = func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api/health")
		}
	}
	
	return config
}

func NewServer(helmClient *k8s.HelmClient, ahClient *upstream.ArtifactHubClient) *Server {
	e := echo.New()
	e.HideBanner = true
	
	// Middleware (로그 레벨 설정 적용)
	e.Use(middleware.LoggerWithConfig(getLogLevel()))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS()) // Enable CORS for dev (Frontend on different port)

	s := &Server{
		helmClient: helmClient,
		ahClient:   ahClient,
		echo:       e,
	}

	// Routes
	s.setupRoutes()
	
	return s
}

func (s *Server) setupRoutes() {
	// Health Check (Kubernetes Probe용)
	s.echo.GET("/api/health", s.handleHealthCheck)
	
	// API Group
	api := s.echo.Group("/api/v1")
	api.GET("/releases", s.handleGetReleases)

	// Static Files (Frontend)
	assets, err := ui.GetFileSystem()
	if err != nil {
		log.Fatalf("Failed to load embedded UI: %v", err)
	}
	fileServer := http.FileServer(http.FS(assets))
	
	// Serve index.html for root and unknown routes (SPA support)
	s.echo.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path
		
		// If API path, let it fall through (Echo handles this usually, but wildcard matches everything)
		if strings.HasPrefix(path, "/api") {
			return echo.ErrNotFound
		}

		// Check if file exists in assets
		f, err := assets.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			// Serve static file
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		// Fallback to index.html for SPA routing
		f, err = assets.Open("index.html")
		if err == nil {
			defer f.Close()
			return c.Stream(http.StatusOK, "text/html", f)
		}
		
		return echo.ErrNotFound
	})
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

// handleHealthCheck Kubernetes liveness/readiness probe용 헬스 체크
func (s *Server) handleHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleGetReleases(c echo.Context) error {
	// 1. List Releases
	releases, err := s.helmClient.ListReleases()
	if err != nil {
		log.Printf("Failed to list releases: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 2. Process concurrently (Logic moved from main.go)
	// TODO: Add caching layer here to avoid hitting API on every refresh
	var wg sync.WaitGroup
	results := make([]model.ComparisonResult, len(releases))

	for i, r := range releases {
		wg.Add(1)
		go func(idx int, rel model.Release) {
			defer wg.Done()
			
			res := model.ComparisonResult{
				Release: rel,
				Status:  model.Unknown,
				CheckedAt: time.Now(),
			}

			// Fetch Upstream
			latest, err := s.ahClient.GetLatestVersion(rel.ChartName)
			if err != nil {
				// Log error but continue
				log.Printf("Failed to check upstream for %s: %v", rel.ChartName, err)
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

	return c.JSON(http.StatusOK, results)
}

// Helper: compareVersions (Duplicate logic from main.go - will consolidate later)
func compareVersions(current, latest string) model.DriftStatus {
	// Clean simple "v" prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

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
		if vLatest.Minor() > vCurrent.Minor() {
			return model.MinorDrift
		}
		return model.PatchDrift
	}

	return model.Sync
}
