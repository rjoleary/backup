package archiver

import "fmt"

type Archiver interface {
	fmt.Stringer
	Name() string
	Validate() error
	Archive(diskImage string) error
}
