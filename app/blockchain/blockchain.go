package blockchain

import (
	"bytes"
	"errors"
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

			Handle(err)

			fmt.Println("Last Hash:", bc.LastHash)
		} else {
			fmt.Println("Genesis block not created")
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			bc.LastHash, err = item.ValueCopy(nil)
		}

		Handle(err)

		return err
	})

	fmt.Println("Blockchain initialized successfully")

	return isGenesis
}

func (chain *Blockchain) AddBlock(data []byte, dbConn *badger.DB, key []byte, dataKey []byte) {

	LastHash, err := repository.GetLastHash(dbConn)

	fmt.Println("data lastHash:", LastHash)

	item, err := repository.GetDataRiwayat(dbConn, LastHash)

	fmt.Println("item riwayat:", item)

	//convert item to dataBlock
	block := Deserialize(item)

	index := block.Index + 1

	Handle(err)

	newBlock, validatePow := CreateBlock(index, data, block.Hash, key, dataKey)

	if validatePow {
		fmt.Println("Block baru valid")
		fmt.Println("Block baru:", newBlock)

	} else {
		fmt.Println("Block baru tidak valid")
		return
	}

	repository.SetDataRiwayat(dbConn, key, newBlock.Serialize())

	fmt.Println("SetDataRiwayat: berhasil")

	repository.SetLastHash(dbConn, key)

	fmt.Println("SetLastHash: berhasil")

	chain.LastHash = key

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

func (chain *Blockchain) GetLastHash(dbConn *badger.DB) []byte {
	var LastHash []byte

	data, err := repository.GetLastHash(dbConn)
	Handle(err)

	fmt.Println("Last Hash:", data)

	chain.LastHash = data

	LastHash = data

	return LastHash
}

func (chain *Blockchain) GetLastHashCek(dbConn *badger.DB) []byte {
	var LastHash []byte

	err := repository.GetLastHashOrCreate(dbConn)

	fmt.Println("create new lash block", LastHash)

	LastHash, err = repository.GetLastHash(dbConn)

	Handle(err)

	chain.LastHash = LastHash

	return LastHash
}

func (chain *Blockchain) GetBlockByKey(dbConn *badger.DB, dataAkhirHash []byte) [][]byte {
	var blocks [][]byte

	iter := chain.Iterator(dbConn, dataAkhirHash)

	//skip lh key
	for {
		block := iter.Next()

		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Prev. key: %x\n", block.PrevKey)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		fmt.Println("Block Data  berhasil keluar")

		blocks = append(blocks, block.Serialize())

		fmt.Println("Block Hash:", block.Key)

		fmt.Println("Panjang block PrevHash:", len(block.PrevHash))

		fmt.Println("Block PrevHash:", block.PrevHash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *Blockchain) GetBlock(blockHash []byte, dbConn *badger.DB) ([]byte, error) {
	var block []byte

	err := dbConn.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			block, err = item.ValueCopy(nil)

			Handle(err)

			// block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *Blockchain) Iterator(dbConn *badger.DB, dataAkhirHash []byte) *BlockChainIterator {
	iter := &BlockChainIterator{dataAkhirHash, dbConn}

	fmt.Println("Iter Last Hash:", iter.CurrentHash)

	return iter
}

// ambil semua block dari blockchain
func (chain *Blockchain) GetAllBlocks(dbConn *badger.DB) [][]byte {
	var blocks [][]byte

	err := dbConn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			// Skip key "lh" (last hash pointer)
			if string(k) == "lh" {
				continue
			}

			err := item.Value(func(v []byte) error {
				// Simpan value block ke slice
				blocks = append(blocks, append([]byte{}, v...))
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	Handle(err)
	return blocks
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {

		fmt.Println("Iter Current Hash yang masuk :", iter.CurrentHash)
		item, err := txn.Get(iter.CurrentHash)

		fmt.Println("Iter Current value GET:", item)

		itemdata, err := item.ValueCopy(nil)
		fmt.Println("Iter Current value:", itemdata)

		block = Deserialize(itemdata)

		fmt.Print("proses deserialize block:", block)

		Handle(err)

		// dataKey = item.KeyCopy(nil)

		return err
	})
	Handle(err)

	fmt.Println("Iter PrevKey:", block.PrevKey)

	iter.CurrentHash = block.PrevKey

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

func (bc *Blockchain) SyncWithPeer(peerBlocks [][]byte, dbConn *badger.DB) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	var LastHash []byte

	fmt.Println("Syncing with peer...")

	//length peerBlocks
	fmt.Println("Total blocks from peer:", len(peerBlocks))

	for _, blockData := range peerBlocks {

		block := Deserialize(blockData)

		fmt.Println("hasil block data :", block)

		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %v\n", block.Data)
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		fmt.Println("Block Hash:", block.Hash)
		repository.SetDataRiwayat(dbConn, block.Hash, block.Serialize())
	}

	LastHashBlock := peerBlocks[len(peerBlocks)-1]

	blockLast := Deserialize(LastHashBlock)
	fmt.Println("Block terakhir:", blockLast)

	fmt.Println("Block terakhir Hash:", blockLast.Hash)

	LastHash = blockLast.Key

	fmt.Println("Last Hash:", LastHash)

	repository.SetLastHash(dbConn, LastHash)
	bc.LastHash = LastHash

	fmt.Println("Syncing completed.")
	fmt.Println("Total blocks synced:", len(peerBlocks))

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
