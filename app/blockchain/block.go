package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Index     int
	Timestamp int64
	Key       []byte
	Data      []byte
	Hash      []byte
	PrevHash  []byte
	Nonce     int
}

func CreateBlock(index int, data []byte, prevHash []byte, key []byte) (*Block, bool) {
	genesisBlock := Block{
		Index:     index,
		Timestamp: time.Now().Unix(),
		Data:      []byte(data),
		PrevHash:  prevHash,
	}
	pow := NewProofOfWork(&genesisBlock)
	nonce, hash := pow.Run()
	genesisBlock.Hash = hash
	genesisBlock.Nonce = nonce
	genesisBlock.Key = key

	return &genesisBlock, pow.Validate()
}

func CreateBlockGenesis(index int, data []byte, prevHash []byte) (*Block, bool) {
	genesisBlock := Block{
		Index:     index,
		Timestamp: time.Now().Unix(),
		Data:      []byte(data),
		PrevHash:  prevHash,
	}
	pow := NewProofOfWork(&genesisBlock)
	nonce, hash := pow.Run()
	genesisBlock.Hash = hash
	genesisBlock.Nonce = nonce
	genesisBlock.Key = hash

	return &genesisBlock, pow.Validate()
}

func Genesis() (*Block, bool) {
	// Genesis block is the first block in the blockchain
	// It has no previous hash and is created with a specific data
	return CreateBlockGenesis(0, []byte("Genesis Block"), []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
