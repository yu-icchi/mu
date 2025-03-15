package log

import (
	"fmt"
	"math"
	"strconv"
	"time"
	"unsafe"
)

type kind int

const (
	kindAny kind = iota
	kindBool
	kindDuration
	kindFloat64
	kindInt64
	kindString
	kindTime
	kindUint64
)

type (
	stringptr    *byte
	timeLocation *time.Location
	timeTime     time.Time
)

type value struct {
	num uint64
	any any
}

func (v value) String() string {
	if strPtr, ok := v.any.(stringptr); ok {
		return unsafe.String(strPtr, v.num)
	}
	var buf []byte
	return string(v.append(buf))
}

func (v value) kind() kind {
	switch typ := v.any.(type) {
	case kind:
		return typ
	case stringptr:
		return kindString
	case timeLocation, timeTime:
		return kindTime
	default:
		return kindAny
	}
}

func (v value) str() string {
	return unsafe.String(v.any.(stringptr), v.num)
}

func (v value) bool() bool {
	return v.num == 1
}

func (v value) float() float64 {
	return math.Float64frombits(v.num)
}

func (v value) duration() time.Duration {
	return time.Duration(int64(v.num)) // #nosec G115
}

func (v value) time() time.Time {
	switch typ := v.any.(type) {
	case timeLocation:
		if typ == nil {
			return time.Time{}
		}
		return time.Unix(0, int64(v.num)).In(typ) // #nosec G115
	case timeTime:
		return time.Time(typ)
	default:
		panic("bad time type")
	}
}

func (v value) append(dst []byte) []byte {
	switch v.kind() {
	case kindString:
		return append(dst, v.str()...)
	case kindInt64:
		return strconv.AppendInt(dst, int64(v.num), 10) // #nosec G115
	case kindUint64:
		return strconv.AppendUint(dst, v.num, 10)
	case kindFloat64:
		return strconv.AppendFloat(dst, v.float(), 'g', -1, 64)
	case kindBool:
		return strconv.AppendBool(dst, v.bool())
	case kindDuration:
		return append(dst, v.duration().String()...)
	case kindTime:
		return append(dst, v.time().String()...)
	case kindAny:
		return fmt.Append(dst, v.any)
	default:
		panic("bad kind")
	}
	return nil // nolint: govet
}

func stringValue(val string) value {
	return value{num: uint64(len(val)), any: stringptr(unsafe.StringData(val))}
}

func int64Value(val int64) value {
	return value{num: uint64(val), any: kindInt64} // #nosec G115
}

func intValue(val int) value {
	return int64Value(int64(val))
}

func uint64Value(val uint64) value {
	return value{num: val, any: kindUint64}
}

func float64Value(val float64) value {
	return value{num: math.Float64bits(val), any: kindFloat64}
}

func boolValue(val bool) value {
	var u uint64
	if val {
		u = 1
	}
	return value{num: u, any: kindBool}
}

func timeValue(val time.Time) value {
	if val.IsZero() {
		return value{any: timeLocation(nil)}
	}
	nsec := val.UnixNano()
	t := time.Unix(0, nsec)
	if val.Equal(t) {
		return value{num: uint64(nsec), any: timeLocation(val.Location())} // #nosec G115
	}
	return value{any: timeTime(val.Round(0))}
}

func durationValue(val time.Duration) value {
	return value{num: uint64(val.Nanoseconds()), any: kindDuration} // #nosec G115
}
