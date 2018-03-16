package blockchain


const subsidy = 10

type Transaction struct {
	Id []byte
	Vin []TxInput
	Vout []TxOutput
}

type TxInput struct {
	TxId []byte
	Vout int
	ScriptSig string
}

type TxOutput struct {
	Value int
	ScriptPubKey string
}
