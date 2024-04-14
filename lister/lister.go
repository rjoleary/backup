package lister

import (
	"fmt"

	"github.com/rjoleary/backup/fetcher"
)

type Lister interface {
	fmt.Stringer
	Name() string
	Validate() error
	List() ([]fetcher.Fetcher, error)
}
