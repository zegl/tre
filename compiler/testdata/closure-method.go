package main

import "external"

type account struct {
	balance int
}

func (a *account) depositor() func(int) int {
	return func(amount int) int {
		a.balance = a.balance + amount
		return a.balance
	}
}

func main() {
	acc := &account{balance: 100}
	deposit := acc.depositor()
	deposit(50)
	external.Printf("%d\n", deposit(25)) // 175
	external.Printf("%d\n", acc.balance) // 175
}
