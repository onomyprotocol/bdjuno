package utils

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"
)

// WatchMethod allows to watch for a method that returns an error.
// It executes the given method in a goroutine, logging any error that might raise.
func WatchMethod(method func() error) {
	go func() {
		err := method()
		if err != nil {
			err = fmt.Errorf("watch method: %s error: %w",
				runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name(), err)
			log.Error().Err(err).Send()
		}
	}()
}
