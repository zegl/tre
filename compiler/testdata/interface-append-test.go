package main



package main

import "external"

func main() {
	a1 := 100
	a2 := 200
	a := []interface{}{a1, a2}

	// 100
	fromSlice0, ok := a[0].(int64)
	if ok {
		external.Printf("%d\n", fromSlice0)
	}

	// 200
	fromSlice1, ok := a[1].(int64)
	if ok {
		external.Printf("%d\n", fromSlice1)
	}

	// 100
	// 200
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}

	a = append(a, 300)

	// 100
	// 200
	// 300
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}

	var b interface{}
	b = 400
	a = append(a, b)
	// 100
	// 200
	// 300
	// 400
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}
}
