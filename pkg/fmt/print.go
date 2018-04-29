package fmt

import "external"

func Println(a string) {
	external.printf("%s\n", a)
}

func Printf(a) {

}
