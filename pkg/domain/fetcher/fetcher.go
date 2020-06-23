package fetcher

import (
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/yurykabanov/scraper/pkg/domain"
	"go.uber.org/ratelimit"
)

type Fetcher struct {
	client     *http.Client
	rb         RequestBuilder
	validators []ResponseValidator
	rl         ratelimit.Limiter
}

type ValidationError struct {
	nested error
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", err.nested.Error())
}

type Option func(f *Fetcher)

func WithRequestBuilder(rb RequestBuilder) Option {
	return func(f *Fetcher) {
		f.rb = rb
	}
}

func WithResponseValidators(validators ...ResponseValidator) Option {
	return func(f *Fetcher) {
		f.validators = append(f.validators, validators...)
	}
}

func WithRateLimiter(rl ratelimit.Limiter) Option {
	return func(f *Fetcher) {
		f.rl = rl
	}
}

func New(client *http.Client, opts ...Option) *Fetcher {
	f := &Fetcher{
		client: client,
		rb:     DefaultRequestBuilder,
	}

	for _, opt := range opts {
		opt(f)
	}

	if f.validators == nil {
		f.validators = []ResponseValidator{AcceptHttpCodeRange(200, 299)}
	}

	return f
}

func (f *Fetcher) Fetch(task *domain.Task) (*domain.Result, error) {
	log.Debugf("Fetching %+v", task)

	req, err := f.rb(task)
	if err != nil {
		return nil, err
	}

	// Respect rate limits if any
	if f.rl != nil {
		f.rl.Take()
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	for _, validator := range f.validators {
		if err := validator(task, req, resp); err != nil {
			return nil, ValidationError{nested: err}
		}
	}

	// TODO: automatic encoding / parameter
	//
	// Example:
	//
	// dec := charmap.Windows1251.NewDecoder()
	// body, nested = dec.Bytes(body)
	// if nested != nil {
	// 	return nil, nested
	// }

	return &domain.Result{
		Url:        task.Url,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(body),
	}, nil
}
