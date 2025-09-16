package slogpretty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"

	"github.com/fatih/color"
)

type PrettyhandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type Prettyhandler struct {
	opts PrettyhandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyhandlerOptions) NewPrettyHandler(
	out io.Writer,
) *Prettyhandler {
	h := &Prettyhandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func (h *Prettyhandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, v := range h.attrs {
		fields[v.Key] = v.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			fmt.Errorf("%s: %w", b, err)
		}
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

func (h *Prettyhandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Prettyhandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *Prettyhandler) WithGroup(name string) slog.Handler {
	// TODO: implement
	return &Prettyhandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}
