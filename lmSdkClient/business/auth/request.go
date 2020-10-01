package auth

import (
	"io"
	"net/http"
	"net/url"
	"context"
)

type request struct {
	method     string
	endpoint   string
	query      url.Values
	header     http.Header
	body       io.Reader
	fullURL    string
}

func NewContext() context.Context {
	return context.Background()
}

func (r *request) validate() (err error) {
	if r.query == nil {
		r.query = url.Values{}
	}

	return nil
}

// RequestOption define option type for request
type RequestOption func(*request)

