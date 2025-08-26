//Originally from pocket-id

package dto

import (
	"fmt"

	"github.com/jinzhu/copier"
)

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
