package server

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strings"

	"paytabs/internal/ds"
	"paytabs/internal/memds"
)

// structure to store server data
type DataServer struct {
	Port uint           // listening port for the server
	mux  *http.ServeMux // url path handler mux
	data ds.Datastore   // datastore for Accounts
}

// GET /list/ Handler
//
func (s *DataServer) listHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("[%v][%v][%v]received request\n", req.RemoteAddr, req.Method, req.URL.Path)

	// reject if this is not a GET
	if req.Method != http.MethodGet {
		log.Printf("[%v][%v][%v]expecting method GET, got %v\n", req.RemoteAddr, req.Method, req.URL.Path, req.Method)
		http.Error(w, fmt.Sprintf("expecting method GET, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	// get the list of all account details
	accts := s.data.List()
	log.Printf("[%v][%v][%v]received copy of %v accounts from the datastore\n", req.RemoteAddr, req.Method, req.URL.Path, len(accts))

	// write the acct details
	js, err := json.Marshal(accts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	log.Printf("[%v][%v][%v]reply sent\n", req.RemoteAddr, req.Method, req.URL.Path)
}

// GET /account/<id> Handler
//
func (s *DataServer) getAccountHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("[%v][%v][%v]received request\n", req.RemoteAddr, req.Method, req.URL.Path)

	// reject if this is not a GET
	if req.Method != http.MethodGet {
		log.Printf("[%v][%v][%v]expecting method GET, got %v\n", req.RemoteAddr, req.Method, req.URL.Path, req.Method)
		http.Error(w, fmt.Sprintf("expecting method GET, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	// get the account-id
	path := strings.Trim(req.URL.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		log.Printf("[%v][%v][%v]expecting /account/<id>, unable to find account-id in the request\n", req.RemoteAddr, req.Method, req.URL.Path)
		http.Error(w, "expecting /account/<id>, unable to find account-id in the request", http.StatusBadRequest)
		return
	}
	id := pathParts[1]

	// get the account details
	acct, err := s.data.Get(id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[%v][%v][%v]got account details for id: %v from datastore\n", req.RemoteAddr, req.Method, req.URL.Path, id)

	// write the acct details
	js, err := json.Marshal(acct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	log.Printf("[%v][%v][%v]reply sent\n", req.RemoteAddr, req.Method, req.URL.Path)
}

// POST /transfer/ Handler
//
func (s *DataServer) transferHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("[%v][%v][%v]received request\n", req.RemoteAddr, req.Method, req.URL.Path)

	// reject if this is not a POST
	if req.Method != http.MethodPost {
		log.Printf("[%v][%v][%v]expecting method POST, got %v\n", req.RemoteAddr, req.Method, req.URL.Path, req.Method)
		http.Error(w, fmt.Sprintf("expecting method POST, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	// structure for POST data expected from client
	type TranferDetail struct {
		FromId string  `json:"from_id"`
		ToId   string  `json:"to_id"`
		Amount float64 `json:"amount"`
	}

	// extract the fund transfer details from the POST request
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Printf("[%v][%v][%v]error retrieving Content-Type\n", req.RemoteAddr, req.Method, req.URL.Path)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		log.Printf("[%v][%v][%v]unexpected Content-Type %v\n", req.RemoteAddr, req.Method, req.URL.Path, mediatype)
		http.Error(w, "require application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	// decode the json data for fund transfer
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	var td TranferDetail
	if err := decoder.Decode(&td); err != nil {
		log.Printf("[%v][%v][%v]error decoding json data - %v\n", req.RemoteAddr, req.Method, req.URL.Path, err.Error())
		http.Error(w, fmt.Sprintf("error decoding json data - %v", err.Error()), http.StatusBadRequest)
		return
	}
	log.Printf("[%v][%v][%v]from_id: %v, to_id: %v, amount: %v\n", req.RemoteAddr, req.Method, req.URL.Path, td.FromId, td.ToId, td.Amount)

	// validate data, make sure amount is a +ve value
	if td.Amount < 0 {
		log.Printf("[%v][%v][%v]fund transfer failed, transfer amount cannot be a negative value\n", req.RemoteAddr, req.Method, req.URL.Path)
		http.Error(w, "fund transfer failed, transfer amount cannot be a negative value", http.StatusBadRequest)
		return
	}

	// perform fund transfer
	tid, balance, err := s.data.Transfer(td.FromId, td.ToId, td.Amount)
	if err != nil {
		log.Printf("[%v][%v][%v]fund transfer failed - %v\n", req.RemoteAddr, req.Method, req.URL.Path, err.Error())
		http.Error(w, fmt.Sprintf("fund transfer failed - %v", err.Error()), http.StatusInternalServerError)
		return
	}
	log.Printf("[%v][%v][%v]fund transfer completed in datastore with tid: %v, balance: %v\n", req.RemoteAddr, req.Method, req.URL.Path, tid, balance)

	// response data sent to the client on successful transfer
	type TranferResponse struct {
		TransactionId uint64  `json:"transaction_id"`
		Balance       float64 `json:"balance"`
	}
	tr := TranferResponse{tid, balance}

	// write response to client
	js, err := json.Marshal(tr)
	if err != nil {
		log.Printf("[%v][%v][%v]json marshall failed - %v\n", req.RemoteAddr, req.Method, req.URL.Path, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	log.Printf("[%v][%v][%v]transfer completed successfully, reply sent\n", req.RemoteAddr, req.Method, req.URL.Path)
}

// Initialize Server
//
func New(port uint, filename string) (*DataServer, error) {
	// initialize in-memory datastore
	log.Printf("[server]initializing in-memory datastore using file: %v\n", filename)
	d, err := memds.Load(filename)
	if err != nil {
		return nil, err
	}

	// instantiate DataServer
	srv := new(DataServer)
	srv.Port = port
	srv.data = d
	log.Println("[server]datastore initialization complete")

	// initialize ServeMux and add handlers
	log.Println("[server]registering handlers")
	mux := http.NewServeMux()

	mux.HandleFunc("/list/", srv.listHandler)
	log.Println("[server]registered handler for GET /list/")

	mux.HandleFunc("/transfer/", srv.transferHandler)
	log.Println("[server]registered handler for POST /transfer/")

	mux.HandleFunc("/account/", srv.getAccountHandler)
	log.Println("[server]registered handler for GET /account/<id>")

	srv.mux = mux
	log.Println("[server]handler registration complete")

	return srv, nil
}

// Start the server. Listen and Serve.
//
func (s *DataServer) Start() error {
	return http.ListenAndServe(fmt.Sprintf("localhost:%d", s.Port), s.mux)
}

// end-of-file
