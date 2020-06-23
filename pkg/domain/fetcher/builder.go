package fetcher

import (
	"errors"
	"net/http"

	"github.com/yurykabanov/scraper/pkg/domain"
)

type RequestBuilder func(*domain.Task) (*http.Request, error)

var (
	ErrTaskUrlNotAbsolute = errors.New("task's URL must be absolute")
)

var DefaultRequestBuilder RequestBuilder = func(task *domain.Task) (*http.Request, error) {
	req, err := http.NewRequest("GET", task.Url, nil)
	if err != nil {
		return nil, err
	}

	if !req.URL.IsAbs() {
		return nil, ErrTaskUrlNotAbsolute
	}

	// TODO: more headers like Chrome
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")

	// TODO: option to mimic google crawler (research needed)

	return req, nil
}
