package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

func (tx *Transaction) SetId() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.Id = hash[:]
}

func (txin *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return txin.ScriptSig == unlockingData
}

func (txout *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return txout.ScriptPubKey == unlockingData
}

func (tx *Transaction) IsBaseTransaction() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxId) == 0 && tx.Vin[0].Vout == -1
}