package sub

type AnotherInt int64

func (ai AnotherInt) Plus5() int64 {
	return ai + 5
}

var PackageVar string

func GetPackageVar() string {
	return PackageVar
}

func World() string {
	return "World"
}
