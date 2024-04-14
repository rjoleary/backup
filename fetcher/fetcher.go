package fetcher

import "fmt"

type Fetcher interface {
	fmt.Stringer
	Name() string
	Validate() error
	Fetch(stagingDir string) error
}
