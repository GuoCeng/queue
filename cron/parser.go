package cron

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ParseOption int

const (
	Second ParseOption = 1 << iota
	Minute
	Hour
	Dom
	Month
	Dow
)

var places = []ParseOption{
	Second,
	Minute,
	Hour,
	Dom,
	Month,
	Dow,
}

var defaults = []string{
	"0",
	"0",
	"0",
	"*",
	"*",
	"*",
}

var standardParser = NewParser(
	Second | Minute | Hour | Dom | Month | Dow,
)

type Parser struct {
	options ParseOption
}

func NewParser(options ParseOption) Parser {
	return Parser{options}
}

func (p Parser) Parse(spec string) (Schedule, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("empty spec string")
	}

	// Split on whitespace.
	fields := strings.Fields(spec)

	// Validate & fill in any omitted or optional fields
	var err error
	fields, err = normalizeFields(fields, p.options)
	if err != nil {
		return nil, err
	}

	field := func(field string, r bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(field, r)
		return bits
	}

	var (
		second     = field(fields[0], seconds)
		minute     = field(fields[1], minutes)
		hour       = field(fields[2], hours)
		dayofmonth = field(fields[3], dom)
		month      = field(fields[4], months)
		dayofweek  = field(fields[5], dow)
	)
	if err != nil {
		return nil, err
	}

	return &SpecSchedule{
		Second: second,
		Minute: minute,
		Hour:   hour,
		Dom:    dayofmonth,
		Month:  month,
		Dow:    dayofweek,
	}, nil
}

func normalizeFields(fields []string, options ParseOption) ([]string, error) {
	// Figure out how many fields we need
	max := 0
	for _, place := range places {
		if options&place > 0 {
			max++
		}
	}

	// Validate number of fields
	if count := len(fields); count < 1 || count > max {
		return nil, fmt.Errorf("expected %d to %d fields, found %d: %s", 1, max, count, fields)
	}

	// Populate all fields not part of options with their defaults
	n := 0
	expandedFields := make([]string, len(places))
	copy(expandedFields, defaults)
	for i, place := range places {
		if options&place > 0 {
			expandedFields[i] = fields[n]
			n++
		}
	}
	return expandedFields, nil
}

func getField(field string, r bounds) (uint64, error) {
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		bit, err := getRange(expr, r)
		if err != nil {
			return bits, err
		}
		bits |= bit
	}
	return bits, nil
}

func getRange(expr string, r bounds) (uint64, error) {
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
		err              error
	)

	var extra uint64
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra = starBit
	} else {
		start, err = parseIntOrName(lowAndHigh[0], r.names)
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end, err = parseIntOrName(lowAndHigh[1], r.names)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = mustParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}

		if singleDigit {
			end = r.max
		}
		if step > 1 {
			extra = 0
		}
	default:
		return 0, fmt.Errorf("too many slashes: %s", expr)
	}

	if start < r.min {
		return 0, fmt.Errorf("beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		return 0, fmt.Errorf("end of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		return 0, fmt.Errorf("beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("step of range should be a positive number: %s", expr)
	}

	return getBits(start, end, step) | extra, nil
}

func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt, nil
		}
	}
	return mustParseInt(expr)
}

func mustParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num), nil
}

func getBits(min, max, step uint) uint64 {
	var bits uint64

	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}
