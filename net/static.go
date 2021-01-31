package net

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/rakyll/statik/fs"

	_ "github.com/wufe/polo/statik"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	sniffDone bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.sniffDone {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.sniffDone = true
	}
	return w.Writer.Write(b)
}

func (server *HTTPServer) initStaticFileSystem() {
	if !server.isDev {
		fileSystem, err := fs.New()
		if err != nil {
			panic(err)
		}
		server.fileSystem = &fileSystem
	}
}

func (server *HTTPServer) serveStatic(router *httprouter.Router) {
	if !server.isDev {

		fileServer := http.FileServer(*server.fileSystem)

		router.GET(string(ServerRouteStatic), func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
			req.URL.Path = ps.ByName("filepath")
			w.Header().Add("Vary", "Accept-Encoding")
			if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
				fileServer.ServeHTTP(w, req)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			fileServer.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, req)
		})
	}
}
