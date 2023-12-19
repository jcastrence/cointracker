package blockchain

import (
	"cointracker/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const rawaddrEndpoint string = "https://blockchain.info/rawaddr/"
const limitParam = 100

type rawaddrResponseJSON struct {
	FinalBalance uint64   `json:"final_balance"`
	Txs          []txJSON `json:"txs"`
}

type txJSON struct {
	Hash       string `json:"hash"`
	BlockIndex uint32 `json:"block_index"`
	Result     int64  `json:"result"`
}

type BlockchainHandler struct{}

func NewBlockchainHandler() BlockchainHandler {
	return BlockchainHandler{}
}

func (b BlockchainHandler) GetAddressData(address string, offset int) (*models.AddressData, error) {
	// Build transactions list
	var transactions []models.Transaction

	for {
		// Request to Blockchain.com for adddress data
		client := &http.Client{Timeout: 10 * time.Second}
		url := fmt.Sprintf("%s%s?limit=%d&offset=%d", rawaddrEndpoint, address, limitParam, offset)
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Decode JSON response
		var respJSON rawaddrResponseJSON
		err = json.NewDecoder(resp.Body).Decode(&respJSON)
		if err != nil {
			return nil, err
		}

		if len(respJSON.Txs) == 0 {
			// No more transactions
			return &models.AddressData{
				Address:      address,
				Balance:      respJSON.FinalBalance,
				Transactions: transactions,
			}, nil
		} else {
			// Append to transaction list
			for _, tx := range respJSON.Txs {
				transactions = append(transactions, models.Transaction{
					Hash:   tx.Hash,
					Block:  tx.BlockIndex,
					Result: tx.Result,
				})
			}
			offset += limitParam
		}
	}
}
