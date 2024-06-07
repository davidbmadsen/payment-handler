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

func handleRow(customers *Customers, row []string) {

	paymentType := row[0]
	customerId, _ := strconv.Atoi(row[1]) // Todo: add error handling
	transactionId, _ := strconv.Atoi(row[2])
	amount, _ := strconv.ParseFloat(row[3], 64)
	
	account := getOrCreateAccount(customers, customerId, paymentType)
	var err error
	switch paymentType {
	case "deposit":
		err = updateAccountBalance(account, customerId, transactionId, amount, false)
	case "withdraw":
		err = updateAccountBalance(account, customerId, transactionId, -amount, true)
	case "dispute":
		err = handleDispute(account, customerId)
	case "chargeback":
		err = handleChargeback(account, customerId)
	case "resolve":
		err = handleResolve(account, customerId)
	default:
		fmt.Printf("Unknown transaction type %v", paymentType)
	}

	if err != nil {
		fmt.Printf("Problem processing transaction %v: %s", transactionId, err)
	}
}

func getOrCreateAccount(customers *Customers, customerId int, txType string) *Account {

	_, accountExists := customers.accounts[customerId]

	if !accountExists {
		// if the account doesn't exist and transaction type isn't "deposit", we cant create a new account
		if txType != "deposit" {
			panic("Can't perform transaction on non-existing account")
		}
		
		customers.accounts[customerId] = &Account{id: customerId, available: 0, hold: 0, total: 0, frozen: false, transactions: make(map[int]*Transaction)}
	}

	return customers.accounts[customerId]
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

	account.transactions[transactionId] = &Transaction{id: customerId, txType: txType, amount: amount}
	return nil
}

func handleDispute(account *Account, transactionId int) error {
	
	if account.transactions[transactionId] == nil {
		return fmt.Errorf("transaction to dispute not found")
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

	// move funds back to available and unlock account
	account.frozen = false
	account.available += account.transactions[transactionId].amount
	account.hold = 0
	return nil
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
