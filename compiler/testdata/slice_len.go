package main

func main() {
	var arr [10]int

	for i := 0; i < 10; i = i + 1 {
		arr[i] = i
	}

	slice := arr[2:5]

	// 3
	printf("%d\n", len(slice))
}
