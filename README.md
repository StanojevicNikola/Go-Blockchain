# goBlockchain


## Project summary
+ Decentralized network of nodes with following protocols that simulate blockchain
+ Block data structure, primarily intended for storing info about transaction between nodes, but with flexibility for storing any data in future upgrades
+ Decentralized nodes network, based on simple P2P protocols and followed by standard consensus protocols for resolving conflicts (raft, paxos, tendermint)
+ Standard blockchain functionalities: mining new blocks, analysis and verification of blocks using Proof of Work and/or Proof of Stake algorithm, storing chain to disk
+ Performing transactions between nodes, using Merkle trees for protection, also implementing wallets to keep track of nodes balances, transaction logs

## Usage
+ run this command from terminal: `./Node {port}`

## Programming language
+ Go

## Non-standard libraries
+ `crypto/rsa` `crypto/x509` - encoded keys and certificates
+ `crypto/sha256` - hashing and validation algorithms
+ `github.com/gorilla/mux` - http server functionalities
+ `vue.js` - JavaScript framework for GUI

## Team 
+ Nikola Stanojević
+ Mateja Trtica
+ Nikoleta Vukajlović
