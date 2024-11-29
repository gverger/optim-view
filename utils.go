package main

import (
	"github.com/phuslu/log"
)

func MustSucceed(err error) {
	if err != nil {
		log.Fatal().Err(err)
	}
}

func Must[T any](value T, err error) T {
	MustSucceed(err)
	return value
}
