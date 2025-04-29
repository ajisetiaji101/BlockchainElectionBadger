package blockchain

import (
	"bytes"
	"fmt"
	"myapp/app/repository"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

const (
	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Blocks   []Block
	Election *Election
	mu       sync.Mutex
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

//addblock

func (bc *Blockchain) SetGenesisBlock(dbConn *badger.DB) bool {

	var isGenesis bool
	// initial err
	var err error

	isGenesis = true

	err = dbConn.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")

			genesisBlock, pow := Genesis()

			fmt.Println("pow hasil --------------------------", pow)

			if pow {
				fmt.Println("Added genesis block:", genesisBlock.Hash)
				fmt.Println("Genesis block is valid")
				isGenesis = true

				fmt.Println("Genesis proved")
				err = txn.Set(genesisBlock.Hash, genesisBlock.Serialize())
				Handle(err)
				err = txn.Set([]byte("lh"), genesisBlock.Hash)
			} else {
				fmt.Println("Failed to validate genesis block")
				isGenesis = false
			}

		}

		if isGenesis {
			fmt.Println("Genesis block already exists")
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			bc.LastHash, err = item.ValueCopy(nil)

			fmt.Println("Last Hash:", bc.LastHash)
		} else {
			fmt.Println("Genesis block not created")
		}

		Handle(err)

		return err
	})

	fmt.Println("Blockchain initialized successfully")

	return isGenesis
}

func (chain *Blockchain) AddBlock(data []byte, dbConn *badger.DB, key string) {

	LastHash, err := repository.GetLastHash(dbConn)

	fmt.Println("data lastHash:", LastHash)

	item, err := repository.GetDataRiwayat(dbConn, LastHash)

	fmt.Println("item riwayat:", item)

	//convert item to dataBlock
	block := Deserialize(item)

	index := block.Index + 1

	Handle(err)

	newBlock, validatePow := CreateBlock(index, data, LastHash)

	if validatePow {
		fmt.Println("Block baru valid")
		fmt.Println("Block baru:", newBlock)
	} else {
		fmt.Println("Block baru tidak valid")
		return
	}

	repository.SetDataRiwayat(dbConn, []byte(key), newBlock.Serialize())

	repository.SetLastHash(dbConn, []byte(key))

	chain.LastHash = []byte(key)

	Handle(err)
}

func (chain *Blockchain) getDataByTes(dbConn *badger.DB) {
	err := dbConn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		blockData, err := item.ValueCopy(nil)

		block := Deserialize(blockData)

		fmt.Println("Last Hash:", lastHash)
		fmt.Println("Block Data:", block.Data)

		return err
	})

	Handle(err)
}

func (chain *Blockchain) Iterator(dbConn *badger.DB) *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, dbConn}

	fmt.Println("Iter Last Hash:", iter.CurrentHash)

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)

		fmt.Println("Iter Current Hash:", item)

		Handle(err)
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}

// func (bc *Blockchain) AddBlock(newBlock Block) bool {
// 	bc.mu.Lock()
// 	defer bc.mu.Unlock()

// 	for _, block := range bc.Blocks {
// 		if bytes.Equal(block.Hash, newBlock.Hash) {
// 			return false
// 		}
// 	}
// 	bc.Blocks = append(bc.Blocks, newBlock)
// 	return true
// }

func (bc *Blockchain) IsValid() bool {
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		if !bytes.Equal(currentBlock.PrevHash, prevBlock.Hash) {
			return false
		}

		pow := NewProofOfWork(&currentBlock)
		if !pow.Validate() {
			return false
		}
	}
	return true
}

func (bc *Blockchain) SyncWithPeer(peerBlocks []Block, election *Election) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(peerBlocks) > len(bc.Blocks) {
		bc.Blocks = peerBlocks
		bc.Election = election
	}
}

// fungsi display blockchain untuk menampilkan blok-blok yang ada di blockchain
func (bc *Blockchain) Display() {
	for _, block := range bc.Blocks {
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %v\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println()
	}
}

// fungsi display blockhain returns block yang ada di blockchain
func (bc *Blockchain) GetBlocks() []Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.Blocks
}
