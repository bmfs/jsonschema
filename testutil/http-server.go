package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type HTTPServer struct {
	*httptest.Server
	endpoints map[string]string
	callCount map[string]int
}

func CreateHTTPServer() *HTTPServer {
	s := &HTTPServer{
		endpoints: map[string]string{},
		callCount: map[string]int{},
	}

	s.Server = httptest.NewServer(http.HandlerFunc(s.Handler))

	return s
}

func (s *HTTPServer) incCounter(route string) {
	if count, exists := s.callCount[route]; exists {
		s.callCount[route] = count + 1
	} else {
		s.callCount[route] = 1
	}
}
func (s *HTTPServer) Handler(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Path
	if schema, exists := s.endpoints[route]; exists {
		fmt.Fprintln(w, schema)
	} else {
		w.WriteHeader(404)
	}
	s.incCounter(route)
}

func (s *HTTPServer) SetRoute(route, schema string) {
	s.endpoints[route] = schema
}

func (s *HTTPServer) GetCount(route string) int {
	if count, exists := s.callCount[route]; exists {
		return count
	}
	return 0
}
