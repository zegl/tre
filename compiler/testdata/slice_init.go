package main

func main() {
	foo := []int{1, 2, 3}
	printf("%d\n", foo[0]) // 1
	printf("%d\n", foo[1]) // 2
	printf("%d\n", foo[2]) // 3

	printf("%d\n", len(foo)) // 3
	printf("%d\n", cap(foo)) // 3

	bar := []int{}
	printf("%d\n", len(bar)) // 0
	printf("%d\n", cap(bar)) // 2
}
