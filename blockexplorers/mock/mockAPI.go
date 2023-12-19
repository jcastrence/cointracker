package mock

import (
	"cointracker/models"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const mockDataDirpath = "blockexplorers/mock/mockData"

type addressDataJSON struct {
	FinalBalance uint64   `json:"final_balance"`
	Txs          []txJSON `json:"txs"`
}

type txJSON struct {
	Hash       string `json:"hash"`
	BlockIndex uint32 `json:"block_index"`
	Result     int64  `json:"result"`
}

type MockHandler struct{}

func NewMockHandler() MockHandler {
	return MockHandler{}
}

func (b MockHandler) GetAddressData(address string, offset int) (*models.AddressData, error) {
	path := filepath.Join(mockDataDirpath, fmt.Sprintf("%soffset%d.json", address, offset))
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var addrData addressDataJSON
	err = json.Unmarshal(jsonBytes, &addrData)
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction
	for _, tx := range addrData.Txs {
		transactions = append(transactions, models.Transaction{
			Hash:   tx.Hash,
			Block:  tx.BlockIndex,
			Result: tx.Result,
		})
	}

	return &models.AddressData{
		Address:      address,
		Balance:      addrData.FinalBalance,
		Transactions: transactions,
	}, nil
}
