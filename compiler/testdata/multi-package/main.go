package main

import (
	"external"
	"sub"
)

var PackageVar string

func main() {
	external.Printf("%s\n", sub.World()) // World

	var a sub.AnotherInt
	a = 10

	external.Printf("%d\n", a.Plus5()) // 15

	sub.PackageVar = "inAnotherPkg"
	PackageVar = "thisPackageVar"

	external.Printf("%s\n", sub.GetPackageVar()) // inAnotherPkg
	external.Printf("%s\n", PackageVar) // thisPackageVar
}
