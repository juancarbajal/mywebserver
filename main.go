package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func textLoad(dir string, port string, reload bool) string {
	reloadStatus := "disabled"
	if reload {
		reloadStatus = "enabled"
	}
	return fmt.Sprintf(`
	Starting file server...
	Serving directory: %s
	Server running on port %s
	Access from local machine: http://localhost:%s
	Live reload: %s
	Use Ctrl+C to stop the server
        `, dir, port, port, reloadStatus)
}

func getAbsoluteDir(dir string) (string, error) {
	newDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("Error getting absolute path: %w", err)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Directory does not exist: %s", newDir)
	}
	return newDir, nil
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/__livereload" {
			writeLog(createLogRecord(r))
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func main() {
	var (
		port   = flag.String("port", "8000", "Port to serve on")
		dir    = flag.String("dir", ".", "Directory to serve files from")
		open   = flag.Bool("open", true, "Open default browser on start")
		reload = flag.Bool("reload", true, "Enable live reload on file changes")
	)
	flag.Parse()

	absDir, err := getAbsoluteDir(*dir)
	if err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = http.FileServer(http.Dir(absDir))

	if *reload {
		reloader, err := NewReloader(absDir)
		if err != nil {
			log.Fatalf("Failed to initialize live reloader: %v", err)
		}
		defer reloader.Close()

		if err := reloader.Start(); err != nil {
			log.Printf("Warning: live reload watching failed: %v", err)
		}
		handler = reloader.Handler(handler)
	}

	mux := http.NewServeMux()
	mux.Handle("/", logRequestHandler(handler))

	addr := "0.0.0.0:" + *port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	serverURL := fmt.Sprintf("http://localhost:%s", *port)
	fmt.Println(textLoad(absDir, *port, *reload))

	if *open {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(serverURL); err != nil {
				fmt.Printf("Note: Could not open default browser (%v). Access URL: %s\n", err, serverURL)
			}
		}()
	}

	server := &http.Server{
		Handler: mux,
	}
	log.Fatal(server.Serve(listener))
}
