package permutations

func permutationsHelper(nums, psf []int, ans *[][]int, vis map[int]struct{}) {
	if len(psf) == len(nums) {
		*ans = append(*ans, append([]int{}, psf...))
		return
	}
	for i := 0; i < len(nums); i++ {
		if _, ok := vis[i]; ok {
			continue
		}
		vis[i] = struct{}{}
		psf = append(psf, nums[i])
		permutationsHelper(nums, psf, ans, vis)
		psf = psf[:len(psf)-1]
		delete(vis, i)
	}
}

func Permutations(nums []int) [][]int {
	ans := [][]int{}
	psf := []int{}
	vis := make(map[int]struct{})
	permutationsHelper(nums, psf, &ans, vis)
	return ans
}
