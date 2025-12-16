package utils

import "strconv"

func IntToString(v int64) string {
    return strconv.FormatInt(v, 10)
}
