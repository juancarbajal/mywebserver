package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func extractLastPart(s string, sep string) string {
	idx := strings.LastIndex(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func createLogRecord(r *http.Request) *HTTPReqInfo {
	var ri HTTPReqInfo
	if r.Method != "" {
		ri.method = r.Method
	}
	if r.URL.String() != "" {
		ri.uri = extractLastPart(r.URL.String(), "&")
	}
	if r.Header.Get("Referer") != "" {
		ri.referer = r.Header.Get("Referer")
	}
	if r.Header.Get("User-Agent") != "" {
		ri.userAgent = r.Header.Get("User-Agent")
	}
	if r.Response != nil {
		if r.Response.StatusCode != 0 {
			ri.code, _ = strconv.Atoi(r.Response.Status)
		}
		if r.Response.ContentLength != 0 {
			ri.code = int(r.Response.Request.ContentLength)
		}
	}
	if r.RemoteAddr != "" {
		// ri.ipaddr = ipAddrFromRemoteAddr(r.RemoteAddr)
		ri.ipaddr = extractLastPart(r.RemoteAddr, ":")
	}
	return &ri
}

// writeLog ...
func writeLog(ri *HTTPReqInfo) error {
	fmt.Printf("%s %s %d %d %s %s %s\n",
		ri.method,
		ri.uri,
		ri.code,
		ri.size,
		ri.ipaddr,
		ri.referer,
		ri.userAgent,
	)
	fmt.Println(ri)
	return nil

}
