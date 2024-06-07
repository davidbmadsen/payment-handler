package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	command := os.Args[1]
	fileName := os.Args[2]
	
	if command != "-f" && command != "--file" {
		log.Fatal("Invalid command. Please use -f or --file")
	}

	file, err := openCsvFile(fileName)

	if err != nil {
		log.Fatal("Error while reading the file: ", err)
	}
	
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		fmt.Println("Error reading records")
	}

	var customers Customers = Customers{accounts: make(map[int]*Account)}

	for index, record := range records {
		// skip header
		if index == 0 {
			continue
		}

		customerId, transaction, err := parseTransaction(record)
		if err != nil {
			fmt.Println("Error parsing transaction:", err)
			continue
		}

		handleTransaction(&customers, customerId, transaction)
	}

	printCustomerAccounts(&customers)
}

func parseTransaction(record []string) (int, *Transaction, error) {
	
	transactionType := record[0]
	customerId, c_err := strconv.Atoi(record[1])
	transactionId, t_err := strconv.Atoi(record[2])
	
	var amount float64 
	var amt_err error

	if transactionType == "deposit" || transactionType == "withdraw" {
		
		if(record[3] == "") {
			return -1, &Transaction{}, fmt.Errorf("amount cant be empty for transaction type %v", transactionType)
		}

		amount, amt_err = strconv.ParseFloat(record[3], 64)
	}
	
	if c_err != nil || t_err != nil || amt_err != nil {
		return -1, &Transaction{}, fmt.Errorf("one or more fields are not valid")
	}

	return customerId, &Transaction{transactionId, transactionType, amount}, nil
}

func openCsvFile(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)

	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	if !strings.HasSuffix(fileName, ".csv") { 
		err = fmt.Errorf("format should be csv")
	}

	return file, err
}

func handleTransaction(customers *Customers, customerId int, transaction *Transaction) {

	account, accountError := getOrCreateAccount(customers, customerId, transaction.txType)

	if accountError != nil {
		fmt.Printf("Error getting account for customer %v: %s", customerId, accountError)
		return
	}

	if account.frozen && transaction.txType != "resolve" {
		fmt.Printf("Account %v is frozen", customerId)
		return
	}

	var err error

	switch transaction.txType {
	case "deposit":
		err = updateAccountBalance(account, customerId, transaction.id, transaction.amount, false)
	case "withdraw":
		err = updateAccountBalance(account, customerId, transaction.id, transaction.amount, true)
	case "dispute":
		err = handleDispute(account, transaction.id)
	case "chargeback":
		err = handleChargeback(account, transaction.id)
	case "resolve":
		err = handleResolve(account, transaction.id)
	default:
		fmt.Printf("Unknown transaction type %v", transaction.txType)
	}

	if err != nil {
		fmt.Printf("Problem processing transaction %v: %s\n", transaction.id, err)
	}
}

func getOrCreateAccount(customers *Customers, customerId int, txType string) (*Account, error) {

	account, accountExists := customers.accounts[customerId]

	if !accountExists {
		// if the account doesn't exist and transaction type isn't "deposit", we cant create a new account
		if txType != "deposit" {
			return account, fmt.Errorf("can't perform transaction on non-existing account")
		}
		
		customers.accounts[customerId] = &Account{id: customerId, available: 0, hold: 0, total: 0, frozen: false, transactions: make(map[int]*Transaction)}
	}

	return customers.accounts[customerId], nil
}

func updateAccountBalance(account *Account, customerId int, transactionId int, amount float64, isWithdrawal bool) error {

	if isWithdrawal {
		if account.available < amount {
			return fmt.Errorf("not enough funds - can't withdraw more than available balance")
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

	// add transaction to account
	account.transactions[transactionId] = &Transaction{id: customerId, txType: txType, amount: amount}
	return nil
}

func handleDispute(account *Account, transactionId int) error {
	
	if account.transactions[transactionId] == nil {
		return fmt.Errorf("transaction to dispute not found")
	}
		
	disputeAmount := account.transactions[transactionId].amount
	if account.available < disputeAmount {
		return fmt.Errorf("not enough funds available to dispute")
	}

	// move funds from available to hold, account total remains unchanged
	account.available -= disputeAmount
	account.hold += disputeAmount

	return nil
}

func handleChargeback(account *Account, transactionId int) error {

	if account.transactions[transactionId] == nil {
		return fmt.Errorf("transaction to charge back not found")
	}

	// freeze account and charge back transaction amount
	account.frozen = true
	account.hold -= account.transactions[transactionId].amount
	account.total -= account.transactions[transactionId].amount
	return nil
}

func handleResolve(account *Account, transactionId int) error {

	if account.transactions[transactionId] == nil {
		return fmt.Errorf("Transaction to resolve not found")
	}

	if account.hold < account.transactions[transactionId].amount {
		return fmt.Errorf("not enough funds on hold to resolve transaction %v", transactionId)
	}

	// move funds back to available and unlock account
	account.frozen = false
	account.available += account.transactions[transactionId].amount
	account.hold -= account.transactions[transactionId].amount
	return nil
}

func printCustomerAccounts(customers *Customers) {
	
	fmt.Println("\ncustomer, available, hold, total, frozen")
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
