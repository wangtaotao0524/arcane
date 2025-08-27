//Originally from pocket-id

package dto

import (
	"fmt"

	"github.com/jinzhu/copier"
)

// MapSlice maps []S -> []D and returns the new slice
func MapSlice[S any, D any](source []S) ([]D, error) {
	dest := make([]D, len(source))
	for i := range source {
		if err := MapStruct(source[i], &dest[i]); err != nil {
			return nil, fmt.Errorf("failed to map field %d: %w", i, err)
		}
	}
	return dest, nil
}

// MapOne maps S -> D and returns the new value
func MapOne[S any, D any](source S) (D, error) {
	var dest D
	if err := MapStruct(source, &dest); err != nil {
		return dest, err
	}
	return dest, nil
}

func MapStructList[S any, D any](source []S, destination *[]D) (err error) {
	*destination = make([]D, len(source))

	for i, item := range source {
		err = MapStruct(item, &((*destination)[i]))
		if err != nil {
			return fmt.Errorf("failed to map field %d: %w", i, err)
		}
	}
	return nil
}

func MapStruct(source any, destination any) error {
	return copier.CopyWithOption(destination, source, copier.Option{
		DeepCopy: true,
	})
}
