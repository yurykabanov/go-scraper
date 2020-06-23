package fetcher

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yurykabanov/scraper/pkg/domain"
)

func TestDefaultRequestBuilder_GoodTask(t *testing.T) {
	task := domain.Task{Url: "http://domain.tld/some/url", TaskRef: "some_task_ref"}

	req, err := DefaultRequestBuilder(&task)

	ua := req.Header.Get("User-Agent")

	assert.Nil(t, err, "error should be nil")
	assert.NotEqual(t, ua, "", "user agent should not be empty string")
	assert.True(t, strings.Contains(ua, "Chrome/"), "user agent should be like Chrome's one")
}

func TestDefaultRequestBuilder_BadTask(t *testing.T) {
	tests := []struct {
		name string
		task domain.Task
	}{
		{
			name: "relative url",
			task: domain.Task{Url: "relative_url", TaskRef: "some_task_ref"},
		},
		{
			name: "bad url",
			task: domain.Task{Url: "http://192.168.0.%31/", TaskRef: "some_task_ref"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := DefaultRequestBuilder(&test.task)

			assert.Nil(t, req, "request should be nil")
			assert.NotNil(t, err, "error should not be nil")
		})
	}
}

