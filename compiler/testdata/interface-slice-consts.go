package main

import "external"

func main() {
	a := []interface{}{111, 222}

	// 111
	fromSlice0, ok := a[0].(int64)
	if ok {
		external.Printf("%d\n", fromSlice0)
	}

	// 222
	fromSlice1, ok := a[1].(int64)
	if ok {
		external.Printf("%d\n", fromSlice1)
	}

	// 111
	// 222
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}

	// 10
	// 11
	// 12
	// 13
	// 14
	// 15
	// 16
	// 17
	// 18
	// 19
	large := []interface{}{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	for k, v := range large {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}
}
