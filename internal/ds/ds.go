// Defines the interface to be supported by a Datastore.
//
package ds

type Account struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance,string"`
}

type Datastore interface {
	List() []Account
	Get(string) (Account, error)
	Transfer(from string, to string, amount float64) (uint64, float64, error)
}

// end-of-file
