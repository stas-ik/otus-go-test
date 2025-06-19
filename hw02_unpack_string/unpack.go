package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

// Unpack распаковывает строку, повторяя символы по указанной цифре и поддерживая экранирование.
func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	if unicode.IsDigit([]rune(input)[0]) {
		return "", ErrInvalidString
	}

	var builder strings.Builder
	runes := []rune(input)
	escape := false

	for runeKey, r := range runes {
		if escape {
			count := 1
			if runeKey+1 < len(runes) && unicode.IsDigit(runes[runeKey+1]) {
				countStr := string(runes[runeKey+1])
				_, err := strconv.Atoi(countStr)
				if err != nil {
					return "", ErrInvalidString
				}
				if runeKey+2 < len(runes) && unicode.IsDigit(runes[runeKey+2]) {
					return "", ErrInvalidString
				}
				runeKey++
			}
			builder.WriteString(strings.Repeat(string(r), count))
			escape = false
			continue
		}

		if r == '\\' {
			escape = true
			continue
		}

		if unicode.IsDigit(r) {
			continue
		}

		if runeKey+1 < len(runes) {
			nextRune := runes[runeKey+1]
			if unicode.IsDigit(nextRune) && !escape {
				countStr := string(nextRune)
				count, err := strconv.Atoi(countStr)
				if err != nil {
					return "", ErrInvalidString
				}
				if runeKey+2 < len(runes) && unicode.IsDigit(runes[runeKey+2]) {
					return "", ErrInvalidString
				}
				builder.WriteString(strings.Repeat(string(r), count))
			} else if !escape {
				builder.WriteString(string(r))
			}
		} else if !escape {
			builder.WriteString(string(r))
		}
	}

	if escape {
		return "", ErrInvalidString
	}

	return builder.String(), nil
}
