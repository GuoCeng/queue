package unit

import (
	"math"
	_ "runtime"
	"time"
	_ "unsafe"
)

const (
	NanoScale   = 1
	MicroScale  = 1000 * NanoScale
	MilliScale  = 1000 * MicroScale
	SecondScale = 1000 * MilliScale
	MinuteScale = 60 * SecondScale
	HourScale   = 60 * MinuteScale
	DayScale    = 24 * HourScale
)

var (
	NANOSECONDS  = New(NanoScale)
	MICROSECONDS = New(MicroScale)
	MILLISECONDS = New(MilliScale)
	SECONDS      = New(SecondScale)
	MINUTES      = New(MinuteScale)
	HOURS        = New(HourScale)
	DAYS         = New(DayScale)
)

// runtimeNano returns the current value of the runtime clock in nanoseconds.
//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

func HiResClockMs() int64 {
	return NANOSECONDS.ToMillis(runtimeNano())
}

func GetHiResClockMs(now time.Time) int64 {
	return NANOSECONDS.ToMillis(runtimeNano())
}

type TimeUnit struct {
	scale      int64
	maxNanos   int64
	maxMicros  int64
	maxMillis  int64
	maxSecs    int64
	microRatio int64
	milliRatio int32
	secRatio   int32
}

func New(s int64) *TimeUnit {
	var ur int64
	if s >= MicroScale {
		ur = s / MicroScale
	} else {
		ur = MicroScale / s
	}
	var mr int64
	if s >= MilliScale {
		mr = s / MilliScale
	} else {
		mr = MilliScale / s
	}
	var sr int64
	if s >= SecondScale {
		sr = s / SecondScale
	} else {
		sr = SecondScale / s
	}
	return &TimeUnit{
		scale:      s,
		maxNanos:   math.MaxInt64 / s,
		maxMicros:  math.MaxInt64 / ur,
		maxMillis:  math.MaxInt64 / mr,
		maxSecs:    math.MaxInt64 / sr,
		microRatio: ur,
		milliRatio: int32(mr),
		secRatio:   int32(sr),
	}
}

func (tu *TimeUnit) cvt(d, dst, src int64) int64 {
	if src == dst {
		return d
	} else if src < dst {
		return d / (dst / src)
	} else {
		r := src / dst
		m := math.MaxInt64 / r
		if d > m {
			return math.MaxInt64
		} else if d < -m {
			return math.MinInt64
		} else {
			return d * r
		}
	}
}

func (tu *TimeUnit) Convert(sourceDuration int64, sourceUnit *TimeUnit) int64 {
	switch tu.scale {
	case NanoScale:
		return sourceUnit.ToNanos(sourceDuration)
	case MicroScale:
		return sourceUnit.ToMicros(sourceDuration)
	case MilliScale:
		return sourceUnit.ToMillis(sourceDuration)
	case SecondScale:
		return sourceUnit.ToSeconds(sourceDuration)
	default:
		return tu.cvt(sourceDuration, tu.scale, sourceUnit.scale)
	}
}

func (tu *TimeUnit) ConvertTime(time time.Time) int64 {
	secs := time.Second()
	nano := time.Nanosecond()
	if secs < 0 && nano > 0 {
		// use representation compatible with integer division
		secs++
		nano -= SecondScale
	}
	var s, nanoVal int64
	if tu.scale == NanoScale {
		nanoVal = int64(nano)
	} else {
		s = tu.scale
		if s >= SecondScale {
			if tu.scale == SecondScale {
				return int64(secs)
			}
			return int64(secs / int(tu.secRatio))
		}
		nanoVal = int64(nano) / s
	}
	var val = int64(secs*int(tu.secRatio)) + nanoVal
	if (int64(secs) >= tu.maxSecs || int64(secs) <= -tu.maxSecs) && (int64(secs) != tu.maxSecs || val <= 0) && (int64(secs) != -tu.maxSecs || val >= 0) {
		if secs > 0 {
			return math.MaxInt64
		}
		return math.MinInt64
	} else {
		return val
	}
}

func (tu *TimeUnit) ToNanos(duration int64) int64 {
	s := tu.scale
	if s == NanoScale {
		return duration
	} else {
		m := tu.maxNanos
		if duration > m {
			return math.MaxInt64
		} else if duration < -m {
			return math.MinInt64
		} else {
			return duration * s
		}
	}
}

func (tu *TimeUnit) ToMicros(duration int64) int64 {
	s := tu.scale
	if s == MicroScale {
		return duration
	} else if s < MicroScale {
		return duration / tu.microRatio
	} else {
		m := tu.maxMicros
		if duration > m {
			return math.MaxInt64
		} else if duration < -m {
			return math.MinInt64
		} else {
			return duration * tu.microRatio
		}
	}
}

func (tu *TimeUnit) ToMillis(duration int64) int64 {
	s := tu.scale
	if s == MilliScale {
		return duration
	} else if s < MilliScale {
		return duration / int64(tu.milliRatio)
	} else {
		m := tu.maxMillis
		if duration > m {
			return math.MaxInt64
		} else if duration < -m {
			return math.MinInt64
		} else {
			return duration * int64(tu.milliRatio)
		}
	}
}

func (tu *TimeUnit) ToSeconds(duration int64) int64 {
	s := tu.scale
	if s == SecondScale {
		return duration
	} else if s < SecondScale {
		return duration / int64(tu.secRatio)
	} else {
		m := tu.maxSecs
		if duration > m {
			return math.MaxInt64
		} else if duration < -m {
			return math.MinInt64
		} else {
			return duration * int64(tu.secRatio)
		}
	}
}

func (tu *TimeUnit) ToMinutes(duration int64) int64 {
	return tu.cvt(duration, MinuteScale, tu.scale)
}

func (tu *TimeUnit) ToHours(duration int64) int64 {
	return tu.cvt(duration, HourScale, tu.scale)
}

func (tu *TimeUnit) ToDays(duration int64) int64 {
	return tu.cvt(duration, DayScale, tu.scale)
}

func (tu *TimeUnit) excessNanos(d, m int64) int {
	s := tu.scale
	if s == NanoScale {
		return int(d - m*MilliScale)
	} else if s == MicroScale {
		return int(d*MicroScale - m*MilliScale)
	} else {
		return 0
	}
}
