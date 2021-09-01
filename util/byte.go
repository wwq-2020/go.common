package util

import (
	"fmt"
	"github.com/wwq-2020/go.common/errorsx"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	byteStrReg = regexp.MustCompile(`(\d+(\.\d+)?)(\w{0,2})`)
)

// errs
var (
	ErrInvalidByteStr = errorsx.Std("invalid")
)

// ParseByteStr ParseByteStr
func ParseByteStr(val string) (int64, error) {
	parts := byteStrReg.FindStringSubmatch(val)
	if len(parts) != 4 {
		return 0, errorsx.Trace(ErrInvalidByteStr)
	}
	leftStr := parts[1]
	rightStr := parts[3]
	left, err := strconv.ParseFloat(leftStr, 64)
	if err != nil {
		return 0, errorsx.Trace(err)
	}
	unit := 1
	switch strings.ToLower(rightStr) {
	case "b":
		unit = 1
	case "kb", "k":
		unit = 1 << 10
	case "mb", "m":
		unit = 1 << 20
	case "gb", "g":
		unit = 1 << 30
	}

	return int64(left * float64(unit)), nil
}

// ToByteStr ToByteStr
func ToByteStr(val int64) string {
	if val < 0 {
		return "0b"
	}
	left := val
	n := 0
	for {
		left = left >> 10
		if left == 0 {
			break
		}
		n++
		if n >= 3 {
			break
		}
	}
	div := 1
	for i := n; i > 0; i-- {
		div *= 1024
	}
	unit := "b"
	switch n {
	case 0:
		unit = "b"
	case 1:
		unit = "k"
	case 2:
		unit = "m"
	case 3:
		unit = "g"
	default:
		panic("unexpected")
	}
	_int, frac := math.Modf(float64(val) / float64(div))
	fracStr := strings.Split(strconv.FormatFloat(frac, 'f', 1, 64), ".")[1]
	if fracStr == "0" {
		return fmt.Sprintf("%d%s", int(_int), unit)
	}
	return fmt.Sprintf("%d.%s%s", int(_int), fracStr, unit)
}
