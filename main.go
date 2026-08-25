package main

import (
	"flag"
	"fmt"
	"github.com/felixge/httpsnoop"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type HTTPReqInfo struct {
	method    string
	uri       string
	referer   string
	ipaddr    string
	code      int
	size      int64
	duration  time.Duration
	userAgent string
}

func textLoad(dir string, port string) string {
	return fmt.Sprintf(`
	Starting file server...
	Serving directory: %s
	Server running on port %s
	Access from local machine: http://localhost:%s
	Use Ctrl+C to stop the server
        `, dir, port, port, port)
}

func getAbsoluteDir(dir string) (string, error) {
	newDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("Error getting absolute path:", err)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Directory does not exist:", newDir)
	}
	return newDir, nil
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInfo{
			method:    r.Method,
			uri:       r.URL.String(),
			referer:   r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
		}
		ri.ipaddr = requestGetRemoteAddress(r)
		m := httpsnoop.CaptureMetrics(h, w, r)
		ri.code = m.Code
		//		ri.size = m.
		ri.duration = m.Duration
		logHTTPReq(ri)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func main() {
	var absDir string
	var err error
	var (
		port = flag.String("port", "8000", "Port to serve on")
		dir  = flag.String("dir", ".", "Directory to serve files from")
	)
	flag.Parse()

	if absDir, err = getAbsoluteDir(*dir); err != nil {
		fmt.Println(err)
	}

	fs := http.FileServer(http.Dir(absDir))
	http.Handle("/", fs)

	fmt.Println(textLoad(absDir, *port))

	addr := "0.0.0.0:" + *port
	log.Fatal(http.ListenAndServe(addr, nil))
}
