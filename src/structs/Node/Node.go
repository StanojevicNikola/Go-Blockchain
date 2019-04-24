package main

import (
	"../Block"
	"../Blockchain"
	"../Transaction"
	"bytes"
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
	bitcoin.CurrentNodeUrl = "http://localhost:8000"
	bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, bitcoin.CurrentNodeUrl)
	nodeID, _ := uuid.NewUUID()
	fmt.Print(nodeID)

	r := mux.NewRouter()

	r.HandleFunc("/blockchain", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.SaveData()
		s, _ := json.MarshalIndent(bitcoin, "", "	")
		fmt.Fprintf(writer, string(s))

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

		bitcoin.AddTransactionToPendingTransactions(newTransaction)

		writer.Write([]byte("New transaction received & accepted."))

	}).Methods("POST")


	r.HandleFunc("/transaction/broadcast", func(writer http.ResponseWriter, request *http.Request) {
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

	}).Methods("POST")


	r.HandleFunc("/mine", func(writer http.ResponseWriter, request *http.Request) {

		//hashiranje bloka i dodavanje u lanac
		var lastBlock = bitcoin.GetLastBlock()
		var previousBlockHash = lastBlock.Hash
		var currentBlockData = BlockData{bitcoin.PendingTransactions, lastBlock.Index + 1}
		var cbdJson, _ = json.MarshalIndent(&currentBlockData, "", "	")

		var nonce = bitcoin.ProofOfWork(previousBlockHash, string(cbdJson))
		var hash = bitcoin.HashBlock(previousBlockHash, string(cbdJson), nonce)
		bitcoin.CreateNewTransaction(12.5, "00", bitcoin.CurrentNodeUrl, "asdasdsa", "asdasd")

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

		fmt.Fprint(writer, bitcoin)

	}).Methods("GET")

	r.HandleFunc("/receive-new-block", func(writer http.ResponseWriter, request *http.Request) {

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

			writer.Write([]byte("New block received & accepted."))
		}else {
			writer.Write([]byte("New block rejected."))
		}
	}).Methods("POST")



	r.HandleFunc("/register-node", func(writer http.ResponseWriter, request *http.Request) {

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

		println("\nNew node url: " + node.NewNodeUrl)

		var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(node.NewNodeUrl)
		var notCurrentNode = bitcoin.CurrentNodeUrl != node.NewNodeUrl

		if(notCurrentNode && nodeNotAlreadyPresent) {
			bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, node.NewNodeUrl)
			writer.Write([]byte("New block registered successfully."))
		}else {
			writer.Write([]byte("New block rejected."))
		}

	}).Methods("POST")


	r.HandleFunc("/register-nodes-bulk", func(writer http.ResponseWriter, request *http.Request) {


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
			print("Lose dekodiranje")
			return
		}

		for _, node := range nodes {

			if node == ""{
				continue
			}

			println("Primljeni cvorovi: ")
			println("Node je:---" + node + "---")

			var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(node)
			var notCurrentNode = bitcoin.CurrentNodeUrl != node

			if(notCurrentNode && nodeNotAlreadyPresent) {
				bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, node)
				writer.Write([]byte("Nodes bulk registered successfully."))
			}else {
				writer.Write([]byte("Nodes bulk rejected."))
			}
		}
	}).Methods("POST")

	r.HandleFunc("/register-and-broadcast-node" , func(writer http.ResponseWriter, request *http.Request) {

		if request.Body == nil {
			http.Error(writer, "Please send a request body", 400)
			return
		}

		type RequestData struct {
			NewNodeUrl string `json: "newNodeUrl"`
		}

		var newNode RequestData

		err := json.NewDecoder(request.Body).Decode(&newNode)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		println("\nNew node url: " + newNode.NewNodeUrl)

		//registrovanje novog noda
		var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(newNode.NewNodeUrl)
		var notCurrentNode = bitcoin.CurrentNodeUrl != newNode.NewNodeUrl

		if(notCurrentNode && nodeNotAlreadyPresent) {
			bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, newNode.NewNodeUrl)
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

		/*for _, node := range bitcoin.NetworkNodes{
			nodesToBroadcast = append(nodesToBroadcast, node)
		}*/

		print("Poslati cvorovi: ")
		for _, node := range nodesToBroadcast{
			println(node)
		}
		print("Kraj liste")

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
		writer.Write([]byte("All network nodes shared successfully."))

	}).Methods("POST")


	r.HandleFunc("/check", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()
		fmt.Fprintf(writer, "%+v", bitcoin)

	}).Methods("GET")


	http.ListenAndServe(":8000", r)
}
