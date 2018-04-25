package main

func main() {
	var arr []int

	printf("%d\n", len(arr)) // 0
	printf("%d\n", cap(arr)) // 2

	arr = append(arr, 100)

	printf("%d\n", len(arr)) // 1
	printf("%d\n", cap(arr)) // 2

	printf("%d\n", arr[0]) // 100

	arr = append(arr, 200)

	printf("%d\n", len(arr)) // 2
	printf("%d\n", cap(arr)) // 2

	printf("%d\n", arr[0]) // 100
	printf("%d\n", arr[1]) // 200

	arr = append(arr, 300)

	printf("%d\n", len(arr)) // 3
	printf("%d\n", cap(arr)) // 4

	printf("%d\n", arr[0]) // 100
	printf("%d\n", arr[1]) // 200
	printf("%d\n", arr[2]) // 300

	arr = append(arr, 400)

	printf("%d\n", len(arr)) // 4
	printf("%d\n", cap(arr)) // 4

	printf("%d\n", arr[0]) // 100
	printf("%d\n", arr[1]) // 200
	printf("%d\n", arr[2]) // 300
	printf("%d\n", arr[3]) // 400

	arr = append(arr, 500)

	printf("%d\n", len(arr)) // 5
	printf("%d\n", cap(arr)) // 8

	printf("%d\n", arr[0]) // 100
	printf("%d\n", arr[1]) // 200
	printf("%d\n", arr[2]) // 300
	printf("%d\n", arr[3]) // 400
	printf("%d\n", arr[4]) // 500

	arr = append(arr, 600)
	arr = append(arr, 700)
	arr = append(arr, 800)
	arr = append(arr, 900)
	arr = append(arr, 1000)

	printf("%d\n", len(arr)) // 10
	printf("%d\n", cap(arr)) // 16

	printf("%d\n", arr[0]) // 100
	printf("%d\n", arr[1]) // 200
	printf("%d\n", arr[2]) // 300
	printf("%d\n", arr[3]) // 400
	printf("%d\n", arr[4]) // 500
	printf("%d\n", arr[5]) // 600
	printf("%d\n", arr[6]) // 700
	printf("%d\n", arr[7]) // 800
	printf("%d\n", arr[8]) // 900
	printf("%d\n", arr[9]) // 1000
}
