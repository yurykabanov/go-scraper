package fetcher

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yurykabanov/scraper/pkg/domain"
)

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{nested: errors.New("some nested error")}
	assert.EqualError(t, err, "validation error: some nested error")
}

type RoundTripperFunc func(req *http.Request) *http.Response

func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req), nil
}

type RoundTripperError struct {
	err error
}

func (rt RoundTripperError) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

func httpClientWithResponse(statusCode int, body io.ReadCloser, headers http.Header) *http.Client {
	return &http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: statusCode,
				Body:       body,
				Header:     headers,
			}
		}),
	}
}

func httpClientWithError(err error) *http.Client {
	return &http.Client{
		Transport: RoundTripperError{err: err},
	}
}

var task = domain.Task{Url: "http://domain.tld/whatever", TaskRef: "some_task_ref"}

func TestFetcher_Fetch_Basic(t *testing.T) {
	statusCode := 234
	body := "some data"
	headers := http.Header{"X-Fake-Response": []string{"yes"}}

	f := New(httpClientWithResponse(statusCode, ioutil.NopCloser(bytes.NewBufferString(body)), headers))

	result, err := f.Fetch(&task)

	assert.NotNil(t, result, "Fetcher should return result")
	assert.Nil(t, err, "Fetcher shouldn't return error")

	assert.Equalf(t, result.StatusCode, statusCode,
		"Fetcher returned result with unexpected status code %d", result.StatusCode)

	assert.Equalf(t, body, result.Body, "Fetcher returned result with unexpected body content: '%s'", result.Body)

	assert.Contains(t, result.Headers, "X-Fake-Response", "Fetcher returned result with missing 'X-Fake-Response' header")
}

func TestFetcher_Fetch_RequestBuilderFailure(t *testing.T) {
	expectedError := errors.New("bad task")

	rb := func(task *domain.Task) (*http.Request, error) {
		return nil, expectedError
	}

	f := New(httpClientWithResponse(200, ioutil.NopCloser(bytes.NewBufferString("")), make(http.Header)), WithRequestBuilder(rb))

	result, err := f.Fetch(&task)

	assert.Nil(t, result, "Fetcher shouldn't return result")
	assert.NotNil(t, err, "Fetcher should return error")

	assert.Equalf(t, expectedError, err, "Fetcher returned unexpected error: %s", err.Error())
}

func TestFetcher_Fetch_HttpClientError(t *testing.T) {
	expectedError := errors.New("some error")
	f := New(httpClientWithError(expectedError))

	result, err := f.Fetch(&task)

	assert.Nil(t, result, "Fetcher shouldn't return result")
	assert.NotNil(t, err, "Fetcher should return error")

	// NOTE: error is wrapped by http.Client
	assert.EqualError(t, err, "Get http://domain.tld/whatever: "+expectedError.Error())
}

type dummyRateLimiter struct {
	mock.Mock
}

func (rl dummyRateLimiter) Take() time.Time {
	args := rl.Called()
	return args.Get(0).(time.Time)
}

func TestFetcher_Fetch_RateLimiter(t *testing.T) {
	rl := dummyRateLimiter{}

	rl.On("Take").Return(time.Time{}).Once()

	f := New(httpClientWithResponse(200, ioutil.NopCloser(bytes.NewBufferString("")), make(http.Header)), WithRateLimiter(rl))

	_, _ = f.Fetch(&task)

	rl.AssertExpectations(t)
}

func TestFetcher_Fetch_ResponseValidators(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{
			name:    "validation passes",
			isValid: true,
		},
		{
			name:    "validation fails",
			isValid: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isCalled := false

			rv := func(*domain.Task, *http.Request, *http.Response) error {
				isCalled = true
				if test.isValid {
					return nil
				}
				return errors.New("some error")
			}

			f := New(httpClientWithResponse(200, ioutil.NopCloser(bytes.NewBufferString("")), make(http.Header)), WithResponseValidators(rv))

			_, _ = f.Fetch(&task)

			assert.True(t, isCalled)
		})
	}
}
