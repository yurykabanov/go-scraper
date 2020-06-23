package fetcher

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yurykabanov/scraper/pkg/domain"
)

func TestHttpCodeUnacceptableError_Error(t *testing.T) {
	err := HttpCodeUnacceptableError{StatusCode: 123}

	assert.EqualError(t, err, "validation error: response code 123 is not acceptable")
}

func TestAcceptHttpCodes(t *testing.T) {
	tests := []struct {
		name            string
		acceptableCodes []int
		response        http.Response
		expected        bool
	}{
		{
			name:            "good response",
			acceptableCodes: []int{200, 204},
			response:        http.Response{StatusCode: 200},
			expected:        true,
		},
		{
			name:            "bad response",
			acceptableCodes: []int{200},
			response:        http.Response{StatusCode: 404},
			expected:        false,
		},
	}

	task := domain.Task{Url: "http://domain.tld/whatever", TaskRef: "some_task_ref"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := AcceptHttpCodes(test.acceptableCodes...)(&task, nil, &test.response)

			assert.False(t, err == nil && !test.expected, "expected response code to be unacceptable, but it is acceptable")

			assert.False(t, err != nil && test.expected, "expected response code to be acceptable, but it is unacceptable")
		})
	}
}

func TestAcceptHttpCodeRange(t *testing.T) {
	tests := []struct {
		name                 string
		acceptableRangeStart int
		acceptableRangeEnd   int
		responses            []http.Response
		expected             bool
	}{
		{
			name:                 "good response",
			acceptableRangeStart: 200,
			acceptableRangeEnd:   202,
			responses: []http.Response{
				{StatusCode: 200},
				{StatusCode: 201},
				{StatusCode: 202},
			},
			expected: true,
		},
		{
			name:                 "bad response",
			acceptableRangeStart: 200,
			acceptableRangeEnd:   202,
			responses: []http.Response{
				{StatusCode: 198},
				{StatusCode: 199},
				{StatusCode: 203},
				{StatusCode: 204},
			},
			expected: false,
		},
	}

	task := domain.Task{Url: "http://domain.tld/whatever", TaskRef: "some_task_ref"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, resp := range test.responses {
				err := AcceptHttpCodeRange(test.acceptableRangeStart, test.acceptableRangeEnd)(&task, nil, &resp)

				assert.False(t, err == nil && !test.expected, "expected response code to be unacceptable, but it is acceptable")

				assert.False(t, err != nil && test.expected, "expected response code to be acceptable, but it is unacceptable")
			}
		})
	}
}
