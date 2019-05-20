package Wallet

import (
	"../Transaction"
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)


type Wallet struct{
	PrivateKey *rsa.PrivateKey
	PublicKey *rsa.PublicKey
	NodeID string
}

func (w *Wallet) GenerateKeys(){
	Reader := rand.Reader
	privateKey, _ := rsa.GenerateKey(Reader, 1024)
	publicKey := privateKey.PublicKey
	w.PrivateKey = privateKey
	w.PublicKey = &publicKey

	pemPrivateFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var pemPrivateBlock = &pem.Block{
		Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(pemPrivateFile, pemPrivateBlock)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	pemPrivateFile.Close()
}

func (w* Wallet) SaveKeys()bool{

	if w.PublicKey != nil && w.PrivateKey != nil{

		fileName := "wallet-" + w.NodeID + ".json"
		f,err := os.OpenFile(fileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0755)
		if err != nil{
			panic(err)
			//return false
		}

		currentWallet, err := json.MarshalIndent(&w, "", "		")
		if err != nil{
			//return false
			panic(err)
		}
		f.Write(currentWallet)
		f.Close()
		return true
	}
	return true
}

func (w* Wallet) LoadKeys() bool{

	fileName := "wallet-"+ w.NodeID +".json"
	f,err := os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil{
		return false
		panic(err)
	}

	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)
	json.Unmarshal(byteValue, &w)

	return true
}

func (w *Wallet) SignTransaction(sender string, recipient string, amount float64) (string){
	currentData := sender + recipient + strconv.FormatFloat(amount,'f', 6, 64)
	hasher := sha256.New()
	hasher.Write([]byte(currentData))
	hashedData := hasher.Sum(nil)
	rand := rand.Reader
	signer, _ := rsa.SignPKCS1v15(rand, w.PrivateKey, crypto.SHA256, hashedData)
	return hex.EncodeToString(signer)
}

func VerifyTransaction(transaction Transaction.Transaction) bool{
	privateKeyFile, err := os.Open("private_key.pem")

	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	pemFileInfo, _ := privateKeyFile.Stat()
	var size int64 = pemFileInfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)
	data , _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil{
		fmt.Print(err)
		os.Exit(1)
	}
	publicKey := &privateKeyImported.PublicKey

	currentData := transaction.Sender + transaction.Recipient + strconv.FormatFloat(transaction.Amount,'f', 6, 64)
	hasher := sha256.New()
	hasher.Write([]byte(currentData))
	hashedData := hasher.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashedData[:], []byte(transaction.Signature))
	if err != nil {
		fmt.Println("greska")
		return false
	}

	return true
}
