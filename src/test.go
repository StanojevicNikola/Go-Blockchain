package main

import (
	"./structs/Wallet"
	"fmt"
)

func main(){


	w := Wallet.Wallet{}
	w.GenerateKeys()
	signed := w.SignTransaction("Ninna", "Trta", 200)
	fmt.Print(signed)
}