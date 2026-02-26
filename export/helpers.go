package export

import (
	"strconv"
	"strings"
)

func parseAmount(value string) (float64, error) {
	value = strings.TrimSpace(value)
	return strconv.ParseFloat(value, 64)
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
