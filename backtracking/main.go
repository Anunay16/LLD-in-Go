package main

import (
	"fmt"
	"practice/backtracking/permutations"
)

func main() {
	permutations := permutations.Permutations([]int{1, 2, 3})
	fmt.Println(permutations)
}
