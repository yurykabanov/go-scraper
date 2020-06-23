package domain

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type Action interface {
	Name() string
	Apply(*FetchedTask, *html.Node) (interface{}, []Task)
}

type internalTaskDefinition struct {
	Type         string `yaml:"type"`
	TaskRef      string `yaml:"task_ref,omitempty"`
	ContentName  string `yaml:"content_name,omitempty"`
	SelectorType string `yaml:"selector_type,omitempty"` // Reserved
	Selector     string `yaml:"selector"`
}

type TaskDefinition struct {
	Name    string   `yaml:"name"`
	Actions []Action `yaml:"-"`
}

func (td *TaskDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data struct {
		Name    string                   `yaml:"name"`
		Actions []internalTaskDefinition `yaml:"actions"`
	}

	err := unmarshal(&data)
	if err != nil {
		return err
	}

	td.Name = data.Name

	// TODO: is there better approach?

	for _, action := range data.Actions {
		var a Action
		switch strings.ToLower(action.Type) {
		case "task":
			a = &TaskAction{
				TaskRef:  action.TaskRef,
				Selector: action.Selector,
			}
		case "text":
			a = &TextAction{
				ContentName: action.ContentName,
				Selector:    action.Selector,
			}
		case "html":
			a = &HtmlAction{
				ContentName: action.ContentName,
				Selector:    action.Selector,
			}
		}

		td.Actions = append(td.Actions, a)
	}

	return nil
}

type FileAction struct {
}

func (a *FileAction) Name() string {
	panic("implement me")
}

func (a *FileAction) Apply(*FetchedTask, *html.Node) (interface{}, []Task) {
	panic("implement me")
}

type HtmlAction struct {
	ContentName string
	Selector    string
}

func (a *HtmlAction) Name() string {
	return a.ContentName
}

func (a *HtmlAction) Apply(task *FetchedTask, root *html.Node) (interface{}, []Task) {
	var data []string

	for _, n := range htmlquery.Find(root, a.Selector) {
		buf := new(bytes.Buffer)

		err := html.Render(buf, n)
		if err != nil {
			logrus.WithError(err).Fatal("Unable to render html partial to buffer")
		}

		data = append(data, buf.String())
	}

	return data, nil
}

type TaskAction struct {
	Selector string
	TaskRef  string
}

func (a *TaskAction) Name() string {
	return a.TaskRef
}

func (a *TaskAction) Apply(task *FetchedTask, root *html.Node) (interface{}, []Task) {
	var tasks []Task
	var baseUrl *url.URL

	n := htmlquery.FindOne(root, "//head/base/@href")
	if n != nil {
		var err error
		baseUrl, err = url.Parse(htmlquery.InnerText(n))
		logrus.WithError(err).Error("Unable to parse base href URL when it's present on page")
	}

	taskUrl, err := url.Parse(task.Url)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to parse task URL after page has been successfully loaded though this should never happen")
	}

	for _, n := range htmlquery.Find(root, a.Selector) {
		var resultUrl *url.URL

		href := htmlquery.InnerText(n)
		hrefUrl, err := url.Parse(href)
		if err != nil {
			logrus.WithError(err).Error("Unable to parse href retrieved with selector. Bad selector?")
			continue
		}

		if hrefUrl.IsAbs() {
			resultUrl = hrefUrl
		} else {
			if baseUrl != nil {
				resultUrl = baseUrl.ResolveReference(hrefUrl)
			} else {
				resultUrl = taskUrl.ResolveReference(hrefUrl)
			}
		}

		tasks = append(tasks, Task{
			Url:     resultUrl.String(),
			TaskRef: a.TaskRef,
		})
	}

	return nil, tasks
}

type TextAction struct {
	ContentName string
	Selector    string
}

func (a *TextAction) Name() string {
	return a.ContentName
}

func (a *TextAction) Apply(task *FetchedTask, root *html.Node) (interface{}, []Task) {
	var data []string

	for _, n := range htmlquery.Find(root, a.Selector) {
		data = append(data, htmlquery.InnerText(n))
	}

	return data, nil
}
