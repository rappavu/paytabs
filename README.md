# Bank REST Service

```
Getting the sources:
$ git clone https://github.com/rappavu/paytabs.git

Building the sources:
$ cd paytabs/cmd/bank
$ go build

This will build bank executable

Executing the command:
$ ./bank  // prints usage info

Usage:
        bank <port> <datafile> [logfile]
        port     - port number for the REST service to listen to.
        datafile - path to json file containing account details to initialize the in-memory datastore.
        logfile  - optional path to server log file, when ommited stdout will be used.

$ ./bank 8080 ../../data/accounts-mock.json  
Initializing in-memory datastore server using file ..\..\data\accounts-mock.json
2021/04/26 09:42:13 [server]initializing in-memory datastore using file: ..\..\data\accounts-mock.json
2021/04/26 09:42:13 [memds]loading data from file: ..\..\data\accounts-mock.json
2021/04/26 09:42:13 [memds]file read complete
2021/04/26 09:42:13 [memds]json data unmarshall complete
2021/04/26 09:42:13 [memds]indexing complete
2021/04/26 09:42:13 [memds]datastore initialization complete
2021/04/26 09:42:13 [server]datastore initialization complete
2021/04/26 09:42:13 [server]registering handlers
2021/04/26 09:42:13 [server]registered handler for GET /list/
2021/04/26 09:42:13 [server]registered handler for POST /transfer/
2021/04/26 09:42:13 [server]registered handler for GET /account/<id>
2021/04/26 09:42:13 [server]handler registration complete
Server Ready. Listening at http://localhost:8080/

--
This terminal will wait with the above message.

After you see "Server Ready" message, server is ready to receive and serve REST requests.

Supported REST API is mentioned below:
GET   /list/         : Returns json array of all accounts in the datastore
POST  /transfer/     : Used to transfer amount from one account to another
GET   /account/<id>  : Returns account details for the given <id>

Structure of data used for account details:
{
    "id": string,
    "name": string,
    "balance": string
}

Structure used by post data for transfer:
{
    "from_id": string,
    "to_id": string
    "amount": float64
}

Structure used by response data for transfer:
{
    "transaction_id": string
    "balance": float64
}

```

// end-of-file
