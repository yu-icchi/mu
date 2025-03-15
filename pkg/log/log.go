package log

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const newline = "%0A"

type options struct {
	title string
	attrs []attr
	errs  []error
}

type attr struct {
	key   string
	value value
}

func (a attr) String() string {
	return a.key + "=" + a.value.String()
}

type Option func(o *options)

func Title(title string) Option {
	return func(o *options) {
		o.title = title
	}
}

func String(key, val string) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: stringValue(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Int64(key string, val int64) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: int64Value(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Int(key string, val int) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: intValue(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Uint64(key string, val uint64) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: uint64Value(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Float64(key string, val float64) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: float64Value(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Bool(key string, val bool) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: boolValue(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Time(key string, val time.Time) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: timeValue(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Duration(key string, val time.Duration) Option {
	return func(o *options) {
		attr := attr{
			key:   key,
			value: durationValue(val),
		}
		o.attrs = append(o.attrs, attr)
	}
}

func Error(err error) Option {
	return func(o *options) {
		o.errs = append(o.errs, err)
	}
}

type Logger interface {
	Debug(msg string, opts ...Option)
	Info(msg string, opts ...Option)
	Warn(msg string, opts ...Option)
	Error(msg string, opts ...Option)
}

func New(w io.Writer) Logger {
	return &logger{
		stdout: w,
	}
}

type logger struct {
	stdout io.Writer
}

func (l *logger) Debug(msg string, opts ...Option) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}
	builder := strings.Builder{}
	builder.WriteString("::debug")
	if opt.title != "" {
		builder.WriteString(" title=")
		builder.WriteString(opt.title)
		builder.WriteString(" ")
	}
	builder.WriteString("::")
	builder.WriteString("DEBUG")
	builder.WriteString(" ")
	builder.WriteString(msg)
	for _, attr := range opt.attrs {
		builder.WriteString(" ")
		builder.WriteString(attr.String())
	}
	builder.WriteString(newline)
	err := errors.Join(opt.errs...)
	if err != nil {
		builder.WriteString("error=")
		builder.WriteString(strings.ReplaceAll(err.Error(), "\n", newline))
	}
	_, _ = fmt.Fprint(l.stdout, builder.String())
}

func (l *logger) Info(msg string, opts ...Option) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}
	builder := strings.Builder{}
	builder.WriteString("::notice")
	if opt.title != "" {
		builder.WriteString(" title=")
		builder.WriteString(opt.title)
		builder.WriteString(" ")
	}
	builder.WriteString("::")
	builder.WriteString("INFO")
	builder.WriteString(" ")
	builder.WriteString(msg)
	for _, attr := range opt.attrs {
		builder.WriteString(" ")
		builder.WriteString(attr.String())
	}
	builder.WriteString(newline)
	err := errors.Join(opt.errs...)
	if err != nil {
		builder.WriteString("error=")
		builder.WriteString(strings.ReplaceAll(err.Error(), "\n", newline))
	}
	_, _ = fmt.Fprint(l.stdout, builder.String())
}

func (l *logger) Warn(msg string, opts ...Option) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}
	builder := strings.Builder{}
	builder.WriteString("::warning")
	if opt.title != "" {
		builder.WriteString(" title=")
		builder.WriteString(opt.title)
		builder.WriteString(" ")
	}
	builder.WriteString("::")
	builder.WriteString("WARN")
	builder.WriteString(" ")
	builder.WriteString(msg)
	for _, attr := range opt.attrs {
		builder.WriteString(" ")
		builder.WriteString(attr.String())
	}
	builder.WriteString(newline)
	err := errors.Join(opt.errs...)
	if err != nil {
		builder.WriteString("error=")
		builder.WriteString(strings.ReplaceAll(err.Error(), "\n", newline))
	}
	_, _ = fmt.Fprint(l.stdout, builder.String())
}

func (l *logger) Error(msg string, opts ...Option) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}
	builder := strings.Builder{}
	builder.WriteString("::error")
	if opt.title != "" {
		builder.WriteString(" title=")
		builder.WriteString(opt.title)
		builder.WriteString(" ")
	}
	builder.WriteString("::")
	builder.WriteString("ERROR")
	builder.WriteString(" ")
	builder.WriteString(msg)
	for _, attr := range opt.attrs {
		builder.WriteString(" ")
		builder.WriteString(attr.String())
	}
	builder.WriteString(newline)
	err := errors.Join(opt.errs...)
	if err != nil {
		builder.WriteString("error=")
		builder.WriteString(strings.ReplaceAll(err.Error(), "\n", newline))
	}
	_, _ = fmt.Fprint(l.stdout, builder.String())
}
