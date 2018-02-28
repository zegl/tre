package main

func main() {
    // 3 - foo
    foo := "foo"
    printf("%d - %s\n", len(foo), foo)

    // 10 - fooyollllo
    foo = foo + "yollllo"
    printf("%d - %s\n", len(foo), foo)

    // 16 - fooyolllloh11100
    foo = foo + "h11100"
    printf("%d - %s\n", len(foo), foo)

    // abbcccdddd
    printf("%s\n", "a" + "bb" + "ccc" + "dddd")
}
