package Block


import (
	"time"
	"../Transaction"
)

type Block struct {
	Index int
	Timestamp time.Time
	Transactions [] Transaction.Transaction
	Nonce int
	Hash string
	PreviousBlockHash string
}
