package main

import (
	"../Block"
	"../Blockchain"
	"../Transaction"
	"../Wallet"
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"src/github.com/gorilla/mux"
)


type BlockData struct{
	Transactions []Transaction.Transaction
	Index int
}

func main() {

	//bitcoin nastaje POSTovanjem novog walleta

	var port = os.Args[1]
	Wallet := Wallet.Wallet{NodeID:port}
	bitcoin := Blockchain.Blockchain{}
	println("Listening on port " + port + "...")
	r := mux.NewRouter()





	r.HandleFunc("/blockchain", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()
		s, _ := json.MarshalIndent(bitcoin, "", "	")


		writer.Write(s)


	}).Methods("GET")


	r.HandleFunc("/transaction", func(writer http.ResponseWriter, request *http.Request) {

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		var newTransaction Transaction.Transaction
		err := json.NewDecoder(request.Body).Decode(&newTransaction)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		var transactionToAdd Transaction.Transaction

		transactionToAdd.Amount = newTransaction.Amount
		transactionToAdd.Recipient = newTransaction.Recipient
		transactionToAdd.Sender = newTransaction.Sender
		transactionToAdd.TransactionID = hex.EncodeToString([]byte(string(rand.Intn(10000))))
		transactionToAdd.Signature = Wallet.SignTransaction(transactionToAdd.Sender, transactionToAdd.Recipient, transactionToAdd.Amount)

		bitcoin.AddTransactionToPendingTransactions(transactionToAdd)
		bitcoin.SaveData()

		type SendData struct {
			Message string
			Funds float64
		}

		dataToSend := SendData{"New transaction received & accepted.",
							bitcoin.CalculateBalance(bitcoin.CurrentNodeUrl)	}
		s, _ := json.MarshalIndent(dataToSend, "", "	")
		writer.Write(s)

	}).Methods("POST")


	r.HandleFunc("/transaction/broadcast", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		var newTransaction Transaction.Transaction
		err := json.NewDecoder(request.Body).Decode(&newTransaction)

		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		dataJson, err := json.Marshal(&newTransaction)
		if err != nil {
			panic(err)
		}

		for _, node := range bitcoin.NetworkNodes{

			req, _ := http.NewRequest("POST", node + "/transaction", bytes.NewBuffer(dataJson))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
		}
		writer.Write([]byte("Transaction created and broadcasted successfully"))

		bitcoin.SaveData()

	}).Methods("POST")


	r.HandleFunc("/mine", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		//hashiranje bloka i dodavanje u lanac
		var lastBlock = bitcoin.GetLastBlock()
		var previousBlockHash = lastBlock.Hash
		var currentBlockData = BlockData{bitcoin.PendingTransactions, lastBlock.Index + 1}
		var cbdJson, _ = json.MarshalIndent(&currentBlockData, "", "	")

		var nonce = bitcoin.ProofOfWork(previousBlockHash, string(cbdJson))
		var hash = bitcoin.HashBlock(previousBlockHash, string(cbdJson), nonce)
		newTransaction := bitcoin.CreateNewTransaction(12.5, "00", bitcoin.CurrentNodeUrl, "asdasdsa", "asdasd")
		bitcoin.AddTransactionToPendingTransactions(newTransaction)
		//

		newBlock := bitcoin.CreateNewBlock(nonce, previousBlockHash, hash)
		bitcoin.SaveData()

		//broadcastovanje bloka svima u lancu
		//pogodimo receive-new-block endpoint

		dataJson, err := json.Marshal(&newBlock)
		if err!=nil{
			panic(err)
		}



		for _, node:= range bitcoin.NetworkNodes {
			req, _ := http.NewRequest("POST", node + "/receive-new-block", bytes.NewBuffer(dataJson))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
		}

		bitcoin.SaveData()

		type SendData struct {
			Message string
			Funds float64
		}

		dataToSend := SendData{"New block mined succesfully.",
			bitcoin.CalculateBalance(bitcoin.CurrentNodeUrl)}

		bitcoin.SaveData()
		s, _ := json.MarshalIndent(dataToSend, "", "	")
		writer.Write(s)


	}).Methods("GET")

	r.HandleFunc("/receive-new-block", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		var newBlock Block.Block
		err := json.NewDecoder(request.Body).Decode(&newBlock)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		lastBlock := bitcoin.GetLastBlock()
		correctHash := lastBlock.Hash ==newBlock.PreviousBlockHash
		correctIndex := lastBlock.Index+1 == newBlock.Index

		if correctHash&&correctIndex{
			bitcoin.Chain = append(bitcoin.Chain, newBlock)
			bitcoin.PendingTransactions = nil
			bitcoin.SaveData()
			writer.Write([]byte("New block received & accepted."))
		}else {
			writer.Write([]byte("New block rejected."))
		}
	}).Methods("POST")



	r.HandleFunc("/register-node", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		type RequestData struct {
			NewNodeUrl string `json: "newNodeUrl"`
		}

		var node RequestData

		err := json.NewDecoder(request.Body).Decode(&node)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}




		var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(node.NewNodeUrl)
		var notCurrentNode = bitcoin.CurrentNodeUrl != node.NewNodeUrl

		if(notCurrentNode && nodeNotAlreadyPresent) {
			bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, node.NewNodeUrl)
			bitcoin.SaveData()
			writer.Write([]byte("New block registered successfully."))
		}else {
			writer.Write([]byte("New block rejected."))
		}

	}).Methods("POST")


	r.HandleFunc("/register-nodes-bulk", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		type RequestData struct {
			NewNodesUrls []string `json: "newNodesUrls"`
		}

		nodes := make([]string, 100)

		err := json.NewDecoder(request.Body).Decode(&nodes)

		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		for _, node := range nodes {

			if node == ""{
				continue
			}

			var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(node)
			var notCurrentNode = bitcoin.CurrentNodeUrl != node

			if(notCurrentNode && nodeNotAlreadyPresent) {
				bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, node)
				bitcoin.SaveData()
				writer.Write([]byte("Nodes bulk registered successfully."))
			}else {
				writer.Write([]byte("Nodes bulk rejected."))
			}
		}
	}).Methods("POST")

	r.HandleFunc("/register-and-broadcast-node" , func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		type RequestData struct {
			NewNodeUrl string
		}

		var newNode RequestData

		err := json.NewDecoder(request.Body).Decode(&newNode)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		//registrovanje novog noda
		var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(newNode.NewNodeUrl)
		var notCurrentNode = bitcoin.CurrentNodeUrl != newNode.NewNodeUrl

		if(notCurrentNode && nodeNotAlreadyPresent) {
			bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, newNode.NewNodeUrl)
			bitcoin.SaveData()
			writer.Write([]byte("New block registered successfully."))
		}else {
			writer.Write([]byte("New block rejected."))
		}


		//broadcastovanje novog noda svim nodovima u mrezi
		for _, node := range bitcoin.NetworkNodes {

			if node != bitcoin.CurrentNodeUrl {
				dataJson, err := json.Marshal(&newNode)
				if err != nil {
					panic(err)
				}
				req, _ := http.NewRequest("POST", node+"/register-node", bytes.NewBuffer(dataJson))
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				writer.Write([]byte("New node broadcasted successfully."))
			}
		}


		// sve cvorove iz mreze saljemo novom cvoru
		nodesToBroadcast := bitcoin.NetworkNodes[:]

		dataJson, err := json.Marshal(&nodesToBroadcast)
		if err != nil{
			panic(err)
		}

		req, _ := http.NewRequest("POST", newNode.NewNodeUrl + "/register-nodes-bulk", bytes.NewBuffer(dataJson))
		client := &http.Client{}
		resp,err := client.Do(req)
		if err != nil{
			panic(err)
		}
		defer resp.Body.Close()

		bitcoin.SaveData()

		writer.Write([]byte("All network nodes shared successfully."))

	}).Methods("POST")


	r.HandleFunc("/check", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()
		s, _ := json.MarshalIndent(bitcoin, "", "	")
		fmt.Fprintf(writer, "%+v", string(s))

	}).Methods("GET")


	r.HandleFunc("/consensus", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()

		winnerChain := bitcoin.Chain
		var replace = false
		maxChainLen := len(bitcoin.Chain)

		NewPendingTransactions := []Transaction.Transaction{}

		for _, node := range bitcoin.NetworkNodes{
			if node == bitcoin.CurrentNodeUrl{
				continue
			}

			//dohvatanje lanaca svih cvorova iz mreze
			req, _ := http.NewRequest("GET", node + "/blockchain", nil)
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil{
				panic(err)
			}
			defer resp.Body.Close()
			var newBlockchain Blockchain.Blockchain
			err = json.NewDecoder(resp.Body).Decode(&newBlockchain)
			if err != nil {
				http.Error(writer, err.Error(), 400)
				return
			}

			//provera da li je dobr poslat lanac
			//newEncoded, err := json.MarshalIndent(newBlockchain, "", "	")
			//if err != nil{
			//	panic(err)
			//}
			//writer.Write(newEncoded)

			//provera duzina i eventualno menjanje lanca
			currentChainLen := len(newBlockchain.Chain)

			if currentChainLen>maxChainLen{
				winnerChain = newBlockchain.Chain
				maxChainLen = currentChainLen
				NewPendingTransactions = newBlockchain.PendingTransactions
				replace = true
			}
		}

		if replace == true{
			bitcoin.Chain = winnerChain
			bitcoin.PendingTransactions = NewPendingTransactions
			bitcoin.SaveData()
			writer.Write([]byte("Chain was replaced!"))
		}else{
			writer.Write([]byte("Chain was held!"))
		}
		bitcoin.SaveData()

	}).Methods("GET")


	//ucitavamo public i private key i to nam omogucava da inicijalizujemo blokchain koji je prethodno sacuvan

	r.HandleFunc("/wallet", func(writer http.ResponseWriter, request *http.Request) {

		if !Wallet.LoadKeys(){
			log.Fatal("Loading wallet failed!")
		}
		//bitcoin = Blockchain.Blockchain{}
		//bitcoin.CreateGenesisBlock()

		//bitcoin.LoadData()
		bitcoin.PublicKey = Wallet.PublicKey
		bitcoin.CurrentNodeUrl = "http://localhost:" + port
		bitcoin.NetworkNodes = []string{"http://localhost:" + port}

		var balance = bitcoin.CalculateBalance(bitcoin.CurrentNodeUrl)


		type SendData struct {
			PrivateKey *rsa.PrivateKey
			PublicKey *rsa.PublicKey
			Funds float64
		}

		toSend := SendData{Wallet.PrivateKey, Wallet.PublicKey, balance}
		s, _ := json.MarshalIndent(toSend, "", "	")
		writer.Write(s)

	}).Methods("GET")


	//kad nemamo nikakav blockchain i krecemo od nule, pravimo wallet i njemu dodeljujemo blockchain
	r.HandleFunc("/wallet", func(writer http.ResponseWriter, request *http.Request) {

		Wallet.GenerateKeys()

		if Wallet.SaveKeys(){
				bitcoin = Blockchain.Blockchain{
												PublicKey:Wallet.PublicKey,
												CurrentNodeUrl: "http://localhost:" + port,
												NetworkNodes: []string{"http://localhost:" + port},
												}
			bitcoin.CreateGenesisBlock()

			type SendData struct {
				PrivateKey *rsa.PrivateKey
				PublicKey *rsa.PublicKey
				Funds float64
			}
			balance := 0.0

			bitcoin.SaveData()

			toSend := SendData{Wallet.PrivateKey, Wallet.PublicKey, balance}
			s, _ := json.MarshalIndent(toSend, "", "	")
			writer.Write(s)

		}else{
			writer.Write([]byte("Saving keys failed"))
		}

	}).Methods("POST")


	r.HandleFunc("/network", func(writer http.ResponseWriter, request *http.Request) {

		http.ServeFile(writer, request, "../../ui/network.html")

	}).Methods("GET")

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "../../ui/node.html")
		//writer.Write(curr)
	})


	http.ListenAndServe(":"+port, r)
}
