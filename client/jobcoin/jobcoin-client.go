package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/panghalamit/jobcoin/pb"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: ./client <JobcoinRpcAddr> <cmd>")
	}
	endpoint := os.Args[1]
	log.Printf("Connecting to %v", endpoint)
	// Connect to the server. We use WithInsecure since we do not configure https in this class.
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	//Ensure connection did not fail.
	if err != nil {
		log.Fatalf("Failed to dial GRPC server %v", err)
	}
	log.Printf("Connected")
	// Create a jobcoin client
	jbc := pb.NewJobcoinClient(conn)

	cmd := os.Args[2]

	if cmd == "1" {
		//Get Balance
		addr := os.Args[3]
		res, err := jbc.GetBalance(context.Background(), &pb.Address{Address: addr})
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Balance of jobcoin address %v is %v\n", addr, res.GetBal().Balance)
	} else if cmd == "2" {
		//Generate New Address
		res, err := jbc.GetNewAddress(context.Background(), &pb.Empty{})
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Generated new jobcoin address %v\n", res.GetAddr().Address)
	} else if cmd == "3" {
		//Fund Address
		addr := os.Args[3]
		fundsString := os.Args[4]
		funds, err := strconv.ParseInt(fundsString, 10, 64)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		res, err := jbc.FundAddress(context.Background(), &pb.AddressBalance{Address: addr, Balance: funds})
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Funded Address %v with %v and it now has balance %v", addr, funds, res.GetBal().Balance)
	} else if cmd == "4" {
		// Transfer balance
		from := os.Args[3]
		to := os.Args[4]
		funds, err := strconv.ParseInt(os.Args[5], 10, 64)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		_, err = jbc.Transfer(context.Background(), &pb.TransferBalance{From: from, To: to, Balance: funds})
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Transferred %v from jobcoin address %v to jobcoin Address %v", funds, from, to)
	} else {
		log.Printf("Unknown command\n")
	}
}
