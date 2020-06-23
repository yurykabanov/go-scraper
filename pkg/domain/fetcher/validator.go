package fetcher

import (
	"fmt"
	"net/http"

	"github.com/yurykabanov/scraper/pkg/domain"
)

type ResponseValidator func(*domain.Task, *http.Request, *http.Response) error

type HttpCodeUnacceptableError struct {
	StatusCode int
}

func (err HttpCodeUnacceptableError) Error() string {
	return fmt.Sprintf("validation error: response code %d is not acceptable", err.StatusCode)
}

var AcceptHttpCodes = func(codes ...int) ResponseValidator {
	return func(task *domain.Task, req *http.Request, resp *http.Response) error {
		for _, c := range codes {
			if resp.StatusCode == c {
				return nil
			}
		}
		return HttpCodeUnacceptableError{StatusCode: resp.StatusCode}
	}
}

var AcceptHttpCodeRange = func(a, b int) ResponseValidator {
	return func(task *domain.Task, req *http.Request, resp *http.Response) error {
		if a <= resp.StatusCode && resp.StatusCode <= b {
			return nil
		}

		return HttpCodeUnacceptableError{StatusCode: resp.StatusCode}
	}
}
