package main

import (
	"./structs/Blockchain"
	"fmt"
)

func main(){

	bitcoin := Blockchain.Blockchain{}

	bitcoin.CreateGenesisBlock()
	bitcoin.CreateNewBlock(123, "asda", "dfgdsfg")
	bitcoin.CreateNewBlock(321, "dfgdsfg", "XXXXas")

	bitcoin.SaveData()
	fmt.Println(bitcoin)

}