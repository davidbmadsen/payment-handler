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

	var customers Customers = Customers{accounts: make(map[int]Account)}

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

	// assing variables for clarity
	paymentType := row[0]
	customerId, _ := strconv.Atoi(row[1]) // Todo: add error handling
	amount, _ := strconv.ParseFloat(row[2], 64)

	switch paymentType {
	case "deposit":
		fmt.Println("Deposit")
		updateAccountBalance(customers, customerId, amount, false)
	case "withdraw":
		fmt.Println("Withdraw")
		updateAccountBalance(customers, customerId, -amount, true)
	case "dispute":
		fmt.Println("Dispute")
	case "chargeback":
		fmt.Println("Chargeback")
	case "resolve":
		fmt.Println("Resolve")
	default:
		fmt.Println("Unknown transaction type")
	}
}

func updateAccountBalance(customers *Customers, customerId int, amount float64, isWithdrawal bool) {
	// get account by id
	account, ok := customers.accounts[customerId]

	if !ok {
		if isWithdrawal {
			panic("Can't withdraw from non-existing account")
		}
		// if account doesnt exist, create account
		account = Account{id: customerId, available: amount, hold: 0, total: amount, frozen: false}
	} else {
		// update account total and available
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
	}
	customers.accounts[customerId] = account
}

func printCustomerAccounts(customers *Customers) {
	fmt.Println("customer, available, hold, total, frozen")
	for _, account := range customers.accounts {
		fmt.Printf("%v, %v, %v, %v, %v\n", account.id, account.available, account.hold, account.total, account.frozen)
	}
}

type Customers struct {
	accounts map[int]Account
}

type Account struct {
	id        int
	available float64
	hold      float64
	total     float64
	frozen    bool
}
