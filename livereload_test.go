package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInjectLiveReload(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "With body tag",
			input:    "<html><head></head><body><h1>Hello</h1></body></html>",
			contains: "/__livereload",
		},
		{
			name:     "With uppercase body tag",
			input:    "<html><HEAD></HEAD><BODY><p>Test</p></BODY></html>",
			contains: "/__livereload",
		},
		{
			name:     "With html tag only",
			input:    "<html><p>No body</p></html>",
			contains: "/__livereload",
		},
		{
			name:     "Fragment without tags",
			input:    "<div>Just a div</div>",
			contains: "/__livereload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := string(injectLiveReload([]byte(tt.input)))
			if !strings.Contains(out, tt.contains) {
				t.Fatalf("expected output to contain %q, got: %s", tt.contains, out)
			}
			if strings.Contains(tt.input, "</body>") {
				idxScript := strings.Index(out, "/__livereload")
				idxBody := strings.Index(strings.ToLower(out), "</body>")
				if idxScript > idxBody {
					t.Fatalf("expected script before </body>")
				}
			}
		})
	}
}

func TestShouldIgnoreFile(t *testing.T) {
	ignored := []string{
		".git",
		".git/config",
		".DS_Store",
		"sub/.git/HEAD",
		"test/.hidden",
		"index.html.tmp",
		"index.html.swp",
		"index.html.swo",
		"index.html~",
		"#index.html#",
	}

	for _, path := range ignored {
		if !shouldIgnoreFile(path) {
			t.Errorf("expected %s to be ignored", path)
		}
	}

	allowed := []string{
		"index.html",
		"css/style.css",
		"js/app.js",
		"images/logo.png",
		"sub/folder/file.txt",
	}

	for _, path := range allowed {
		if shouldIgnoreFile(path) {
			t.Errorf("expected %s NOT to be ignored", path)
		}
	}
}

func TestLiveReloadSSEAndFileChange(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(testFile, []byte("<h1>Initial</h1>"), 0644); err != nil {
		t.Fatal(err)
	}

	reloader, err := NewReloader(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	defer reloader.Close()

	if err := reloader.Start(); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(reloader.Handler(http.FileServer(http.Dir(tempDir))))
	defer server.Close()

	// 1. Verify HTML injection
	resp, err := http.Get(server.URL + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var bodyBuilder strings.Builder
	for scanner.Scan() {
		bodyBuilder.WriteString(scanner.Text())
	}
	bodyStr := bodyBuilder.String()
	if !strings.Contains(bodyStr, "/__livereload") {
		t.Fatalf("expected HTML response to contain live reload script, got: %s", bodyStr)
	}

	// 2. Connect to SSE stream
	req, err := http.NewRequest(http.MethodGet, server.URL+"/__livereload", nil)
	if err != nil {
		t.Fatal(err)
	}

	sseResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer sseResp.Body.Close()

	if sseResp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected Content-Type text/event-stream, got %s", sseResp.Header.Get("Content-Type"))
	}

	reader := bufio.NewReader(sseResp.Body)

	// Read initial connected message
	initialLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(initialLine, "connected") {
		t.Fatalf("expected initial connected event, got: %s", initialLine)
	}

	// 3. Modify file and check reload event
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("<h1>Updated</h1>"), 0644); err != nil {
		t.Fatal(err)
	}

	receivedReload := make(chan bool, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.Contains(line, "data: reload") {
				receivedReload <- true
				return
			}
		}
	}()

	select {
	case <-receivedReload:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for reload SSE message")
	}
}
