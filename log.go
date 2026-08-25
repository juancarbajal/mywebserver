package main

import (
	"github.com/kjk/common/tree/main/siser"
	// "httplogsiser"
	// "muloghttp"
	"strconv"
	"sync"
	"time"
)

var (
	muLogHTTP sync.Mutex
)

func logHTTPReq(ri *HTTPReqInfo) {
	var rec siser.Record
	rec.Name = "httplog"
	rec.Append("method", ri.method)
	rec.Append("uri", ri.uri)
	if ri.referer != "" {
		rec.Append("referer", ri.referer)
	}
	rec.Append("ipaddr", ri.ipaddr)
	rec.Append("code", strconv.Itoa(ri.code))
	rec.Append("size", strconv.FormatInt(ri.size, 10))
	durMs := ri.duration / time.Millisecond
	rec.Append("duration", strconv.FormatInt(int64(durMs), 10))
	rec.Append("ua", ri.userAgent)

	muLogHTTP.Lock()
	defer muLogHTTP.Unlock()
	_, _ = httpLogSiser.WriteRecord(&rec)
}
