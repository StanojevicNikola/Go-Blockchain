package main

import (
	"../Blockchain"
	"../Transaction"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"src/github.com/gorilla/mux"
)

type BlockData struct{
	Transactions []Transaction.Transaction
	Index int
}

func main() {

	bitcoin := Blockchain.Blockchain{}
	bitcoin.CreateGenesisBlock()
	nodeID, _ := uuid.NewUUID()
	fmt.Print(nodeID)

	r := mux.NewRouter()

	r.HandleFunc("/blockchain", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.SaveData()
		fmt.Fprint(writer, bitcoin)

	}).Methods("GET")

	r.HandleFunc("/transaction", func(writer http.ResponseWriter, request *http.Request) {

		

	}).Methods("POST")

	r.HandleFunc("/mine", func(writer http.ResponseWriter, request *http.Request) {
		var lastBlock = bitcoin.GetLastBlock()
		var previousBlockHash = lastBlock.Hash
		var currentBlockData = BlockData{bitcoin.PendingTransactions, lastBlock.Index + 1}
		var cbdJson, _ = json.MarshalIndent(&currentBlockData, "", "	")

		var nonce = bitcoin.ProofOfWork(previousBlockHash, string(cbdJson))
		var hash = bitcoin.HashBlock(previousBlockHash, string(cbdJson), nonce)
		bitcoin.CreateNewTransaction(12.5, "00", bitcoin.CurrentNodeUrl, "asdasdsa", "asdasd")

		bitcoin.CreateNewBlock(nonce, previousBlockHash, hash)
		//bitcoin.Chain = append(bitcoin.Chain, newBlock)

		fmt.Fprint(writer, bitcoin)
		bitcoin.SaveData()
	})
	
	
	http.ListenAndServe(":8000", r)

}



//const lastBlock = bitcoin.getLastBlock();
//const previousBlockHash = lastBlock['hash'];
//const currentBlockData = {
//transactions:bitcoin.pendingTransactions,
//index: lastBlock['index'] + 1
//}

//const nonce = bitcoin.proofOfWork(previousBlockHash, currentBlockData);
//const hash = bitcoin.hashBlock(previousBlockHash, currentBlockData, nonce);
//bitcoin.createNewTransaction(12.5, "00", nodeAdress);
//const newBlock = bitcoin.createNewBlock(nonce, previousBlockHash, hash);
//
//const requestPromises = [];
//
//bitcoin.networkNodes.forEach(networkNodeUrl => {
//const requestOptions = {
//uri: networkNodeUrl + '/receive-new-block',
//method: 'POST',
//body: { newBlock: newBlock },
//json: true
//}
//requestPromises.push(rp(requestOptions));
//});
//
//Promise.all(requestPromises)
//.then(data => {
//
//const requestOptions = {
//uri: bitcoin.currentNodeUrl + '/transaction/broadcast',
//method: 'POST',
//body:{
//amount: 12.5,
//sender: "00",
//recipient: nodeAdress
//},
//json: true
//}
//return rp(requestOptions);
//})
//.then(data => {
//res.json({
//"note":"New block mined & broadcast successfully!",
//"block":newBlock
//});
//});