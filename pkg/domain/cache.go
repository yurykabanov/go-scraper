package domain

type Cache interface {
	Get(hash string) (*Result, error)
	Put(hash string, resp *Result) error
}
