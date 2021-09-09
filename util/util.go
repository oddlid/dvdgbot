package util

import (
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

var _log = log.With().Str("package", "util").Logger()

func EnvDefStr(key, fallback string) string {
	val, found := os.LookupEnv(key)
	if !found {
		_log.Debug().
			Str("func", "EnvDefStr").
			Str("env_key", key).
			Str("fallback", fallback).
			Msg("Undefined, returning fallback")
		return fallback
	}
	return val // might still be empty, if set, but empty in ENV
}

func EnvDefInt(key string, fallback int) int {
	val, found := os.LookupEnv(key)
	if !found {
		_log.Debug().
			Str("func", "EnvDefInt").
			Str("env_key", key).
			Int("fallback", fallback).
			Msg("Undefined, returning fallback")
		return fallback
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		_log.Error().
			Str("func", "EnvDefInt").
			Str("env_key", key).
			Int("fallback", fallback).
			Str("value", val).
			Err(err).
			Msg("Conversion error")
		return fallback
	}
	return intVal
}
