package main

import (
	"cointracker/blockexplorers/mock"
	"cointracker/database"
	"cointracker/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"cointracker/crypto"

	"github.com/gorilla/mux"
)

const port = 8000

var dbh *database.DatabaseHandler
var beh BlockExplorerHandler

type BlockExplorerHandler interface {
	GetAddressData(address string, offset int) (*models.AddressData, error)
}

func main() {
	// Initialize connection with datagbase
	dbh = database.NewDatabaseHandler()
	if err := dbh.Initialize(); err != nil {
		log.Fatal(err)
	}

	// Initialize block	explorer
	beh = mock.NewMockHandler()

	// Endpoints
	r := mux.NewRouter()
	r.HandleFunc("/createaccount", createAccount).Methods("POST")
	r.HandleFunc("/addaddresses", addAddresses).Methods("POST")
	r.HandleFunc("/removeaddresses", removeAddresses).Methods("DELETE")
	r.HandleFunc("/getaccountinfo", getAccountInfo).Methods("GET")
	r.HandleFunc("/updateaccount", updateAccount).Methods("PUT")

	http.Handle("/", r)
	fmt.Println("Cointracker running on port: ", port)
	serverAddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}

// Handle Funcs

// Create Account
type createAccountParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var p createAccountParams
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if account already exists
	_, err = dbh.RetrieveAccount(p.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new account
			passwordHash, err := crypto.HashPassword(p.Password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = dbh.InsertAccount(p.Username, passwordHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Printf("New account created: %s\n", p.Username)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// No error, account already exists, return error
		http.Error(w, "Account already exists", http.StatusBadRequest)
		return
	}
}

// Add Addresses
type addAddressesParams struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Addresses []string `json:"addresses"`
}

func addAddresses(w http.ResponseWriter, r *http.Request) {
	var p addAddressesParams
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateAccount(p.Username, p.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addedCount := 0
	for _, addr := range p.Addresses {
		added, err := addAddress(addr, p.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if added {
			addedCount++
		}
	}

	fmt.Printf("%d address(es) added for account: %s\n", addedCount, p.Username)
}

// Returns true if address was added, false otherwise
func addAddress(address, username string) (bool, error) {
	// Check if address already exists
	addr, err := dbh.RetrieveAddress(address)
	if addr != nil {
		return false, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	// Get blockchain data for address
	addressData, err := beh.GetAddressData(address, 0)
	if err != nil {
		return false, err
	}

	// Insert address into db
	err = dbh.InsertAddress(address, addressData.Balance, username)
	if err != nil {
		return false, err
	}

	// Insert transactions into db
	for _, tx := range addressData.Transactions {
		err = dbh.InsertTransaction(tx.Hash, tx.Block, tx.Result, address)
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Remove Addresses
type removeAddressesParams struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Addresses []string `json:"addresses"`
}

func removeAddresses(w http.ResponseWriter, r *http.Request) {
	var p removeAddressesParams
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateAccount(p.Username, p.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	removedCount := 0
	for _, addr := range p.Addresses {
		removed, err := removeAddress(addr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if removed {
			removedCount++
		}
	}

	fmt.Printf("%d address(es) removed for account: %s\n", removedCount, p.Username)
}

// Returns true if address was removed, false otherwise
func removeAddress(address string) (bool, error) {
	// Check if address exists
	_, err := dbh.RetrieveAddress(address)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// Delete address from db
	err = dbh.DeleteAddress(address)
	if err != nil {
		return false, err
	}

	// Delete transactions from db
	err = dbh.DeleteTransactions(address)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Get Account Info
type getAccountInfoParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type getAccountInfoResponse struct {
	AccountBalance        uint64               `json:"accountBalance"`
	TotalTransactionCount int                  `json:"totalTransactionCount"`
	Addresses             []models.AddressData `json:"addresses"`
}

func getAccountInfo(w http.ResponseWriter, r *http.Request) {
	var p getAccountInfoParams
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateAccount(p.Username, p.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get account addresses
	addresses, err := dbh.RetrieveAddresses(p.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := getAccountInfoResponse{}
	for _, addr := range addresses {
		// Total account balance
		response.AccountBalance += addr.Balance
		// Prepare address info
		addressData := models.AddressData{
			Address: addr.AddrKey,
			Balance: addr.Balance,
		}
		// Prepare transaction info for address
		transactions, err := dbh.RetrieveTransactions(addr.AddrKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, tx := range transactions {
			addressData.Transactions = append(addressData.Transactions, models.Transaction{
				Hash:   tx.TxHash,
				Block:  tx.Block,
				Result: tx.Result,
			})
			response.TotalTransactionCount++
		}

		response.Addresses = append(response.Addresses, addressData)
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	fmt.Printf("Info requested for: %s\n", p.Username)
}

// Update Account
type updateAccountParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func updateAccount(w http.ResponseWriter, r *http.Request) {
	var p updateAccountParams
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateAccount(p.Username, p.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get account addresses
	addresses, err := dbh.RetrieveAddresses(p.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updateCount int
	for _, addr := range addresses {
		updated, err := updateAddress(addr.AddrKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updated {
			updateCount++
		}
	}

	fmt.Printf("%d address(es) updated for account: %s\n", updateCount, p.Username)
}

// Returns true if address was updated, false if otherwise
func updateAddress(address string) (bool, error) {
	// Obtain count to determine offset
	count, err := dbh.CountTransactions(address)
	if err != nil {
		return false, err
	}

	// Request new transaction data from block explorer
	addressData, err := beh.GetAddressData(address, count)
	if err != nil {
		return false, err
	}

	if len(addressData.Transactions) > 0 {
		// New transactions found, insert transactions into db
		for _, tx := range addressData.Transactions {
			err = dbh.InsertTransaction(tx.Hash, tx.Block, tx.Result, address)
		}
		if err != nil {
			return false, err
		}
		// Update address balance
		err = dbh.UpdateAddressBalance(address, addressData.Balance)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Helper funcs

// Validation
func validateAccount(username, password string) error {
	acc, err := dbh.RetrieveAccount(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Account not found")
		}
		return err
	}
	valid := crypto.CheckPassword(password, acc.PasswordHash)
	if valid {
		return nil
	} else {
		return errors.New("Invalid password")
	}
}
