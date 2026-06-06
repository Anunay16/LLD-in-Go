package permutations

import "slices"

func permutationsWithDupHelper(nums, psf []int, ans *[][]int, vis map[int]struct{}) {
	if len(psf) == len(nums) {
		*ans = append(*ans, append([]int{}, psf...))
		return
	}
	for i := 0; i < len(nums); i++ {
		if _, used := vis[i]; used {
			continue
		}
		// Skip a duplicate only if the previous identical element has NOT been used in the current permutation path.
		if i > 0 && nums[i] == nums[i-1] {
			if _, prevUsed := vis[i-1]; !prevUsed {
				continue
			}
		}

		vis[i] = struct{}{}
		psf = append(psf, nums[i])
		permutationsWithDupHelper(nums, psf, ans, vis)
		psf = psf[:len(psf)-1]
		delete(vis, i)
	}
}

func PermutationsWithDup(nums []int) [][]int {
	ans := [][]int{}
	psf := []int{}
	vis := make(map[int]struct{})
	slices.Sort(nums)
	permutationsWithDupHelper(nums, psf, &ans, vis)
	return ans
}
