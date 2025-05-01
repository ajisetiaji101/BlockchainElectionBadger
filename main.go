package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"myapp/app/blockchain"
	"myapp/app/db"
	"myapp/app/model"
	"myapp/app/peer"
	"myapp/app/pkg/hmac"
	"myapp/app/pkg/signature"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dgraph-io/badger/v4"
)

const (
	dbPath = "./tmp/blocks"
)

type CommandLine struct {
	blockchainDB *blockchain.Blockchain
}

func main() {
	address := flag.String("address", "localhost:5000", "Address for node p2p network")
	init := flag.Bool("init", false, "init blockchain")
	flag.Parse()

	bootstrapAddress := "localhost:4000"
	// Membuat jaringan P2P dan kontrak voting.
	privateKey, err := signature.LoadOrCreateKeyPair("key.pem")
	if err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}
	publicKey, err := signature.SerializePublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Println("Error serializing public key:", err)
		return
	}

	dbConn, err := db.InitDB(dbPath)

	println(string(publicKey))
	p2p := peer.NewP2PNetwork(bootstrapAddress, *address, privateKey, publicKey, dbConn)
	p2p.RegisterToBootstrap()

	// Inisialisasi blockchain dengan instance Election
	p2p.Blockchain = &blockchain.Blockchain{
		Blocks:   []blockchain.Block{},
		Election: blockchain.NewElection([]string{}),
	}

	peers, err := p2p.GetPeersFromBootstrap()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, peer := range peers {
		fmt.Println("get peer from bootstrap", peer, peer.Address)
		if peer.Address != *address {
			p2p.AddPeer(peer)
			println("menambahkan: ", peer.Address)
		}
	}

	if *init {
		// p2p.Blockchain.Election.AddWarungCode("Alice")
		// p2p.Blockchain.Election.AddWarungCode("Bob")
		// p2p.Blockchain.Election.AddWarungCode("Charlie")
		println("prepare set genesis block")
		// Inisialisasi blockchain dan set genesis block
		powValid := p2p.Blockchain.SetGenesisBlock(dbConn)

		if !powValid {
			fmt.Println("Blockchain is not valid")
			p2p.BroadcastBlockchain()
		}

		defer dbConn.Close()
		//panggil printchain
		cli := CommandLine{blockchainDB: p2p.Blockchain}
		cli.printChain(dbConn)
	} else {
		// Sinkronisasi blockchain untuk peer baru
		fmt.Println("Requesting blockchain from peers...")

		p2p.Blockchain.GetLastHashCek(dbConn)

		p2p.RequestBlockchainFromPeers()
	}

	// Mendengarkan koneksi untuk menerima blok.
	go p2p.ListenForBlocks(*address)

	// go handleUserInput(p2p)

	// handling peer shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p2p.NotifyBootstrapOnShutdown()
		os.Exit(0)
	}()

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addRiwayatTransaksi(w, r, p2p)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		getRiwayatTransaksi(w, r, p2p)
	})

	// http.HandleFunc("/getbywarungcode", func(w http.ResponseWriter, r *http.Request) {
	// 	getRiwayatTransaksiByWarungCode(w, r, p2p)
	// })

	http.HandleFunc("/getblockelection", func(w http.ResponseWriter, r *http.Request) {
		getBlockdanElection(w, r, p2p)
	})

	go func() {
		http.ListenAndServe(":4040", nil)
	}()
	fmt.Println("Server started on port 4040")

	// go handleUserInput(p2p)

	// Menjaga agar program tetap berjalan.
	select {}
}

func handleUserInput(p2p *peer.P2PNetwork) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ketik perintah. Contoh: vote uniqueCodePayment warungCode atau showresult")
	for scanner.Scan() {
		input := scanner.Text()
		args := strings.Fields(input)

		if len(args) == 0 {
			fmt.Println("Masukkan perintah yang valid.")
			continue
		}

		switch args[0] {
		case "vote":
			if len(args) < 3 {
				fmt.Println("Perintah vote harus diikuti oleh voterID dan kandidatID.")
				continue
			}
			// voterID := args[1]
			// candidateID := args[2]
			// uniqueCodePayment := args[1]
			// amount := 1000
			// tax := "0"
			// warungCode := args[2]
			// payment := "payment"
			// date := "2023-10-01"
			// nik := "1234567890"
			// nikCustomer := "0987654321"

			// p2p.HandleVote(uniqueCodePayment, amount, tax, warungCode, payment, date, nik, nikCustomer)
			// fmt.Printf("History dari %s untuk %s telah dicatat.\n", uniqueCodePayment, warungCode)

		case "showresult":
			fmt.Println("Hasil history saat ini:")
			p2p.Blockchain.Election.DisplayResults()
		case "showblock":
			fmt.Println("Menampilkan blockchain:")
			p2p.Blockchain.Display()
		// case "caridatawarung":
		// 	if len(args) < 2 {
		// 		fmt.Println("Perintah caridatawarung harus diikuti oleh warungCode.")
		// 		continue
		// 	}

		// 	warungCode := args[1]
		// 	blocks := p2p.Blockchain.GetBlocksByWarungCode(warungCode)
		// 	if len(blocks) == 0 {
		// 		fmt.Println("Tidak ada data untuk warung code:", warungCode)

		// 		continue
		// 	}
		// 	fmt.Println("Data untuk warung code:", warungCode)
		// 	p2p.Blockchain.GetBlocksByWarungCode(warungCode)

		// 	for _, block := range blocks {
		// 		fmt.Printf("Index: %d\n", block.Index)
		// 		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		// 		fmt.Printf("Data: %v\n", block.Data)
		// 		fmt.Printf("Hash: %x\n", block.Hash)
		// 		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		// 		fmt.Printf("Nonce: %d\n", block.Nonce)
		// 		fmt.Println()
		// 	}
		case "exit":
		default:
			fmt.Println("Perintah tidak dikenal:", args[0])
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error membaca input:", err)
	}
}

func addRiwayatTransaksi(w http.ResponseWriter, r *http.Request, p2p *peer.P2PNetwork) {
	hmacSecret := r.Header.Get("X-HMAC")
	fmt.Println("HMAC Secret: ", hmacSecret)

	timeStampSecret := r.Header.Get("X-Timestamp")
	fmt.Println("Timestamp Secret: ", timeStampSecret)

	//ubah ke int64
	timeStampSecretInt, _ := strconv.ParseInt(timeStampSecret, 10, 64)

	// Verifikasi HMAC
	if hmacSecret == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "failed",
			"error":  "HMAC Secret is required",
		})
		return
	}

	// Verifikasi HMAC
	hmac.VerifyHMAC(os.Getenv("HMAC_KEY_BLOCKCHAIN_ELECTION"), hmacSecret, timeStampSecretInt, 60)

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	//beri log
	fmt.Println("Add request riwayat transaksi received")

	var requestData model.AddDataRiwayat

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Cek apakah uniqueCodePayment sudah ada di blockchain
	//nanti dicatat

	p2p.AddRiwayatTrx(requestData)
	// p2p.Blockchain.AddRiwayatTrx(requestData, p2p.dbConn)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"unique_code_payment": requestData.UniqueCodePayment,
			"amount":              requestData.Amount,
			"tax":                 requestData.Tax,
			"warung_code":         requestData.WarungCode,
			"payment":             requestData.Payment,
			"date":                requestData.Date,
			"nik":                 requestData.Nik,
			"nik_customer":        requestData.NikCustomer,
		},
		"message": "Riwayat transaksi berhasil ditambahkan",
	})
}

func getRiwayatTransaksi(w http.ResponseWriter, r *http.Request, p2p *peer.P2PNetwork) {
	// Verifikasi HMAC
	hmacSecret := r.Header.Get("X-HMAC")
	fmt.Println("HMAC Secret: ", hmacSecret)

	timeStampSecret := r.Header.Get("X-Timestamp")
	fmt.Println("Timestamp Secret: ", timeStampSecret)

	//ubah ke int64
	timeStampSecretInt, _ := strconv.ParseInt(timeStampSecret, 10, 64)

	// Verifikasi HMAC
	if hmacSecret == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "failed",
			"error":  "HMAC Secret is required",
		})
		return
	}

	hmac.VerifyHMAC(os.Getenv("HMAC_KEY_BLOCKCHAIN_ELECTION"), hmacSecret, timeStampSecretInt, 60)

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"blocks": p2p.Blockchain.GetBlocks(),
		},
		"message": "Riwayat transaksi berhasil diambil",
	})
}

// func getBlockdanElection
func getBlockdanElection(w http.ResponseWriter, r *http.Request, p2p *peer.P2PNetwork) {
	// Verifikasi HMAC
	hmacSecret := r.Header.Get("X-HMAC")
	fmt.Println("HMAC Secret: ", hmacSecret)

	timeStampSecret := r.Header.Get("X-Timestamp")
	fmt.Println("Timestamp Secret: ", timeStampSecret)

	//ubah ke int64
	timeStampSecretInt, _ := strconv.ParseInt(timeStampSecret, 10, 64)

	// Verifikasi HMAC
	if hmacSecret == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "failed",
			"error":  "HMAC Secret is required",
		})
		return
	}

	hmac.VerifyHMAC(os.Getenv("HMAC_KEY_BLOCKCHAIN_ELECTION"), hmacSecret, timeStampSecretInt, 60)

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"blocks":   p2p.Blockchain.GetBlocks(),
			"election": p2p.Blockchain.Election,
		},
	})
}

// func getRiwayatTransaksiByWarungCode(w http.ResponseWriter, r *http.Request, p2p *peer.P2PNetwork) {
// 	// Verifikasi HMAC
// 	hmacSecret := r.Header.Get("X-HMAC")
// 	fmt.Println("HMAC Secret: ", hmacSecret)

// 	timeStampSecret := r.Header.Get("X-Timestamp")
// 	fmt.Println("Timestamp Secret: ", timeStampSecret)

// 	//ubah ke int64
// 	timeStampSecretInt, _ := strconv.ParseInt(timeStampSecret, 10, 64)

// 	// Verifikasi HMAC
// 	if hmacSecret == "" {
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]interface{}{
// 			"status": "failed",
// 			"error":  "HMAC Secret is required",
// 		})
// 		return
// 	}

// 	hmac.VerifyHMAC(os.Getenv("HMAC_KEY_BLOCKCHAIN_ELECTION"), hmacSecret, timeStampSecretInt, 60)

// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var requestData struct {
// 		WarungCode string `json:"warung_code"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"status": "success",
// 		"data": map[string]interface{}{
// 			"blocks": p2p.Blockchain.GetBlocksByWarungCode(requestData.WarungCode),
// 		},
// 		"message": "Riwayat transaksi berhasil diambil",
// 	})
// }

func (cli *CommandLine) printChain(dbConn *badger.DB) {
	fmt.Println("Printing the chain...")

	//looping semuanya block dari dconn
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
				block := blockchain.Deserialize(v)
				fmt.Printf("Index: %d\n", block.Index)
				fmt.Printf("Timestamp: %d\n", block.Timestamp)
				fmt.Printf("Data: %v\n", block.Data)
				fmt.Printf("Hash: %x\n", block.Hash)
				fmt.Printf("PrevHash: %x\n", block.PrevHash)
				fmt.Printf("Nonce: %d\n", block.Nonce)
				fmt.Println()
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	blockchain.Handle(err)
	fmt.Println("Chain printed successfully.")
}
