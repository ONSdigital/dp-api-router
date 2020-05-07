package proxy

import "net/http"

//go:generate moq -out ./mock/reverse_proxy.go -pkg mock . IReverseProxy

// IReverseProxy represens the required methods from httputils.ReverseProxy
type IReverseProxy interface {
	ServeHTTP(rw http.ResponseWriter, req *http.Request)
}
