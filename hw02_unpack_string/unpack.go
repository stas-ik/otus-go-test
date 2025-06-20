package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var builder strings.Builder
	runes := []rune(input)
	length := len(runes)

	for i := 0; i < length; {
		r := runes[i]

		if unicode.IsDigit(r) {
			return "", ErrInvalidString
		}

		if r == '\\' {
			i++
			if i >= length {
				return "", ErrInvalidString
			}
			escaped := runes[i]
			if escaped != '\\' && !unicode.IsDigit(escaped) {
				return "", ErrInvalidString
			}
			i++

			start := i
			for i < length && unicode.IsDigit(runes[i]) {
				i++
			}

			count := 1
			if start < i {
				numStr := string(runes[start:i])
				var err error
				count, err = strconv.Atoi(numStr)
				if err != nil {
					return "", ErrInvalidString
				}
			}

			builder.WriteString(strings.Repeat(string(escaped), count))
			continue
		}

		ch := r
		i++

		count := 1
		if i < length && unicode.IsDigit(runes[i]) {
			start := i
			i++
			if i < length && unicode.IsDigit(runes[i]) {
				return "", ErrInvalidString
			}
			numStr := string(runes[start:i])
			var err error
			count, err = strconv.Atoi(numStr)
			if err != nil {
				return "", ErrInvalidString
			}
		}

		builder.WriteString(strings.Repeat(string(ch), count))
	}

	return builder.String(), nil
}
