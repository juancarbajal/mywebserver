package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const liveReloadScript = `
<!-- Live Reload Script -->
<script>
(function() {
  if (!window.EventSource) return;
  var es = new EventSource('/__livereload');
  es.onmessage = function(e) {
    if (e.data === 'reload') {
      console.log('[mywebserver] File change detected, reloading page...');
      window.location.reload();
    }
  };
})();
</script>
`

// shouldIgnoreFile returns true if the file or path should be ignored by the watcher.
func shouldIgnoreFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return true
	}
	if strings.HasPrefix(base, "#") || strings.HasSuffix(base, "#") || strings.HasSuffix(base, "~") {
		return true
	}
	if strings.HasSuffix(base, ".tmp") || strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".swo") {
		return true
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

// injectLiveReload inserts the live reload script into HTML content.
func injectLiveReload(html []byte) []byte {
	lower := bytes.ToLower(html)
	script := []byte(liveReloadScript)

	// Try finding </body>
	if idx := bytes.LastIndex(lower, []byte("</body>")); idx != -1 {
		var res bytes.Buffer
		res.Grow(len(html) + len(script))
		res.Write(html[:idx])
		res.Write(script)
		res.Write(html[idx:])
		return res.Bytes()
	}

	// Try finding </html>
	if idx := bytes.LastIndex(lower, []byte("</html>")); idx != -1 {
		var res bytes.Buffer
		res.Grow(len(html) + len(script))
		res.Write(html[:idx])
		res.Write(script)
		res.Write(html[idx:])
		return res.Bytes()
	}

	// Fallback: append at the end
	var res bytes.Buffer
	res.Grow(len(html) + len(script))
	res.Write(html)
	res.Write(script)
	return res.Bytes()
}

// Reloader monitors a directory for changes and notifies SSE clients.
type Reloader struct {
	dir     string
	watcher *fsnotify.Watcher
	clients map[chan string]struct{}
	mu      sync.Mutex
	done    chan struct{}
}

// NewReloader creates a new Reloader for the given directory.
func NewReloader(dir string) (*Reloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	r := &Reloader{
		dir:     dir,
		watcher: watcher,
		clients: make(map[chan string]struct{}),
		done:    make(chan struct{}),
	}

	return r, nil
}

// Start begins watching the directory tree and listening for file changes.
func (r *Reloader) Start() error {
	err := filepath.WalkDir(r.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldIgnoreFile(path) && path != r.dir {
				return filepath.SkipDir
			}
			return r.watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	go r.watchLoop()
	return nil
}

// Close stops the watcher and cleans up resources.
func (r *Reloader) Close() error {
	select {
	case <-r.done:
		return nil
	default:
		close(r.done)
	}
	return r.watcher.Close()
}

// Broadcast sends a message to all connected SSE clients.
func (r *Reloader) Broadcast(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for ch := range r.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (r *Reloader) watchLoop() {
	var (
		debounceTimer   *time.Timer
		debounceMu      sync.Mutex
		lastChangedFile string
	)

	for {
		select {
		case <-r.done:
			return

		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// Ignore chmod-only events
			if event.Has(fsnotify.Chmod) && !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
				continue
			}

			if shouldIgnoreFile(event.Name) {
				continue
			}

			// If a new directory was created, watch it recursively
			if event.Has(fsnotify.Create) {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					_ = filepath.WalkDir(event.Name, func(p string, d os.DirEntry, err error) error {
						if err == nil && d.IsDir() {
							if shouldIgnoreFile(p) {
								return filepath.SkipDir
							}
							_ = r.watcher.Add(p)
						}
						return nil
					})
				}
			}

			debounceMu.Lock()
			lastChangedFile = event.Name
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
				debounceMu.Lock()
				file := lastChangedFile
				debounceMu.Unlock()

				rel, err := filepath.Rel(r.dir, file)
				if err != nil {
					rel = file
				}
				fmt.Printf("[reload] File changed: %s -> reloading browser...\n", rel)
				r.Broadcast("reload")
			})
			debounceMu.Unlock()

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[reload] Watcher error: %v\n", err)
		}
	}
}

// serveSSE handles Server-Sent Events requests from the browser.
func (r *Reloader) serveSSE(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan string, 16)
	r.mu.Lock()
	r.clients[ch] = struct{}{}
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.clients, ch)
		r.mu.Unlock()
	}()

	// Send initial connection confirmation
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	notify := req.Context().Done()
	for {
		select {
		case <-notify:
			return
		case <-r.done:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

type reloadResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
	isHTML        bool
	buf           bytes.Buffer
}

func (rw *reloadResponseWriter) WriteHeader(code int) {
	if rw.headerWritten {
		return
	}
	rw.statusCode = code
	ctype := rw.Header().Get("Content-Type")
	if strings.Contains(ctype, "text/html") && code == http.StatusOK {
		rw.isHTML = true
		return
	}
	rw.headerWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *reloadResponseWriter) Write(b []byte) (int, error) {
	if !rw.headerWritten && !rw.isHTML {
		ctype := rw.Header().Get("Content-Type")
		if ctype == "" {
			ctype = http.DetectContentType(b)
			rw.Header().Set("Content-Type", ctype)
		}
		if strings.Contains(ctype, "text/html") && (rw.statusCode == 0 || rw.statusCode == http.StatusOK) {
			rw.isHTML = true
			if rw.statusCode == 0 {
				rw.statusCode = http.StatusOK
			}
		} else {
			rw.headerWritten = true
			if rw.statusCode == 0 {
				rw.statusCode = http.StatusOK
			}
			rw.ResponseWriter.WriteHeader(rw.statusCode)
		}
	}

	if rw.isHTML {
		return rw.buf.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *reloadResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Handler wraps the next handler to inject reload script into HTML responses and serve SSE.
func (r *Reloader) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/__livereload" {
			r.serveSSE(w, req)
			return
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if req.Method == http.MethodHead {
			next.ServeHTTP(w, req)
			return
		}

		rw := &reloadResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, req)

		if rw.isHTML {
			body := rw.buf.Bytes()
			injected := injectLiveReload(body)
			rw.Header().Del("ETag")
			rw.Header().Del("Accept-Ranges")
			rw.Header().Del("Content-Length")
			rw.Header().Set("Content-Length", strconv.Itoa(len(injected)))
			if rw.statusCode == 0 {
				rw.statusCode = http.StatusOK
			}
			rw.headerWritten = true
			rw.ResponseWriter.WriteHeader(rw.statusCode)
			_, _ = rw.ResponseWriter.Write(injected)
		}
	})
}
