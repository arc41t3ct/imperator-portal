package cache

type Cache interface {
	Has(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, ...int) error
	Forget(string) error
	EmptyMatching(string) error
	Empty() error
}
