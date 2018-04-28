package main

func main() {
	a := [][]int{}
	printf("%d\n", len(a)) // 0

	a = append(a, []int{100})
	printf("%d\n", len(a))  // 1
	printf("%d\n", a[0][0]) // 100

	a = append(a, []int{200, 201})
	printf("%d\n", len(a))  // 2
	printf("%d\n", a[0][0]) // 100
	printf("%d\n", a[1][0]) // 200
	printf("%d\n", a[1][1]) // 201
}
