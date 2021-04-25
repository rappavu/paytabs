// Implements CLI command 'bank' to start an in-memory datastore server.
//
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"paytabs/internal/server"
)

// Prints the usage information for the 'bank' command.
//
func printUsage() {
	fmt.Println(`
Usage:
	bank <port> <datafile> [logfile]
	port     - port number for the REST service to listen to.
	datafile - path to json file containing account details to initialize the in-memory datastore.
	logfile  - optional path to server log file, when ommited stdout will be used.
	`)
}

// bank command
//
// Validates the commandline arguments and starts the in-memory datastore server.
func main() {
	// validate command line arguments
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	// get port number
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: error getting bind port number from the command line argument - %v\n", err.Error())
		os.Exit(1)
	}
	if port < 0 {
		fmt.Printf("ERROR: bind port number needs to be a positive value")
		os.Exit(1)
	}

	// get path to data file
	file := os.Args[2]

	// get path to server log file, if present
	if len(os.Args) > 3 {
		// open the file for writing
		fp, err := os.OpenFile(os.Args[3], os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("ERROR: error opening log file: %v - %v\n", os.Args[3], err.Error())
			os.Exit(1)
		}

		// set the server log outputs to the given file
		log.SetOutput(fp)
	}

	// initialize server
	fmt.Printf("Initializing in-memory datastore server using file %v\n", file)
	srv, err := server.New(uint(port), file)
	if err != nil {
		log.Fatalf("Failed to start server - %v\n", err.Error())
	}

	// start the server
	fmt.Printf("Server Ready. Listening at http://localhost:%v/\n", srv.Port)
	log.Fatal(srv.Start())
}

// end-of-file
