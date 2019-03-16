package Blockchain

import (
	"../Block"
	"../Transaction"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Blockchain struct {
	Chain []Block.Block
	PendingTransactions []Transaction.Transaction
	CurrentNodeUrl string
}

func (b *Blockchain) CreateGenesisBlock(){

	var block Block.Block

	block = Block.Block{
		Index:             len(b.Chain) + 1,
		Timestamp:         time.Now(),
		Transactions:      b.PendingTransactions,
		Nonce:             100,
		Hash:              "0",
		PreviousBlockHash: "0",
	}

	b.Chain = append(b.Chain, block)
	b.PendingTransactions = nil

}

func (b *Blockchain) CreateNewBlock(nonce int,
	previousBlockHash string,
	hash string) (Block.Block){

	var block Block.Block

	block = Block.Block{
		Index:             len(b.Chain) + 1,
		Timestamp:         time.Now(),
		Transactions:      b.PendingTransactions,
		Nonce:             nonce,
		Hash:              hash,
		PreviousBlockHash: previousBlockHash,
	}

	b.Chain = append(b.Chain, block)
	b.PendingTransactions = nil

	return block
}

func (b *Blockchain) GetLastBlock() (*Block.Block){
	return &b.Chain[len(b.Chain)-1]
}

func (b *Blockchain) CreateNewTransaction(amount float64,
	sender string,
	recipient string,
	transactionID string) (Transaction.Transaction) {
	transaction := Transaction.Transaction{
		Amount: amount,
		Sender: sender,
		Recipient: recipient,
		TransactionID: transactionID,
	}

	return transaction
}

func (b *Blockchain) AddTransactionToPendingTransactions(transaction Transaction.Transaction) {
	b.PendingTransactions = append(b.PendingTransactions, transaction)
}

func (b *Blockchain) HashBlock(previousBlockHash string, currentBlockData string, nonce int) string{
	dataAsString := previousBlockHash + strconv.Itoa(nonce) + currentBlockData
	// staviti pravi hash
	hash := dataAsString
	return hash
}

func (b *Blockchain) ProofOfWork(previousBlockHash string, currentBlockData string) int{
	var nonce int = 0
	var hash = b.HashBlock(previousBlockHash, currentBlockData, nonce)
	for hash[0:4] != "0000"{
		nonce++
		hash = b.HashBlock(previousBlockHash, currentBlockData, nonce)
	}

	return nonce
}

func (b* Blockchain) ChainIsValid() bool{

	// da se proveri hesiranje jsona
	var validChain = true

	for i:=1;i<len(b.Chain);i++{
		var currentBlock = b.Chain[i]
		var previousBlock= b.Chain[i-1]

		currentBlockTransactions, err := json.Marshal(&currentBlock.Transactions)
		if err != nil{
			panic(err)
		}
		currentBlockIndex, err := json.Marshal(&currentBlock.Index)
		if err != nil{
			panic(err)
		}

		currentBlockData := string(currentBlockTransactions) + string(currentBlockIndex)
		currentBlockNonce := currentBlock.Nonce

		var blockHash = b.HashBlock(previousBlock.Hash, currentBlockData, currentBlockNonce)
		if blockHash[0:4] != "0000" {
			validChain = false
		}
		if currentBlock.PreviousBlockHash != previousBlock.Hash{
			validChain = false
		}
	}

	var genesisBlock = b.Chain[0]
	var correctNonce bool = genesisBlock.Nonce==100
	var correctHash bool = genesisBlock.Hash == "0"
	var correctPreviousHash bool = genesisBlock.PreviousBlockHash == "0"
	var correctTransactions bool = len(genesisBlock.Transactions)==0

	if !correctNonce || !correctPreviousHash || !correctHash || !correctTransactions {
		validChain = false;
	}

	return validChain
}


func (b *Blockchain) SaveData() {

	f, err := os.OpenFile("blockchain.json", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil{
		panic(err)
	}
	currentBlockchain, err := json.Marshal(&b)
	if err != nil{
		panic(err)
	}
	f.Write(currentBlockchain)
	f.Close()
}

func (b* Blockchain) LoadData(){
	jsonFile, err := os.OpenFile("blockchain.json", os.O_RDONLY, 0755)
	if err != nil{
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &b)
}