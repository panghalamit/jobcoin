package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/panghalamit/jobcoin/pb"
	"google.golang.org/grpc"
)

func main() {
	var clientPort int
	var txFee int64
	flag.IntVar(&clientPort, "port", 3000, "Port on which jobcoin server listen to requests")
	flag.Int64Var(&txFee, "txFee", 1, "transaction fee for jobcoin miners")
	flag.Parse()
	// Convert port to a string form
	portString := fmt.Sprintf(":%d", clientPort)
	// Create socket that listens on the supplied port
	c, err := net.Listen("tcp", portString)
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()
	blockchain, err := InitJobcoin(txFee)
	if err != nil {
		log.Fatalf("failed to initialize Jobcoin %v", err)
	}
	pb.RegisterJobcoinServer(s, blockchain)
	log.Printf("Going to listen on port %v", clientPort)
	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
	log.Printf("Done listening")
}
