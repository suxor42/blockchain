package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"time"
)

func (block *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Id)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func (blockchain *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := blockchain.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	blockchain.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		blockchain.tip = newBlock.Hash
		return nil
	})
}

func (blockchain *Blockchain) Iterator() *Iterator {
	return &Iterator{blockchain.tip, blockchain.Db}
}

func (blockchain *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTransactions []Transaction
	spentTxOutputs := make(map[string][]int)
	chainIterator := blockchain.Iterator()

	for {
		block := chainIterator.Next()

		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.Id)
		Outputs:
			for value, scriptPubKey := range tx.Vout {
				// Was output spent
				if spentTxOutputs[txId] != nil {
					for _, spentOut := range spentTxOutputs[txId] {
						//
						if spentOut == value {
							continue Outputs
						}
					}
				}
				if scriptPubKey.CanBeUnlockedWith(address) {
					unspentTransactions = append(unspentTransactions, *tx)
					fmt.Printf("%s\n", unspentTransactions)
				}
			}
			if !tx.IsBaseTransaction() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxId := hex.EncodeToString(in.TxId)
						spentTxOutputs[inTxId] = append(spentTxOutputs[inTxId], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTransactions
}

func (blockchain *Blockchain) FindUnspentTransactionOutputs(address string) []TxOutput{
	var unspentTransactionOutputs []TxOutput
	unspentTransactions := blockchain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				unspentTransactionOutputs = append(unspentTransactionOutputs)
			}
		}
	}
	fmt.Printf("%s\n", unspentTransactionOutputs)
	return unspentTransactionOutputs
}

func (blockchain *Blockchain) Balance(address string) int {
	balance := 0
	unspentTransactionOutputs := blockchain.FindUnspentTransactionOutputs(address)

	for _, out := range unspentTransactionOutputs {
		balance += out.Value
	}
	return balance
}

func (iterator *Iterator) Next() *Block{
	var block *Block
	err := iterator.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		serializedBlock := bucket.Get(iterator.currentHash)
		block = DeserializeBlock(serializedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	iterator.currentHash = block.PrevBlockHash
	return block
}

func (proofOfWork *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			proofOfWork.block.PrevBlockHash,
			proofOfWork.block.HashTransactions(),
			IntToHex(proofOfWork.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		}, []byte{})
	return data
}

func (proofOfWork *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", proofOfWork.block.Transactions)

	for nonce < maxNonce {
		data := proofOfWork.prepareData(nonce)
		hash = sha256.Sum256(data)
		if nonce % 1000000 == 0 {
			fmt.Printf("\r%x\t%d", hash, nonce)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(proofOfWork.target) == -1 {
			break
		} else {
			nonce++
		}

	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (proofOfWork *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := proofOfWork.prepareData(proofOfWork.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(proofOfWork.target) == -1

	return isValid
}

func NewBaseTransaction(to, data string) *Transaction{
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{subsidy, to}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetId()
	return &tx
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	proofOfWork := NewProofOfWork(block)
	nonce, hash := proofOfWork.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func NewGenisisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlockChain(address string) *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			genesis := NewGenisisBlock(NewBaseTransaction(address, coinBaseData))

			bucket, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}

			tip = genesis.Hash
		} else {
			tip = bucket.Get([]byte("l"))
		}

		return nil

	})

	return &Blockchain{tip, db}
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	proofOfWork := &ProofOfWork{block, target}
	return proofOfWork
}

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func DeserializeBlock(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}
