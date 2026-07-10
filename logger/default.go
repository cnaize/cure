package logger

import "github.com/rs/zerolog"

var DefaultLogger = zerolog.New(zerolog.NewConsoleWriter()).
	With().
	Str("app", "cure").
	Timestamp().
	Logger().
	Level(zerolog.InfoLevel)
