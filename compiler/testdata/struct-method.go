// yoloyolo = 100
// yoloyolo = 100
// witharg = 500

package main

import "external"

type myint struct {
	A int
}

func (m myint) Yolo() {
	external.Printf("yoloyolo = %d\n", m.A)
	m.A = 200
}

func (m myint) WithArg(arg int) {
	external.Printf("witharg = %d\n", arg)
}

func main() {
	var abc myint
	abc.A = 100
	abc.Yolo()
	abc.Yolo()
	abc.WithArg(500)
}
