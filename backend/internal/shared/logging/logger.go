package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, fields map[string]any)
}

type ZeroLogger struct {
	base zerolog.Logger
}

func (z *ZeroLogger) Info(ctx context.Context, msg string, fields map[string]any) {
	e := z.withContext(ctx).Info()
	for k, v := range fields {
		e = e.Any(k, v)
	}
	e.Msg(msg)
}

func (z *ZeroLogger) Warn(ctx context.Context, msg string, fields map[string]any) {
	e := z.withContext(ctx).Warn()

	for k, v := range fields {
		e = e.Any(k, v)
	}
	e.Msg(msg)
}

func (z *ZeroLogger) Error(ctx context.Context, msg string, fields map[string]any) {
	e := z.withContext(ctx).Error()
	for k, v := range fields {
		e = e.Any(k, v)
	}
	e.Msg(msg)
}

func (z *ZeroLogger) withContext(ctx context.Context) *zerolog.Logger {
	log := z.base

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		log = log.With().Str("request_id", reqID).Logger()
	}

	return &log
}

func Init(serviceName string) Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	file, err := os.OpenFile("logs/logs.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open log file")

	}

	multi := zerolog.MultiLevelWriter(consoleWriter, file)

	base := zerolog.New(multi).
		With().
		Timestamp().
		Str("service", serviceName).Logger()
	return &ZeroLogger{base: base}
}

type requestIDKeyType struct{}

var RequestIDKey = requestIDKeyType{}
