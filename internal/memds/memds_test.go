package memds

import (
	"testing"
)

func TestLoad(t *testing.T) {
	// test loading the sample json accounts file paytabs/data/accounts-mock.json
	d, err := Load("../../data/accounts-mock.json")
	if err != nil {
		t.Error("failed to load json file")
	}
	if d == nil {
		t.Error("failed to get datastore pointer")
	}
	if d != nil && len(d.accounts) != 500 {
		t.Errorf("expecting 500 records, got %v\n", len(d.accounts))
	}
}

func TestList(t *testing.T) {

}

func TestTransfer(t *testing.T) {

}

func TestGet(t *testing.T) {

}
