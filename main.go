package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	log "github.com/sirupsen/logrus"
)

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = func(res *http.Response) error {

		if res.Header.Get("Content-Type") == "text/html" {

			res.Header.Add("set-cookie", "CaracalSession=CICCIO")

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
			}

			buffer := bytes.NewBufferString("<div style=\"position:fixed;bottom:0;right:0;padding:30px;z-index:9999;background:white;\">SESSION: TESTSESSION</div>")
			buffer.Write(body)

			res.Body = ioutil.NopCloser(buffer)
			res.Header["Content-Length"] = []string{fmt.Sprint(buffer.Len())}
		}

		return nil
	}

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	log.Println(req.URL.RequestURI())

	proxy.ServeHTTP(res, req)
}

func main() {

	port := flag.String("port", "9999", "Port")
	flag.Parse()

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
			serveReverseProxy("http://localhost:8008", res, req)
		})

		server := &http.Server{
			Addr:    ":" + *port,
			Handler: mux,
		}
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
		wg.Done()
	}()

	log.Print("OK")
	wg.Wait()
}
