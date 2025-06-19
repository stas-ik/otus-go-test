package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	fmt.Println(reverse.String("Hello, OTUS!"))

	s := [][]int{
		{1, 2, 3},
		{4, 5, 6, 7, 8},
		{9, 10, 11, 12, 13, 14, 15},
	}
	result := Concat(s...)
	fmt.Println(result)
}

func Concat(slices ...[]int) []int {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]int, totalLen)
	pos := 0
	for _, s := range slices {
		copy(result[pos:], s)
		pos += len(s)
	}

	return result
}
