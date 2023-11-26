package net_fixture

import (
	"net"
	"net/http"
)

type HTTPServerFixture struct {
	handler http.Handler
}
type TearDownHTTPServer = func()

func NewHTTPServerFixture() *HTTPServerFixture {
	var handler http.Handler
	handler = &handlerImpl{}
	return &HTTPServerFixture{
		handler: handler,
	}
}

func (s *HTTPServerFixture) Setup() (port int, tearDown TearDownHTTPServer) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port = listener.Addr().(*net.TCPAddr).Port
	go func() {
		http.Serve(listener, s.handler)
	}()
	return port, func() {
		listener.Close()
	}
}

func (s *HTTPServerFixture) SetHandler(handler http.Handler) {
	s.handler = handler
}

type handlerImpl struct{}

func (h *handlerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"result": "ok"}`))
}
