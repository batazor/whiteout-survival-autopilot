package imagefinder

import "fmt"

func ErrImageNotLoaded(name string) error {
	return fmt.Errorf("failed to load %s image", name)
}
