package blockchain


type Block struct {
	Timestamp		int64
	Data			[]byte
	PrevBlockHash	[]byte
	Hash			[]byte
}

type Blockchain struct {
	blocks []*Block
}