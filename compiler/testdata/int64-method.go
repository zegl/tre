// yoloyolo = 100

package main

import "external"

type myint int64

func (m myint) Yolo() {
	external.Printf("yoloyolo = %d\n", m)
}

func main() {
	var abc myint
	abc = 100
	abc.Yolo()

	f1 := func() int {
		return 100
	}
}
