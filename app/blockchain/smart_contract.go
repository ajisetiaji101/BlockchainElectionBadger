package blockchain

import (
	"fmt"
)

type Election struct {
	// Candidates []string
	// Votes      map[string]int
	// Voters     map[string]bool
	Amounts      map[string]int
	WarungCodes  []string
	Niks         []string
	TotalHistory map[string]int
	Records      map[string]bool
}

func NewElection(warungCodes []string) *Election {
	return &Election{
		WarungCodes:  warungCodes,
		Amounts:      make(map[string]int),
		Niks:         make([]string, 0),
		TotalHistory: make(map[string]int),
		Records:      make(map[string]bool),
	}
}

func (e *Election) AddWarungCode(warungCode string) error {
	if e.isValidCandidate(warungCode) {
		return fmt.Errorf("warung code %s already exists", warungCode)
	}
	e.WarungCodes = append(e.WarungCodes, warungCode)
	return nil
}

func (e *Election) Vote(uniqueCodePayment, warungCode string) error {
	// Validasi kandidat
	if !e.isValidCandidate(warungCode) {
		return fmt.Errorf("warung code %s is not valid", warungCode)
	}

	// Cek apakah pemilih sudah memberikan suara
	if e.Records[uniqueCodePayment] {
		return fmt.Errorf("unique code payment %s has already record", uniqueCodePayment)
	}

	// Tambahkan suara
	e.TotalHistory[warungCode]++
	e.Records[warungCode] = true
	return nil
}

func (e *Election) isValidCandidate(warungCode string) bool {
	for _, c := range e.WarungCodes {
		if c == warungCode {
			return true
		}
	}
	return false
}

func (e *Election) GetResults() map[string]int {
	return e.TotalHistory
}

func (e *Election) DisplayResults() {
	results := e.GetResults()
	for warungCode, record := range results {
		fmt.Print("isi result", results)
		fmt.Printf("Warung Code : %s, Record: %d\n", warungCode, record)
	}
}
