package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type EndpointRequestBuilder interface {
	BuildRequestFromEndpoint(endpoint *Endpoint) *http.Request
}

type defaultRequestBuilder struct {
}

func NewDefaultRequestBuilder() EndpointRequestBuilder {
	return &defaultRequestBuilder{}
}

func (d defaultRequestBuilder) BuildRequestFromEndpoint(endpoint *Endpoint) *http.Request {
	uri, err := url.Parse(endpoint.Url)
	if err != nil {
		panic(err)
	}

	request := &http.Request{
		Method: strings.ToUpper(endpoint.Method),
		URL:    uri,
		Header: endpoint.Headers,
		Body:   ioutil.NopCloser(bytes.NewReader([]byte(endpoint.Body))),
	}

	// if has auth set it
	if endpoint.Auth != nil {
		if len(endpoint.Auth.Password) == 0 || len(endpoint.Auth.Username) == 0 {
			panic("must supply a username and password with auth")
		}

		// default basic auth
		if endpoint.Auth.Type == "" {
			endpoint.Auth.Type = "basic"
		}
		return d.addAuthToRequest(request, endpoint.Auth)
	}
	return request
}

func (d defaultRequestBuilder) addAuthToRequest(request *http.Request, auth *Auth) *http.Request {
	if auth == nil {
		return request
	}
	if strings.ToLower(auth.Type) == "basic" {
		request.SetBasicAuth(auth.Username, auth.Password)
	}
	return request
}
