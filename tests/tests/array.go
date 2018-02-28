package main

func main() {
    var array [4]int

    array[0] = 100
    array[1] = 200
    array[2] = 300
    array[3] = 400

    // 100
    // 200
    // 300
    // 400
    printf("%d\n", array[0])
    printf("%d\n", array[1])
    printf("%d\n", array[2])
    printf("%d\n", array[3])

    // len = 4
    printf("len = %d\n", len(array))
}
