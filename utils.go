package main

import (
	"github.com/phuslu/log"
)

func MustSucceed(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error")
	}
}

func Must[T any](value T, err error) T {
	MustSucceed(err)
	return value
}

func Keys[T comparable, U any](dict map[T]U) []T {
	keys := make([]T, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	return keys
}
