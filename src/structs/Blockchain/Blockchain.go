package Blockchain

import (
	"../Block"
	"../Transaction"
	"../Wallet"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
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
	NetworkNodes []string
	PublicKey *rsa.PublicKey
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

func (b *Blockchain) NodeNotPresent(target string) bool {
	for _, n := range b.NetworkNodes {
		if target == n {
			return false
		}
	}
	return true
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
	transactionID string,
	signature string) (Transaction.Transaction) {
	transaction := Transaction.Transaction{
		Amount: amount,
		Sender: sender,
		Recipient: recipient,
		TransactionID: transactionID,
		Signature: signature,
	}

	return transaction
}

func (b *Blockchain) AddTransactionToPendingTransactions(transaction Transaction.Transaction) {
	b.PendingTransactions = append(b.PendingTransactions, transaction)
}

func (b *Blockchain) HashBlock(previousBlockHash string, currentBlockData string, nonce int) string{

	dataAsString := previousBlockHash + strconv.Itoa(nonce) + currentBlockData
	// staviti pravi hash
	hasher := sha256.New()
	hasher.Write([]byte(dataAsString))
	hashedData := hasher.Sum(nil)
	hash := hex.EncodeToString(hashedData)

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

		currentBlockTransactions, err := json.MarshalIndent(&currentBlock.Transactions, "", "	")
		if err != nil{
			panic(err)
		}
		currentBlockIndex, err := json.MarshalIndent(&currentBlock.Index, "", "	")
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

/*
	mora da se sredi da se lanac automatski sinhronizuje kad cvor udje u mrezu
	plus, sad se desava da kad dva cvora imaju lance iste duzine ne zna se koji da se uzme
*/



//sredi da se ne pravi novi json fajl u poslednjem Node folderu, nego globalno
func (b *Blockchain) SaveData() {

	f, err := os.OpenFile("blockchain.json", os.O_WRONLY|os.O_CREATE | os.O_TRUNC, 0755)
	if err != nil{
		panic(err)
	}
	currentBlockchain, err := json.MarshalIndent(&b, "", "		")
	if err != nil{
		panic(err)
	}
	f.Write(currentBlockchain)
	f.Close()
}
//Initialize blockchain + open transactions data from a file.
func (b* Blockchain) LoadData(){
	jsonFile, err := os.OpenFile("blockchain.json", os.O_RDONLY, 0755)
	if err != nil{
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &b)
}

func (b* Blockchain) CalculateBalance(nodeID string) (float64) {
	var amountSent float64 = 0
	var amountReceived float64 = 0
	for _, element := range b.Chain{
		for _, tr := range element.Transactions{
			if tr.Recipient == nodeID{
				amountReceived = amountReceived + tr.Amount
			}else if tr.Sender == nodeID{
				amountSent = amountSent + tr.Amount
			}
		}
	}

	return amountReceived - amountSent
}

func (b* Blockchain) VerifyTransaction(transaction Transaction.Transaction) bool{

	sender_balance := b.CalculateBalance(transaction.Sender)
	valid := (sender_balance>=transaction.Amount) && Wallet.VerifyTransaction(transaction)
	return valid
}