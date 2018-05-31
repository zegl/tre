package main

import "external"

type myint struct {
	A int
}

func (m *myint) Yolo() {
	external.Printf("yolo = %d\n", m.A)
	m.A = 200
}

func main() {
	var abc myint
	abc.A = 100
	abc.Yolo() // yolo = 100
	abc.Yolo() // yolo = 200
	abc.Yolo() // yolo = 200
}
