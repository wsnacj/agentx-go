package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActSavePDFUsesArtifactResolverForRemoteBackend(t *testing.T) {
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				savePDFResult: BrowserSavePDFResult{
					Backend:    "proxy-save-pdf",
					BrowserApp: "Chromium",
					FinalURL:   "https://93.184.216.34/report",
					Title:      "Report",
				},
			},
			remotePath:    "/remote/node/browser/report.pdf",
			remoteContent: []byte("%PDF-remote-browser-report"),
		},
		capabilities: BrowserCapabilities{
			SavePDF: true,
		},
	}
	prevNow := browserNow
	browserNow = func() time.Time {
		return time.Date(2026, 3, 10, 15, 16, 17, 18000000, time.UTC)
	}
	defer func() {
		browserNow = prevNow
	}()

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            root,
		Backend:         backend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"save_pdf","tab_index":2,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act save_pdf: %v", err)
	}
	if len(backend.resolveReqs) != 1 {
		t.Fatalf("expected one browser artifact resolve request, got %#v", backend.resolveReqs)
	}
	if backend.resolveReqs[0].Kind != "save_pdf" || backend.resolveReqs[0].BackendPath != "/remote/node/browser/report.pdf" {
		t.Fatalf("unexpected browser_act save_pdf resolve request: %#v", backend.resolveReqs[0])
	}
	var payload struct {
		Kind           string                         `json:"kind"`
		Path           string                         `json:"path"`
		FilesTouched   []string                       `json:"files_touched"`
		Bytes          int64                          `json:"bytes"`
		Backend        string                         `json:"backend"`
		FinalURL       string                         `json:"final_url"`
		Status         string                         `json:"status"`
		Force          bool                           `json:"force"`
		ReviewDecision string                         `json:"review_decision"`
		ReviewReady    bool                           `json:"review_ready"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
		Artifacts      []struct {
			Kind string `json:"kind"`
			Path string `json:"path"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "save_pdf" || payload.Path != ".agentx/browser/page-20260310-151617.018.pdf" || payload.Bytes != int64(len(backend.remoteContent)) || payload.Backend != "proxy-save-pdf" || payload.FinalURL != "https://93.184.216.34/report" || payload.Status != "saved" || !payload.Force || payload.ReviewDecision != "save_pdf_review_confirmed" || !payload.ReviewReady {
		t.Fatalf("unexpected browser_act save_pdf payload: %#v", payload)
	}
	if !reflect.DeepEqual(payload.FilesTouched, []string{payload.Path}) {
		t.Fatalf("expected browser_act save_pdf files_touched to mirror artifact path, got %#v", payload.FilesTouched)
	}
	if payload.Summary == nil || payload.Summary.Category != "artifact" || payload.Summary.SummaryCode != "save_pdf_completed" {
		t.Fatalf("unexpected browser_act save_pdf summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "artifact" || payload.Display.SummaryCode != "save_pdf_completed" {
		t.Fatalf("unexpected browser_act save_pdf display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "artifact" || payload.Surface.SummaryCode != "save_pdf_completed" {
		t.Fatalf("unexpected browser_act save_pdf surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "review" || payload.View.Category != "artifact" || payload.View.SummaryCode != "save_pdf_completed" {
		t.Fatalf("unexpected browser_act save_pdf view: %#v", payload.View)
	}
	if len(payload.Artifacts) != 1 || payload.Artifacts[0].Kind != "pdf" || payload.Artifacts[0].Path != payload.Path {
		t.Fatalf("unexpected browser_act save_pdf artifacts payload: %#v", payload.Artifacts)
	}
}

func TestRegisterBrowserTools_ActSavePDFRequiresReview(t *testing.T) {
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				savePDFResult: BrowserSavePDFResult{
					Backend:    "proxy-save-pdf",
					BrowserApp: "Chromium",
					FinalURL:   "https://93.184.216.34/report",
					Title:      "Report PDF",
				},
			},
			remotePath:    "/remote/node/browser/report.pdf",
			remoteContent: []byte("%PDF-1.7 remote-browser-report"),
		},
		capabilities: BrowserCapabilities{
			SavePDF: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            root,
		Backend:         backend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"save_pdf","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act save_pdf review gate: %v", err)
	}
	if len(backend.resolveReqs) != 0 {
		t.Fatalf("expected save_pdf to be blocked before artifact resolution, got %#v", backend.resolveReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Path           string `json:"path"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Force          bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "save_pdf" || !strings.HasPrefix(payload.Path, ".agentx/browser/page-") || payload.Status != "review_required" || payload.ReviewDecision != "save_pdf_review_required" || payload.ReviewReady || payload.Force {
		t.Fatalf("unexpected browser_act save_pdf review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActSaveHTMLUsesArtifactResolverForRemoteBackend(t *testing.T) {
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				saveHTMLResult: BrowserSaveHTMLResult{
					Backend:    "proxy-save-html",
					BrowserApp: "Chromium",
					FinalURL:   "https://93.184.216.34/report",
					Title:      "Report HTML",
				},
			},
			remotePath:    "/remote/node/browser/report.html",
			remoteContent: []byte("<html><body>remote-browser-report</body></html>"),
		},
		capabilities: BrowserCapabilities{
			SaveHTML: true,
		},
	}
	prevNow := browserNow
	browserNow = func() time.Time {
		return time.Date(2026, 3, 10, 15, 16, 17, 19000000, time.UTC)
	}
	defer func() {
		browserNow = prevNow
	}()

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            root,
		Backend:         backend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"save_html","tab_index":2,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act save_html: %v", err)
	}
	if len(backend.resolveReqs) != 1 {
		t.Fatalf("expected one browser artifact resolve request, got %#v", backend.resolveReqs)
	}
	if backend.resolveReqs[0].Kind != "save_html" || backend.resolveReqs[0].BackendPath != "/remote/node/browser/report.html" {
		t.Fatalf("unexpected browser_act save_html resolve request: %#v", backend.resolveReqs[0])
	}
	var payload struct {
		Kind           string                         `json:"kind"`
		Path           string                         `json:"path"`
		FilesTouched   []string                       `json:"files_touched"`
		Bytes          int64                          `json:"bytes"`
		Backend        string                         `json:"backend"`
		FinalURL       string                         `json:"final_url"`
		Status         string                         `json:"status"`
		Force          bool                           `json:"force"`
		ReviewDecision string                         `json:"review_decision"`
		ReviewReady    bool                           `json:"review_ready"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
		Artifacts      []struct {
			Kind string `json:"kind"`
			Path string `json:"path"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "save_html" || payload.Path != ".agentx/browser/page-20260310-151617.019.html" || payload.Bytes != int64(len(backend.remoteContent)) || payload.Backend != "proxy-save-html" || payload.FinalURL != "https://93.184.216.34/report" || payload.Status != "saved" || !payload.Force || payload.ReviewDecision != "save_html_review_confirmed" || !payload.ReviewReady {
		t.Fatalf("unexpected browser_act save_html payload: %#v", payload)
	}
	if !reflect.DeepEqual(payload.FilesTouched, []string{payload.Path}) {
		t.Fatalf("expected browser_act save_html files_touched to mirror artifact path, got %#v", payload.FilesTouched)
	}
	if payload.Summary == nil || payload.Summary.Category != "artifact" || payload.Summary.SummaryCode != "save_html_completed" {
		t.Fatalf("unexpected browser_act save_html summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "artifact" || payload.Display.SummaryCode != "save_html_completed" {
		t.Fatalf("unexpected browser_act save_html display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "artifact" || payload.Surface.SummaryCode != "save_html_completed" {
		t.Fatalf("unexpected browser_act save_html surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "review" || payload.View.Category != "artifact" || payload.View.SummaryCode != "save_html_completed" {
		t.Fatalf("unexpected browser_act save_html view: %#v", payload.View)
	}
	if len(payload.Artifacts) != 1 || payload.Artifacts[0].Kind != "html" || payload.Artifacts[0].Path != payload.Path {
		t.Fatalf("unexpected browser_act save_html artifacts payload: %#v", payload.Artifacts)
	}
}

func TestRegisterBrowserTools_ActSaveHTMLRequiresReview(t *testing.T) {
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				saveHTMLResult: BrowserSaveHTMLResult{
					Backend:    "proxy-save-html",
					BrowserApp: "Chromium",
					FinalURL:   "https://93.184.216.34/report",
					Title:      "Report HTML",
				},
			},
			remotePath:    "/remote/node/browser/report.html",
			remoteContent: []byte("<html><body>remote-browser-report</body></html>"),
		},
		capabilities: BrowserCapabilities{
			SaveHTML: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            root,
		Backend:         backend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"save_html","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act save_html review gate: %v", err)
	}
	if len(backend.resolveReqs) != 0 {
		t.Fatalf("expected save_html to be blocked before artifact resolution, got %#v", backend.resolveReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Path           string `json:"path"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Force          bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "save_html" || !strings.HasPrefix(payload.Path, ".agentx/browser/page-") || payload.Status != "review_required" || payload.ReviewDecision != "save_html_review_required" || payload.ReviewReady || payload.Force {
		t.Fatalf("unexpected browser_act save_html review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActWaitDownloadUsesArtifactResolverForRemoteBackend(t *testing.T) {
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				waitDownloadResult: BrowserWaitDownloadResult{
					Backend:     "proxy-wait-download",
					BrowserApp:  "Chromium",
					FinalURL:    "https://93.184.216.34/archive.zip",
					Title:       "archive.zip",
					ContentType: "application/zip",
				},
			},
			remotePath:    "/remote/node/browser/archive.zip",
			remoteContent: []byte("PK\x03\x04remote-browser-wait-download"),
		},
		capabilities: BrowserCapabilities{
			WaitDownload: true,
		},
	}
	prevNow := browserNow
	browserNow = func() time.Time {
		return time.Date(2026, 3, 10, 16, 17, 18, 20000000, time.UTC)
	}
	defer func() {
		browserNow = prevNow
	}()

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            root,
		Backend:         backend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"wait_download","tab_index":2,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act wait_download: %v", err)
	}
	if len(backend.resolveReqs) != 1 {
		t.Fatalf("expected one browser artifact resolve request, got %#v", backend.resolveReqs)
	}
	if backend.resolveReqs[0].Kind != "wait_download" || backend.resolveReqs[0].BackendPath != "/remote/node/browser/archive.zip" {
		t.Fatalf("unexpected browser_act wait_download resolve request: %#v", backend.resolveReqs[0])
	}
	var payload struct {
		Kind           string                         `json:"kind"`
		Path           string                         `json:"path"`
		FilesTouched   []string                       `json:"files_touched"`
		Bytes          int64                          `json:"bytes"`
		Backend        string                         `json:"backend"`
		FinalURL       string                         `json:"final_url"`
		ContentType    string                         `json:"content_type"`
		Status         string                         `json:"status"`
		Force          bool                           `json:"force"`
		ReviewDecision string                         `json:"review_decision"`
		ReviewReady    bool                           `json:"review_ready"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
		Artifacts      []struct {
			Kind string `json:"kind"`
			Path string `json:"path"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "wait_download" || payload.Path != ".agentx/browser/download-20260310-161718.020.zip" || payload.Bytes != int64(len(backend.remoteContent)) || payload.Backend != "proxy-wait-download" || payload.FinalURL != "https://93.184.216.34/archive.zip" || payload.ContentType != "application/zip" || payload.Status != "downloaded" || !payload.Force || payload.ReviewDecision != "wait_download_review_confirmed" || !payload.ReviewReady {
		t.Fatalf("unexpected browser_act wait_download payload: %#v", payload)
	}
	if !reflect.DeepEqual(payload.FilesTouched, []string{payload.Path}) {
		t.Fatalf("expected browser_act wait_download files_touched to mirror artifact path, got %#v", payload.FilesTouched)
	}
	if payload.Summary == nil || payload.Summary.Category != "artifact" || payload.Summary.SummaryCode != "wait_download_completed" {
		t.Fatalf("unexpected browser_act wait_download summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "artifact" || payload.Display.SummaryCode != "wait_download_completed" {
		t.Fatalf("unexpected browser_act wait_download display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "artifact" || payload.Surface.SummaryCode != "wait_download_completed" {
		t.Fatalf("unexpected browser_act wait_download surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "review" || payload.View.Category != "artifact" || payload.View.SummaryCode != "wait_download_completed" {
		t.Fatalf("unexpected browser_act wait_download view: %#v", payload.View)
	}
	if len(payload.Artifacts) != 1 || payload.Artifacts[0].Kind != "download" || payload.Artifacts[0].Path != payload.Path {
		t.Fatalf("unexpected browser_act wait_download artifacts payload: %#v", payload.Artifacts)
	}
}

func TestRegisterBrowserTools_ActWaitDownloadRequiresReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityRemoteArtifactBrowserBackend{
		remoteArtifactBrowserBackend: &remoteArtifactBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			remotePath:         "/remote/node/browser/archive.zip",
			remoteContent:      []byte("PK\x03\x04remote-browser-wait-download"),
		},
		capabilities: BrowserCapabilities{
			WaitDownload: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"wait_download","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act wait_download review gate: %v", err)
	}
	if len(backend.waitDownloadReqs) != 0 || len(backend.resolveReqs) != 0 {
		t.Fatalf("expected wait_download to be blocked before backend work, got reqs=%#v resolve=%#v", backend.waitDownloadReqs, backend.resolveReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Force          bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "wait_download" || payload.Status != "review_required" || payload.ReviewDecision != "wait_download_review_required" || payload.ReviewReady || payload.Force {
		t.Fatalf("unexpected browser_act wait_download review payload: %#v", payload)
	}
}
