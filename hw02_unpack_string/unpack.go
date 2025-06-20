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
			escaped, newIndex, err := processEscapedChar(runes, i, length)
			if err != nil {
				return "", err
			}
			i = newIndex
			builder.WriteString(escaped)
			continue
		}

		result, newIndex, err := processRegularChar(runes, i, length)
		if err != nil {
			return "", err
		}
		i = newIndex
		builder.WriteString(result)
	}

	return builder.String(), nil
}

func processEscapedChar(runes []rune, index, length int) (string, int, error) {
	if index+1 >= length {
		return "", 0, ErrInvalidString
	}

	escaped := runes[index+1]
	if escaped != '\\' && !unicode.IsDigit(escaped) {
		return "", 0, ErrInvalidString
	}

	count, newIndex, err := parseCount(runes, index+2, length)
	if err != nil {
		return "", 0, err
	}

	return strings.Repeat(string(escaped), count), newIndex, nil
}

func processRegularChar(runes []rune, index, length int) (string, int, error) {
	ch := runes[index]
	count, newIndex, err := parseCount(runes, index+1, length)
	if err != nil {
		return "", 0, err
	}

	return strings.Repeat(string(ch), count), newIndex, nil
}

func parseCount(runes []rune, index, length int) (int, int, error) {
	if index >= length || !unicode.IsDigit(runes[index]) {
		return 1, index, nil
	}

	start := index
	index++
	if index < length && unicode.IsDigit(runes[index]) {
		return 0, 0, ErrInvalidString
	}

	count, err := strconv.Atoi(string(runes[start:index]))
	if err != nil {
		return 0, 0, ErrInvalidString
	}

	return count, index, nil
}
