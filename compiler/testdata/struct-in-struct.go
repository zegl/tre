package main

import "external"

type Book struct {
	book_id int
}

type Bookshelf struct {
	book0 Book
	book1 Book
}

func main() {
	var shelf Bookshelf
	var b0 Book

	shelf.book0 = b0
	shelf.book0.book_id = 1000

	shelf.book1.book_id = 2000

	// 1000
	external.Printf("%d\n", shelf.book0.book_id)

	// 2000
	external.Printf("%d\n", shelf.book1.book_id)
}
