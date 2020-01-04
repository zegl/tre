package sub

type AnotherInt int64

func (ai AnotherInt) Plus5() int64 {
	return ai + 5
}

func World() string {
	return "World"
}
