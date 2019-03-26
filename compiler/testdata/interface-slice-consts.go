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

	a = append(a, 333)

	// 111
	// 222
	// 333
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}

	var b interface{}
	b = 444
	a = append(a, b)
	// 111
	// 222
	// 333
	// 444
	for k, v := range a {
		fromV, ok := v.(int64)
		if ok {
			external.Printf("%d\n", fromV)
		}
	}
}
