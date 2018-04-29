package main

import "external"

func main() {
	var arr []int

	external.Printf("%d\n", len(arr)) // 0
	external.Printf("%d\n", cap(arr)) // 2

	arr = append(arr, 100)

	external.Printf("%d\n", len(arr)) // 1
	external.Printf("%d\n", cap(arr)) // 2

	external.Printf("%d\n", arr[0]) // 100

	arr = append(arr, 200)

	external.Printf("%d\n", len(arr)) // 2
	external.Printf("%d\n", cap(arr)) // 2

	external.Printf("%d\n", arr[0]) // 100
	external.Printf("%d\n", arr[1]) // 200

	arr = append(arr, 300)

	external.Printf("%d\n", len(arr)) // 3
	external.Printf("%d\n", cap(arr)) // 4

	external.Printf("%d\n", arr[0]) // 100
	external.Printf("%d\n", arr[1]) // 200
	external.Printf("%d\n", arr[2]) // 300

	arr = append(arr, 400)

	external.Printf("%d\n", len(arr)) // 4
	external.Printf("%d\n", cap(arr)) // 4

	external.Printf("%d\n", arr[0]) // 100
	external.Printf("%d\n", arr[1]) // 200
	external.Printf("%d\n", arr[2]) // 300
	external.Printf("%d\n", arr[3]) // 400

	arr = append(arr, 500)

	external.Printf("%d\n", len(arr)) // 5
	external.Printf("%d\n", cap(arr)) // 8

	external.Printf("%d\n", arr[0]) // 100
	external.Printf("%d\n", arr[1]) // 200
	external.Printf("%d\n", arr[2]) // 300
	external.Printf("%d\n", arr[3]) // 400
	external.Printf("%d\n", arr[4]) // 500

	arr = append(arr, 600)
	arr = append(arr, 700)
	arr = append(arr, 800)
	arr = append(arr, 900)
	arr = append(arr, 1000)

	external.Printf("%d\n", len(arr)) // 10
	external.Printf("%d\n", cap(arr)) // 16

	external.Printf("%d\n", arr[0]) // 100
	external.Printf("%d\n", arr[1]) // 200
	external.Printf("%d\n", arr[2]) // 300
	external.Printf("%d\n", arr[3]) // 400
	external.Printf("%d\n", arr[4]) // 500
	external.Printf("%d\n", arr[5]) // 600
	external.Printf("%d\n", arr[6]) // 700
	external.Printf("%d\n", arr[7]) // 800
	external.Printf("%d\n", arr[8]) // 900
	external.Printf("%d\n", arr[9]) // 1000
}
