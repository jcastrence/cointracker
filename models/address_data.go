package models

// Represents Bitcoin address data
type AddressData struct {
	Address      string        `json:"address"`      // Address key
	Balance      uint64        `json:"balance"`      // Current balance in Satoshis
	Transactions []Transaction `json:"transactions"` // List of all known transactions associated with this address
}

// Represents a Bitcoin transaction
type Transaction struct {
	Hash   string `json:"hash"`   // Transaction hash
	Block  uint32 `json:"block"`  // Transaction block
	Result int64  `json:"result"` // Result of transaction to address balance
}
