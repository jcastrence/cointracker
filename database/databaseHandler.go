package database

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/mattn/go-sqlite3"
)

const (
	driverName = "sqlite3"
	dbFilename = "cointracker.db"
)

type DatabaseHandler struct {
	db     *sql.DB
	driver sqlite3.SQLiteDriver
}

type Account struct {
	Username     string
	PasswordHash string
}

type Address struct {
	AddrKey  string
	Balance  uint64
	Username string
}

type Transaction struct {
	TxHash  string
	Block   uint32
	Result  int64
	AddrKey string
}

func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{}
}

func (h *DatabaseHandler) Open() (err error) {
	// Open connection to database
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dbDir := filepath.Join(wd, "db/")
	_ = os.Mkdir(dbDir, os.ModePerm)
	h.db, err = sql.Open(driverName, filepath.Join(dbDir, dbFilename))
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) Close() (err error) {
	err = h.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) CreateTables() (err error) {
	// Create non existing tables
	_, err = h.db.Exec("CREATE TABLE IF NOT EXISTS accounts(username TEXT PRIMARY KEY, passwordHash TEXT NOT NULL);" +
		"CREATE TABLE IF NOT EXISTS addresses(addrKey TEXT PRIMARY KEY, balance INT NOT NULL, username TEXT NOT NULL);" +
		"CREATE TABLE IF NOT EXISTS transactions(txHash TEXT PRIMARY KEY, block INT NOT NULL, result INT NOT NULL, addrKey INT NOT NULL);")
	if err != nil {
		return err
	}

	// Index tables for faster lookups
	_, err = h.db.Exec("CREATE INDEX IF NOT EXISTS address_username_idx ON addresses USING HASH(username);" +
		"CREATE INDEX IF NOT EXISTS transaction_addrKey_idx ON transactions USING HASH(addrKey);")

	return nil
}

func (h *DatabaseHandler) Initialize() (err error) {
	err = h.Open()
	if err != nil {
		return err
	}
	err = h.CreateTables()
	if err != nil {
		return err
	}
	return nil
}

// Account queries
func (h *DatabaseHandler) InsertAccount(username, passwordHash string) (err error) {
	_, err = h.db.Exec("INSERT INTO accounts VALUES(?,?);", username, passwordHash)
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) RetrieveAccount(username string) (account *Account, err error) {
	row := h.db.QueryRow("SELECT username, passwordHash FROM accounts WHERE username=?;", username)
	if err != nil {
		return nil, err
	}
	account = &Account{}
	err = row.Scan(&account.Username, &account.PasswordHash)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// Address queries
func (h *DatabaseHandler) InsertAddress(addrKey string, balance uint64, username string) (err error) {
	_, err = h.db.Exec("INSERT INTO addresses VALUES(?,?,?);", addrKey, balance, username)
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) DeleteAddress(addrKey string) (err error) {
	_, err = h.db.Exec("DELETE FROM addresses WHERE addrKey=?;", addrKey)
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) RetrieveAddress(addrKey string) (address *Address, err error) {
	row := h.db.QueryRow("SELECT addrKey, balance, username FROM addresses WHERE addrKey=?;", addrKey)
	if err != nil {
		return nil, err
	}

	address = &Address{}
	err = row.Scan(&address.AddrKey, &address.Balance, &address.Username)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func (h *DatabaseHandler) RetrieveAddresses(username string) (addresses []*Address, err error) {
	rows, err := h.db.Query("SELECT addrKey, balance, username FROM addresses WHERE username=?;", username)
	if err != nil {
		return nil, err
	}

	addresses = []*Address{}
	for rows.Next() {
		address := &Address{}
		err = rows.Scan(&address.AddrKey, &address.Balance, &address.Username)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (h *DatabaseHandler) UpdateAddressBalance(addrKey string, newBalance uint64) (err error) {
	_, err = h.db.Exec("UPDATE addresses SET balance=? WHERE addrKey=?;", newBalance, addrKey)
	if err != nil {
		return err
	}
	return nil
}

// Transaction queries
func (h *DatabaseHandler) InsertTransaction(txHash string, block uint32, result int64, address string) (err error) {
	_, err = h.db.Exec("INSERT INTO transactions VALUES(?,?,?,?);", txHash, block, result, address)
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) DeleteTransactions(addrKey string) (err error) {
	_, err = h.db.Exec("DELETE FROM transactions WHERE addrKey=?;", addrKey)
	if err != nil {
		return err
	}
	return nil
}

func (h *DatabaseHandler) RetrieveTransactions(addrKey string) (transactions []*Transaction, err error) {
	rows, err := h.db.Query("SELECT txHash, block, result, addrKey FROM transactions WHERE addrKey=?;", addrKey)
	if err != nil {
		return nil, err
	}

	transactions = []*Transaction{}
	for rows.Next() {
		transaction := &Transaction{}
		err = rows.Scan(&transaction.TxHash, &transaction.Block, &transaction.Result, &transaction.AddrKey)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *DatabaseHandler) CountTransactions(addrKey string) (count int, err error) {
	rows, err := h.db.Query("SELECT COUNT(*) FROM transactions WHERE addrKey=?;", addrKey)
	if err != nil {
		return 0, err
	}

	if rows.Next() {
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		return count, nil
	} else {
		return 0, errors.New("Unexpected path")
	}
}
