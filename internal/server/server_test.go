package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"paytabs/internal/ds"
	"reflect"
	"testing"
)

const datafile string = "../../data/accounts-mock.json"

var gSrv *DataServer
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
		fmt.Println("Server test setup failed. Server tests skipped.")
		return
	}

	// initialize server
	s, err := New(8080, datafile)
	if err != nil {
		fmt.Printf("Error initializing datastore server using file: %s", datafile)
		fmt.Println("Server test setup failed. Server tests skipped.")
		return
	}
	if s == nil {
		fmt.Println("server.New() failed. Server tests skipped.")
		return
	}
	gSrv = s

	// run the tests
	os.Exit(m.Run())
}

func TestGetList(t *testing.T) {
	// setup request
	req := httptest.NewRequest("GET", "http://localhost:8080/list/", nil)
	w := httptest.NewRecorder()

	// send request
	gSrv.listHandler(w, req)

	// validate response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expecting Status  %v, received %v\n", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expecting Content-Type: application/json, received %v\n", resp.Header.Get("Content-Type"))
	}

	// decode json data
	var accts []ds.Account
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&accts); err != nil {
		t.Fatal("Error decoding json data")
	}

	if !reflect.DeepEqual(gAccounts, accts) {
		t.Fatal("Received []Accounts data does not match with the expected")
	}
}

func TestPostTransfer(t *testing.T) {
	// setup request
	amount := 0.5
	td := TranferDetail{gAccounts[0].Id, gAccounts[1].Id, amount}
	jbytes, _ := json.Marshal(td)
	req := httptest.NewRequest("POST", "http://localhost:8080/transfer/", bytes.NewReader(jbytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// send request
	gSrv.transferHandler(w, req)

	// validate response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expecting Status  %v, received %v\n", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expecting Content-Type: application/json, received %v\n", resp.Header.Get("Content-Type"))
	}

	// decode json data
	var tr TranferResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&tr); err != nil {
		t.Fatal("Error decoding json data")
	}

	expectedBalance := gAccounts[0].Balance - amount
	if expectedBalance != tr.Balance {
		t.Fatalf("Expecting balance: %v, but received %v\n", expectedBalance, tr.Balance)
	}
}
