package main

import (
	"fmt"
	"practice/backtracking/permutations"
)

func main() {
	permutations1 := permutations.Permutations([]int{1, 2, 1})
	permutations2 := permutations.PermutationsWithDup([]int{1, 2, 1})
	fmt.Println(permutations1)
	fmt.Println(permutations2)
}
