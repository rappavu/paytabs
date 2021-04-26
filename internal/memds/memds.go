// Implements an in-memory datastore. Supports ds.Datastore interface
//
package memds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"paytabs/internal/ds"
)

// structure representing a transaction
type transaction struct {
	tid    uint64    // transaction id
	date   time.Time // date and time of the transaction
	from   string    // transfered from
	to     string    // transfered to
	amount float64   // amouont transfered
}

// structure for in-mempory datastore containing all the account details and
// transactions performed
type datastore struct {
	accounts     []ds.Account   // list of accounts
	locks        []sync.Mutex   // row locks
	index        map[string]int // index for id
	transactions []transaction  // list of transactions handled
	tlock        sync.Mutex     // transaction lock
	nextTid      uint64         // next transaction id
}

// Load Account data from a file.
//
// Account data is expected in jason format in the specified file.
func Load(filename string) (*datastore, error) {
	log.Printf("[memds]loading data from file: %v\n", filename)

	// open the json file
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("[memds]failed to open file: %s - %s\n", filename, err)
		return nil, err
	}

	// read the contents of the file to an array, jbytes
	jbytes := make([]byte, 0, 32*1024) // initial size 32K for json bytes container
	bytes := make([]byte, 32*1024)     // read in 32K blocks
	for {
		n, err := f.Read(bytes)
		if err == io.EOF {
			break // no more data to read
		}
		if err != nil {
			log.Printf("[memds]error reading file: %s - %s\n", filename, err)
			return nil, err
		}
		if n > 0 {
			jbytes = append(jbytes, bytes[:n]...)
		}
	}
	log.Println("[memds]file read complete")

	// parse the json bytes and populate the Accounts array
	var accounts []ds.Account
	err = json.Unmarshal(jbytes, &accounts)
	if err != nil {
		log.Printf("[memds]failed to parse json data from file: %s - %s\n", filename, err)
		log.Printf("%s\n", string(jbytes))
		return nil, err
	}
	log.Println("[memds]json data unmarshall complete")

	// populate index and record locks
	n := len(accounts)
	index := make(map[string]int, n)
	locks := make([]sync.Mutex, n)
	for i := 0; i < n; i++ {
		index[accounts[i].Id] = i
		locks[i] = sync.Mutex{}
	}
	log.Printf("[memds]indexing complete")

	// construct the in-memory datastore and return
	d := new(datastore)
	d.accounts = accounts
	d.index = index
	d.locks = locks
	d.nextTid = 1 // initial transaction id
	log.Println("[memds]datastore initialization complete")

	return d, nil
}

// Locks the whole table by acquiring all the row locks.
func (d *datastore) lockTable() {
	log.Println("[memds]attempting to lock table")

	// acquire all the row locks
	// we need to lock in ascending order to rows to prevent deadlock
	for i := 0; i < len(d.locks); i++ {
		d.locks[i].Lock()
	}
	log.Println("[memds]table locked")
}

// Unlocks the table by releasing all the row locks.
func (d *datastore) unlockTable() {
	log.Println("[memds]attempting to release table lock")

	// release all the row locks
	// we need to unlock in decending order of rows to prevent deadlock
	for i := len(d.locks) - 1; i >= 0; i-- {
		d.locks[i].Unlock()
	}
	log.Println("[memds]table unlocked")
}

// List all the Accounts in the datastore.
//
// Note: Expensive opertion. Datastore is locked until
// a copy of all the accounts is completed.
func (d *datastore) List() []ds.Account {
	log.Printf("[memds]List() called")

	// to prevent concurrent access to the accounts data
	// lock the whole table until we are done
	d.lockTable()

	// make a copy of all the accounts
	dst := make([]ds.Account, len(d.accounts))
	copy(dst, d.accounts)
	log.Println("[memds]List: copy complete")

	// we are done, release the table
	d.unlockTable()

	log.Println("[memds]returing from List()")
	return dst
}

// Get the Account details for the given account-id.
//
// Returns error if an Account with such id does not exist.
func (d *datastore) Get(id string) (ds.Account, error) {
	log.Printf("[memds]Get() called with id: %v\n", id)

	// find the location of the Account given its id using index
	i, ok := d.index[id]
	if !ok {
		log.Printf("[memds]Get: account with id: %v does not exist\n", id)
		return ds.Account{}, fmt.Errorf("account with id: %v does not exist", id)
	}

	// lock this Account to prevent concurrent access
	d.locks[i].Lock()
	defer d.locks[i].Unlock()

	// return the copy of Account details
	return ds.Account(d.accounts[i]), nil
}

// Transfer amount from and to the specified accounts.
//
// Returns transaction-id and account balance for from-account on success.
// Returns error is any of the from/to account id is invalid or
// the available balance in the from account is insufficient to do the transfer
func (d *datastore) Transfer(from string, to string, amount float64) (uint64, float64, error) {
	log.Printf("[memds]Transfer() called with from: %v, to: %v, amount: %v\n", from, to, amount)

	// find the location of the from Account given its id using index
	si, ok := d.index[from] // si - source index
	if !ok {
		log.Printf("[memds]Transfer: from account with id: %v does not exist\n", from)
		return 0, 0, fmt.Errorf("from account with id: %v does not exist", from)
	}

	// find the location of the to Account given its id using index
	di, ok := d.index[to] // di - destination index
	if !ok {
		log.Printf("[memds]Transfer: to account with id: %v does not exist\n", to)
		return 0, 0, fmt.Errorf("to account with id: %v does not exist", to)
	}

	// both from and to accounts cannot be same
	if si == di {
		log.Printf("[memds]Transfer: from account id: %s and to accound id: %s are same\n", from, to)
		return 0, 0, fmt.Errorf("from account id: %s and to accound id: %s are same", from, to)
	}

	// lock both from and to accounts to prevent concurrent access
	// lock these accounts in ascending order of indexes to prevent dead lock
	var a1, a2 int // lock a1 account first and then a2
	if si < di {
		a1 = si
		a2 = di
	} else {
		a1 = di
		a2 = si
	}
	d.locks[a1].Lock()
	d.locks[a2].Lock()
	defer d.locks[a2].Unlock()
	defer d.locks[a1].Unlock()

	// check if we have sufficient funds
	if (d.accounts[si].Balance - amount) < 0 {
		log.Printf("[memds]Transfer: account id: %s does not have sufficient funds, available balance: %v\n", from, d.accounts[si].Balance)
		return 0, 0, fmt.Errorf("account id: %v does not have sufficient funds, available balance: %v", from, d.accounts[si].Balance)
	}

	// add a transaction entry
	d.tlock.Lock()
	t := transaction{
		tid:    d.nextTid,
		date:   time.Now(),
		from:   from,
		to:     to,
		amount: amount,
	}
	d.nextTid += 1
	d.transactions = append(d.transactions, t)
	d.tlock.Unlock()

	// do the transfer
	d.accounts[si].Balance -= amount
	d.accounts[di].Balance += amount

	log.Printf("[memds]returning from Transfer() with tid: %v, balance: %v\n", t.tid, d.accounts[si].Balance)
	return t.tid, d.accounts[si].Balance, nil
}

// end-of-file
