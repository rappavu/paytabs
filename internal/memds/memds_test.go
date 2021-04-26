package memds

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"paytabs/internal/ds"
)

const datafile string = "../../data/accounts-mock.json"

var gAccounts []ds.Account

// Test Setup
func TestMain(m *testing.M) {
	// disable logging when tests are run
	log.SetOutput(ioutil.Discard)

	// read the data file and initialize gAccounts
	f, _ := os.Open(datafile)
	bytes, _ := io.ReadAll(f)
	if err := json.Unmarshal(bytes, &gAccounts); err != nil {
		fmt.Printf("Error reading json data file: %v\n", datafile)
		fmt.Println("In-memory DataStore test setup failed. In-memory DataStore tests skipped.")
		return
	}
}

func TestLoad(t *testing.T) {
	// test loading the sample json accounts file paytabs/data/accounts-mock.json
	d, err := Load(datafile)
	if err != nil {
		t.Fatal("Failed to load json file")
	}
	if d == nil {
		t.Fatal("Failed to get datastore pointer")
	}

	// validate results
	if !reflect.DeepEqual(gAccounts, d.accounts) {
		t.Fatal("Loaded []Accounts data does not match with the expected")
	}
}

func TestList(t *testing.T) {
	d, _ := Load(datafile)
	accts := d.List()

	// validate results
	if !reflect.DeepEqual(gAccounts, accts) {
		t.Fatal("Received []Accounts data does not match with the expected")
	}
}

func TestListParallel(t *testing.T) {
	d, _ := Load(datafile)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			accts := d.List()
			// validate results
			if !reflect.DeepEqual(gAccounts, accts) {
				t.Fatalf("%v: Received []Accounts data does not match with the expected", name)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	d, _ := Load(datafile)
	amount := 1.5 // amount to transfer
	// we know first gAccount[0] has more than 1.5 in its balance
	_, bal, err := d.Transfer(gAccounts[0].Id, gAccounts[1].Id, amount)
	if err != nil {
		t.Fatalf("Failed to transfer funds")
	}
	expectedBalance := (gAccounts[0].Balance - amount)
	if expectedBalance != bal {
		t.Fatalf("Received an incorrect balance value, received: %v, expected: %v\n", bal, expectedBalance)
	}
}

func TestTransferParallel(t *testing.T) {
	d, _ := Load(datafile)
	amount := 1.0
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// make sure there is no deadlock
			d.Transfer(gAccounts[0].Id, gAccounts[1].Id, amount)
			d.Transfer(gAccounts[1].Id, gAccounts[0].Id, amount)
		})
	}
}

func TestListAndTransferParallel(t *testing.T) {
	d, _ := Load(datafile)
	amount := 1.0
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// make sure there is no deadlock
			d.List()
			d.Transfer(gAccounts[0].Id, gAccounts[1].Id, amount)
			d.Transfer(gAccounts[1].Id, gAccounts[0].Id, amount)
			d.List()
		})
	}
}

func TestGet(t *testing.T) {
	d, _ := Load(datafile)
	acct, err := d.Get(gAccounts[0].Id)
	if err != nil {
		t.Fatal("Error getting account details")
	}

	// validate results
	if acct != gAccounts[0] {
		t.Fatal("Received account details does not match with the expected")
	}
}

func TestGetParallel(t *testing.T) {
	d, _ := Load(datafile)
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// make sure there is no deadlock
			acct, _ := d.Get(gAccounts[i].Id)
			// validate results
			if acct != gAccounts[0] {
				t.Fatal("Received account details does not match with the expected")
			}
		})
	}
}

func TestValidateTransferParallel(t *testing.T) {
	d, _ := Load(datafile)
    t.Run("group", func(t *testing.T) {
        t.Run("List", TestListParallel)
        t.Run("Transfer", TestTransferParallel)
        t.Run("Get", TestGetParallel)
    })

	// after all the tests are finished 
	// as all debit and credit evens out, datastore should remain unchanged
	accts := d.List()

	// validate results
	if !reflect.DeepEqual(gAccounts, accts) {
		t.Fatal("Received []Accounts data does not match with the expected")
	}

}

// end-of-file
