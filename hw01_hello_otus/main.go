package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	fmt.Println(reverse.String("Hello, OTUS!"))

	s := { {1,2,3}, {4,5,6,7,8}, {9,10,11,12,13,14,15} }
	Concat(s)

}

func Concat(slises ...[]int) []int {
	var length int
	for _, v := range slices {
		length += len(v)
	}
	sl := make([]int, 0, length)
	return sl
}


//сделать через copy
