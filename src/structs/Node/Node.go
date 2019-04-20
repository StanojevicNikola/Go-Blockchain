package main

import (
	"../Block"
	"../Blockchain"
	"../Transaction"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
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
		req,_ := http.NewRequest("POST", "/receive-new-block", bytes.NewBuffer(dataJson))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err!=nil{
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Print(string(body))

		fmt.Fprint(writer, bitcoin)

	})

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

		var nodes RequestData

		err := json.NewDecoder(request.Body).Decode(&nodes)
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		for _, node := range nodes.NewNodesUrls {

			var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(node)
			var notCurrentNode = bitcoin.CurrentNodeUrl != node

			if(notCurrentNode && nodeNotAlreadyPresent) {
				bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, node)
				writer.Write([]byte("New block registered successfully."))
			}else {
				writer.Write([]byte("New block rejected."))
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

		var nodeNotAlreadyPresent = bitcoin.NodeNotPresent(newNode.NewNodeUrl)
		var notCurrentNode = bitcoin.CurrentNodeUrl != newNode.NewNodeUrl

		if(notCurrentNode && nodeNotAlreadyPresent) {
			bitcoin.NetworkNodes = append(bitcoin.NetworkNodes, newNode.NewNodeUrl)
			writer.Write([]byte("New block registered successfully."))
		}else {
			writer.Write([]byte("New block rejected."))
		}

		for _, node := range bitcoin.NetworkNodes {

			dataJson, err := json.Marshal(&newNode)
			if err!=nil{
				panic(err)
			}
			req,_ := http.NewRequest("POST", node + "/register-node", bytes.NewBuffer(dataJson))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err!=nil{
				panic(err)
			}
			defer resp.Body.Close()
			writer.Write([]byte("New block registered successfully."))
		}



	}).Methods("POST")

	/*

	app.post('/register-and-broadcast-node', function(req, res){

		const newNodeUrl = req.body.newNodeUrl;
		if(bitcoin.networkNodes.indexOf(newNodeUrl) == -1)
			bitcoin.networkNodes.push(newNodeUrl);

		const regNodesPromises = [];
		bitcoin.networkNodes.forEach(networkNodeUrl => {
			const requestOptions = {
			uri: networkNodeUrl + '/register-node',
			method: 'POST',
			body: {newNodeUrl: newNodeUrl},
			json: true
			};
			regNodesPromises.push(rp(requestOptions));
		});

		Promise.all(regNodesPromises)
		.then(data => {
			const bulkRegisterOptions = {
			uri: newNodeUrl+'/register-nodes-bulk',
			method: 'POST',
			body: {allNetworkNodes: [...bitcoin.networkNodes, bitcoin.currentNodeUrl]},
	json:true
	};
	return rp(bulkRegisterOptions);
	})
	.then(data => {
	res.json({note: "New node registered with network successfully"});
	})

	});

*/

	r.HandleFunc("/check", func(writer http.ResponseWriter, request *http.Request) {

		bitcoin.LoadData()
		fmt.Fprint(writer, bitcoin)

	}).Methods("GET")

	
	http.ListenAndServe(":8000", r)

}
