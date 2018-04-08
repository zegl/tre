package main

func main() {
	var arr [10]int

	for i := 0; i < 10; i = i + 1 {
		arr[i] = i
	}

	slice := arr[2:4]

	// 2
	printf("%d\n", slice[0])

	// 3
	printf("%d\n", slice[1])
}
