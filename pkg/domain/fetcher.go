package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

type Fetcher interface {
	Fetch(task *Task) (*Result, error)
}

type Result struct {
	Url        string      `json:"url"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
}

func (r *Result) Hash() string {
	hasher := sha256.New()
	hasher.Write([]byte(r.Url))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
