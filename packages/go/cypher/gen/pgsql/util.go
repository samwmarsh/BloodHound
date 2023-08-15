package pgsql

import (
	"io"
	"strconv"
	"strings"
)

func JoinUint[T uint | uint8 | uint16 | uint32 | uint64](values []T, separator string) string {
	builder := strings.Builder{}

	for idx := 0; idx < len(values); idx++ {
		if idx > 0 {
			builder.WriteString(separator)
		}

		builder.WriteString(strconv.FormatUint(uint64(values[idx]), 10))
	}

	return builder.String()
}

func JoinInt[T int | int8 | int16 | int32 | int64](values []T, separator string) string {
	builder := strings.Builder{}

	for idx := 0; idx < len(values); idx++ {
		if idx > 0 {
			builder.WriteString(separator)
		}

		builder.WriteString(strconv.FormatInt(int64(values[idx]), 10))
	}

	return builder.String()
}

func WriteStrings(writer io.Writer, strings ...string) (int, error) {
	totalBytesWritten := 0

	for idx := 0; idx < len(strings); idx++ {
		if bytesWritten, err := io.WriteString(writer, strings[idx]); err != nil {
			return totalBytesWritten, err
		} else {
			totalBytesWritten += bytesWritten
		}
	}

	return totalBytesWritten, nil
}
