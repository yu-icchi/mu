package log

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogger_Debug(t *testing.T) {
	t.Parallel()
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	buf := new(strings.Builder)
	logger := New(buf)
	logger.Debug("test", Title("title"),
		Bool("flag", true), String("str", "test"),
		Int("int", 100), Int64("int64", 10000),
		Uint64("uint", 200), Uint64("uint64", 20000),
		Float64("float64", 3.14),
		Time("date", time.Date(2025, 1, 20, 12, 10, 30, 0, time.UTC)),
		Duration("duration", 3*time.Hour),
		Error(err1), Error(fmt.Errorf("%w: %w", err1, err2)))
	require.Equal(t, "::debug title=title ::DEBUG test flag=true str=test int=100 int64=10000 uint=200 uint64=20000 float64=3.14 date=2025-01-20 12:10:30 +0000 UTC duration=3h0m0s%0Aerror=err1%0Aerr1: err2", buf.String())
}

func TestLogger_Info(t *testing.T) {
	t.Parallel()
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	buf := new(strings.Builder)
	logger := New(buf)
	logger.Info("test", Title("title"),
		Bool("flag", true), String("str", "test"),
		Int("int", 100), Int64("int64", 10000),
		Uint64("uint", 200), Uint64("uint64", 20000),
		Float64("float64", 3.14),
		Time("date", time.Date(2025, 1, 20, 12, 10, 30, 0, time.UTC)),
		Duration("duration", 3*time.Hour),
		Error(err1), Error(fmt.Errorf("%w: %w", err1, err2)))
	require.Equal(t, "::notice title=title ::INFO test flag=true str=test int=100 int64=10000 uint=200 uint64=20000 float64=3.14 date=2025-01-20 12:10:30 +0000 UTC duration=3h0m0s%0Aerror=err1%0Aerr1: err2", buf.String())
}

func TestLogger_Warn(t *testing.T) {
	t.Parallel()
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	buf := new(strings.Builder)
	logger := New(buf)
	logger.Warn("test", Title("title"),
		Bool("flag", true), String("str", "test"),
		Int("int", 100), Int64("int64", 10000),
		Uint64("uint", 200), Uint64("uint64", 20000),
		Float64("float64", 3.14),
		Time("date", time.Date(2025, 1, 20, 12, 10, 30, 0, time.UTC)),
		Duration("duration", 3*time.Hour),
		Error(err1), Error(fmt.Errorf("%w: %w", err1, err2)))
	require.Equal(t, "::warning title=title ::WARN test flag=true str=test int=100 int64=10000 uint=200 uint64=20000 float64=3.14 date=2025-01-20 12:10:30 +0000 UTC duration=3h0m0s%0Aerror=err1%0Aerr1: err2", buf.String())
}

func TestLogger_Error(t *testing.T) {
	t.Parallel()
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	buf := new(strings.Builder)
	logger := New(buf)
	logger.Error("test", Title("title"),
		Bool("flag", true), String("str", "test"),
		Int("int", 100), Int64("int64", 10000),
		Uint64("uint", 200), Uint64("uint64", 20000),
		Float64("float64", 3.14),
		Time("date", time.Date(2025, 1, 20, 12, 10, 30, 0, time.UTC)),
		Duration("duration", 3*time.Hour),
		Error(err1), Error(fmt.Errorf("%w: %w", err1, err2)))
	require.Equal(t, "::error title=title ::ERROR test flag=true str=test int=100 int64=10000 uint=200 uint64=20000 float64=3.14 date=2025-01-20 12:10:30 +0000 UTC duration=3h0m0s%0Aerror=err1%0Aerr1: err2", buf.String())
}
