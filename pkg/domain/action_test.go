package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func fetchedTaskWithBody(body string) *FetchedTask {
	return &FetchedTask{
		Task: Task{
			Url:     "http://domain.tld/some_path/some_page",
			TaskRef: "some_task_ref",
		},
		Result: &Result{
			StatusCode: 200,
			Url:        "http://domain.tld/some_path/some_page",
			Body:       body,
			Headers:    map[string][]string{},
		},
	}
}

func TestHtmlAction_Apply(t *testing.T) {
	body := `
<html>
  <body>
		<div id="some-id">Some content</div>
    <div id="another-id">Another content</div>
	</body>
</html>
`
	task := fetchedTaskWithBody(body)
	root, _ := html.Parse(strings.NewReader(body))

	action := HtmlAction{ContentName: "whatever", Selector: "//div[@id='some-id']"}

	result, newTasks := action.Apply(task, root)

	assert.Equal(t, []string{`<div id="some-id">Some content</div>`}, result, "html action should return queried content")
	assert.Nil(t, newTasks, "html action shouldn't return any new tasks")
}

func TestTextAction_Apply(t *testing.T) {
	body := `
<html>
  <body>
		<div id="some-id">Some content <strong>here</strong></div>
    <div id="another-id">Another content</div>
	</body>
</html>
`
	task := fetchedTaskWithBody(body)
	root, _ := html.Parse(strings.NewReader(body))

	action := TextAction{ContentName: "whatever", Selector: "//div[@id='some-id']"}

	result, newTasks := action.Apply(task, root)

	assert.Equal(t, []string{"Some content here"}, result, "text action should return queried content")
	assert.Nil(t, newTasks, "text action shouldn't return any new tasks")
}

func TestTaskAction_Apply(t *testing.T) {
	body := `
<html>
  <body>
		<a href="path_relative">Path relative</a>
    <a href="/root_relative">Root relative</a>
	</body>
</html>
`
	task := fetchedTaskWithBody(body)
	root, _ := html.Parse(strings.NewReader(body))

	action := TaskAction{TaskRef: "whatever", Selector: "//a/@href"}

	result, newTasks := action.Apply(task, root)

	assert.Nil(t, result, "task action shouldn't return queried content")
	assert.Equal(t, []Task{
		{Url: "http://domain.tld/some_path/path_relative", TaskRef: "whatever"},
		{Url: "http://domain.tld/root_relative", TaskRef: "whatever"},
	}, newTasks, "task action should return new tasks with absolute urls")
}
