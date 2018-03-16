package blockchain

import (
	"github.com/boltdb/bolt"
	"math"
	"math/big"
)

var (
	maxNonce     = math.MaxInt64
	dbFile       = "blockchain.Db"
	blocksBucket = "blocks"
)

const targetBits = 24
const coinBaseData = "mips awesome blockchain starts now"


type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

type Iterator struct {
	currentHash []byte
	db *bolt.DB
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}