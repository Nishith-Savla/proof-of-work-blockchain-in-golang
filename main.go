package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/spewerspew/spew"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const difficulty = 1

type Block struct {
	Index      int    `json:"index"`
	Timestamp  string `json:"timestamp"`
	Data       int    `json:"data"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prev_hash"`
	Difficulty int    `json:"difficulty"`
	Nonce      string `json:"nonce"`
}

type Message struct {
	Data int `json:"data"`
}

var Blockchain []Block

var mutex sync.Mutex

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{
			Index:      0,
			Timestamp:  t.String(),
			Data:       0,
			PrevHash:   "",
			Difficulty: difficulty,
			Nonce:      "",
		}
		genesisBlock.Hash = calculateHash(genesisBlock)
		spew.Dump(genesisBlock)
		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()

	log.Fatalln(run())
}

func run() error {
	router := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	log.Println("HTTP Server is listening on port ", httpPort)
	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func makeMuxRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", handleGetBlockchain).Methods(http.MethodGet)
	router.HandleFunc("/", handleWriteBlock).Methods(http.MethodPost)
	return router
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m Message

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	newBlock := generateBlock(Blockchain[len(Blockchain)-1], m.Data)

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	log.Println(w.Write(response))
}

func isBlockValid(newBlock, prevBlock Block) bool {
	if prevBlock.Index+1 != newBlock.Index {
		return false
	}

	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Data) + block.PrevHash + block.Nonce
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(prevBlock Block, data int) Block {
	newBlock := Block{
		Index:      prevBlock.Index + 1,
		Timestamp:  time.Now().String(),
		Data:       data,
		Hash:       "",
		PrevHash:   prevBlock.Hash,
		Difficulty: difficulty,
	}

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		hash := calculateHash(newBlock)
		if !isHashValid(hash, newBlock.Difficulty) {
			fmt.Println(hash, "Do more work!")
			time.Sleep(time.Second)
			continue
		}

		fmt.Println(hash, "Work done!")
		newBlock.Hash = hash
		break
	}

	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}
