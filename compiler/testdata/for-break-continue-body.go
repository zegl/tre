package main

import "external"

// 1
// 0
// 1
// 2111
// 9
// 31
// 8
// 7

func firstOver(limit int) int {
	for i := 0; i < 100; i++ {
		if (i * i) > limit {
			return i
		}
	}
	return -1
}

func retInBody() int {
	for i := 0; i < 10; i++ {
		return i + 7
	}
	return -1
}

func main() {
	n := 0
	for i := 0; i < 10; i++ {
		n = n + 1
		break
		n = n + 100
	}
	external.Printf("%d\n", n)

	m := 0
	for i := 0; i < 5; i++ {
		continue
		m = m + 1
	}
	external.Printf("%d\n", m)

	k := 0
	for _, v := range []int{1, 2, 3} {
		k = k + v
		break
	}
	external.Printf("%d\n", k)

	s := 0
	for i := 0; i < 4; i++ {
		switch i {
		case 1:
			s = s + 10
			break
		case 2:
			s = s + 100
			fallthrough
		case 3:
			s = s + 1000
		default:
			s = s + 1
		}
	}
	external.Printf("%d\n", s)

	c := 0
	for i := 0; i < 6; i++ {
		switch i {
		case 2:
			continue
		}
		if i == 4 {
			continue
		}
		c = c + i
	}
	external.Printf("%d\n", c)

	t := 0
	for i := 0; i < 10; i++ {
		if i > 2 {
			if i > 5 {
				t = t + 1000
			}
			t = t + 1
			break
		}
		t = t + 10
	}
	external.Printf("%d\n", t)

	external.Printf("%d\n", firstOver(50))
	external.Printf("%d\n", retInBody())
}
