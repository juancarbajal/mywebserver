package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func textLoad(dir string, port string) string {
	return fmt.Sprintf(`
	Starting file server...
	Serving directory: %s
	Server running on port %s
	Access from local machine: http://localhost:%s
	Use Ctrl+C to stop the server
        `, dir, port, port)
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
		writeLog(createLogRecord(r))

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
	http.Handle("/", logRequestHandler(fs))

	fmt.Println(textLoad(absDir, *port))

	addr := "0.0.0.0:" + *port
	log.Fatal(http.ListenAndServe(addr, nil))
}
