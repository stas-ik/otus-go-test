package hw02unpackstring

import (
	"errors"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	if unicode.IsDigit([]rune(input)[0]) {
		return "", ErrInvalidString
	}

	result := ""
	runes := []rune(input)
	escape := false

	for runeKey, r := range runes {
		if escape {
			count := 1
			if runeKey+1 < len(runes) && unicode.IsDigit(runes[runeKey+1]) {
				count = int(runes[runeKey+1] - '0')
				if runeKey+2 < len(runes) && unicode.IsDigit(runes[runeKey+2]) {
					return "", ErrInvalidString
				}
				runeKey++
			}
			for i := 1; i <= count; i++ {
				result += string(r)
			}
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
				number := int(nextRune - '0')
				if runeKey+2 < len(runes) && unicode.IsDigit(runes[runeKey+2]) {
					return "", ErrInvalidString
				}
				for i := 1; i <= number; i++ {
					result += string(r)
				}
			} else if !escape {
				result += string(r)
			}
		}

		if runeKey == len(runes)-1 && !escape {
			result += string(r)
		}
	}

	if escape {
		return "", ErrInvalidString
	}

	return result, nil
}
