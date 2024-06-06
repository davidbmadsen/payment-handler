package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {

	fileName := os.Args[1]

	file, err := os.Open(fileName)

	if err != nil {
		log.Fatal("Error while reading the file", err)
	}
	fmt.Println("Loaded CSV file:", fileName)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		fmt.Println("Error reading records")
	}

	var customers Customers = Customers{accounts: make(map[int]*Account)}

	for index, record := range records {
		if index == 0 {
			// skip header
			continue
		}
		handleRow(&customers, record)
	}

	printCustomerAccounts(&customers)
}

func handleRow(customers *Customers, row []string) {

	paymentType := row[0]
	customerId, _ := strconv.Atoi(row[1]) // Todo: add error handling
	transactionId, _ := strconv.Atoi(row[2])
	amount, _ := strconv.ParseFloat(row[3], 64)
	
	account := getOrCreateAccount(customers, customerId, paymentType)

	switch paymentType {
	case "deposit":
		updateAccountBalance(account, customerId, transactionId, amount, false)
	case "withdraw":
		updateAccountBalance(account, customerId, transactionId, -amount, true)
	case "dispute":
		handleDispute(account, customerId)
	case "chargeback":
		handleChargeback(account, customerId)
	case "resolve":
		handleResolve(account, customerId)
	default:
		fmt.Printf("Unknown transaction type %v", paymentType)
	}
}

func getOrCreateAccount(customers *Customers, customerId int, txType string) *Account {

	_, accountExists := customers.accounts[customerId]

	if !accountExists {
		// if the account doesn't exist and transaction type isn't "deposit", we cant create a new account
		if txType != "deposit" {
			panic("Can't perform transaction on non-existing account")
		}
		fmt.Printf("Creating account for customer %v\n", customerId)
		customers.accounts[customerId] = &Account{id: customerId, available: 0, hold: 0, total: 0, frozen: false, transactions: make(map[int]*Transaction)}
	}

	return customers.accounts[customerId]
}

func updateAccountBalance(account *Account, customerId int, transactionId int, amount float64, isWithdrawal bool) {

	if isWithdrawal {
		if account.available < amount {
			panic("Not enough funds")
		}
		account.total -= amount
		account.available -= amount
	} else {
		account.total += amount
		account.available += amount
	}

	var txType string
	if isWithdrawal {
		txType = "withdraw"
	} else {
		txType = "deposit"
	}

	account.transactions[transactionId] = &Transaction{id: customerId, txType: txType, amount: amount}
}

func handleDispute(account *Account, transactionId int) {
	
	if account.transactions[transactionId] == nil {
		panic("Transaction to dispute not found")
	}
		
	disputeAmount := account.transactions[transactionId].amount
	if account.available < disputeAmount {
		// assuming we want to dispute max available funds up to dispute amount
		account.hold = account.available
		account.available = 0
	}

	// move funds from available to hold, account total remains unchanged
	account.available -= disputeAmount
	account.hold += disputeAmount
}

func handleChargeback(account *Account, transactionId int){

	if account.transactions[transactionId] == nil {
		panic("Transaction to charge back not found")
	}

	// freeze account and charge back transaction amount
	account.frozen = true
	account.hold -= account.transactions[transactionId].amount
	account.total -= account.transactions[transactionId].amount
}

func handleResolve(account *Account, transactionId int) {

	if account.transactions[transactionId] == nil {
		panic("Transaction to resolve not found")
	}

	// move funds back to available and unlock account
	account.frozen = false
	account.available += account.transactions[transactionId].amount
	account.hold = 0
}

func printCustomerAccounts(customers *Customers) {
	fmt.Println("customer, available, hold, total, frozen")
	for _, account := range customers.accounts {
		fmt.Printf("%v, %v, %v, %v, %v\n", account.id, account.available, account.hold, account.total, account.frozen)
	}
}

type Customers struct {
	accounts map[int]*Account
}

type Transaction struct {
	id     int
	txType string
	amount float64
}

type Account struct {
	id           int
	available    float64
	hold         float64
	total        float64
	frozen       bool
	transactions map[int]*Transaction
}
