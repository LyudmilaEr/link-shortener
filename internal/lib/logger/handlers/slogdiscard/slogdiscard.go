package slogdiscard

import (
	"context"
	"log/slog"
)

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

type DiscardHandler struct{}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// Просто игнорируем запись журнала
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Возвращает тот же обработчик, тк нет атрибутивов для сохранения
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	// Возвращает тот же обработчик, тк нет атрибутивов для сохранения
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// Всегда возвращает false, тк запись журнала игнорируется
	return false
}

// New is a convenience function that creates a new discard logger
func New() *slog.Logger {
	return NewDiscardLogger()
}
