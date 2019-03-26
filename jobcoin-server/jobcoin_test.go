package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/panghalamit/jobcoin/pb"
	"google.golang.org/grpc"
)

func startJobcoin() {
	// Get hostname
	name, err := os.Hostname()
	if err != nil {
		// Without a host name we can't really get an ID, so die.
		log.Fatalf("Could not get hostname")
	}
	id := fmt.Sprintf("%s:%d", name, 3001)
	log.Printf("Starting jobcoin with ID %s", id) // Create socket that listens on the supplied port
	c, err := net.Listen("tcp", ":3001")
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()
	blockchain, err := InitJobcoin(1)
	if err != nil {
		log.Fatalf("failed to initialize Jobcoin %v", err)
	}
	pb.RegisterJobcoinServer(s, blockchain)
	log.Printf("Going to listen on port %v", "3001")
	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
	log.Printf("Done listening")
}

func connectToJobcoin() (pb.JobcoinClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 5000 * time.Millisecond
	conn, err := grpc.Dial("0:3001", grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewJobcoinClient(nil), err
	}
	return pb.NewJobcoinClient(conn), nil
}

func TestGenerateAddress(t *testing.T) {
	go startJobcoin()

	// jobcoin Client
	cl, err := connectToJobcoin()
	if err != nil {
		t.Errorf("%v", err)
	}
	//time.Sleep(time.Duration(5000) * time.Millisecond)
	res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Printf("%v/n", res.GetAddr().Address)
	if err != nil {
		t.Errorf("%v", err)
	} else if res.GetAddr().Address == "" {
		t.Errorf("Address returned nil, expected a 20 byte random string")
	} else {
		log.Printf("Address returned %v\n", res.GetAddr().Address)
	}
}

func TestFundAddress(t *testing.T) {
	go startJobcoin()

	// jobcoin Client
	cl, err := connectToJobcoin()
	if err != nil {
		t.Errorf("%v\n", err)
	}
	res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	log.Printf("%v\n", res.GetAddr().Address)
	if err != nil {
		t.Errorf("%v\n", err)
	} else if res.GetAddr().Address == "" {
		t.Errorf("Address returned nil, expected a 20 byte random string")
	} else {
		log.Printf("Address returned %v\n", res.GetAddr().Address)
	}
	addr := res.GetAddr().Address
	bal := int64(1000)
	res, err = cl.FundAddress(context.Background(), &pb.AddressBalance{Address: addr, Balance: bal})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 1000 {
		t.Errorf("FundAddress failure, expected balance %v, got %v\n", 1000, res.GetBal().Balance)
	}
}

func TestGetBalance(t *testing.T) {
	go startJobcoin()

	// jobcoin Client
	cl, err := connectToJobcoin()
	if err != nil {
		t.Errorf("%v\n", err)
	}
	res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	log.Printf("%v\n", res.GetAddr().Address)
	if err != nil {
		t.Errorf("%v\n", err)
	} else if res.GetAddr().Address == "" {
		t.Errorf("Address returned nil, expected a 20 byte random string")
	} else {
		log.Printf("Address returned %v\n", res.GetAddr().Address)
	}
	addr := res.GetAddr().Address
	bal := int64(1000)
	res, err = cl.FundAddress(context.Background(), &pb.AddressBalance{Address: addr, Balance: bal})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 1000 {
		t.Errorf("FundAddress failure, expected balance %v, got %v\n", 1000, res.GetBal().Balance)
	}
	res, err = cl.GetBalance(context.Background(), &pb.Address{Address: addr})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 1000 {
		t.Errorf("GetBalance failure, expected balance %v, got %v\n", 1000, res.GetBal().Balance)

	}
}

func TestTransfer(t *testing.T) {
	go startJobcoin()

	// jobcoin Client
	cl, err := connectToJobcoin()
	if err != nil {
		t.Errorf("%v\n", err)
	}
	res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	from := res.GetAddr().Address
	res, err = cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	to := res.GetAddr().Address
	if err != nil {
		t.Errorf("%v\n", err)
	}
	bal := int64(1000)
	res, err = cl.FundAddress(context.Background(), &pb.AddressBalance{Address: from, Balance: bal})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 1000 {
		t.Errorf("FundAddress failure, expected balance %v, got %v\n", 1000, res.GetBal().Balance)
	}

	res, err = cl.Transfer(context.Background(), &pb.TransferBalance{From: from, To: to, Balance: int64(500)})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	res, err = cl.GetBalance(context.Background(), &pb.Address{Address: to})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 500 {
		t.Errorf("GetBalance failure, expected balance %v, got %v\n", 500, res.GetBal().Balance)
	}
	res, err = cl.GetBalance(context.Background(), &pb.Address{Address: from})
	if err != nil {
		t.Errorf("%v\n", err)
	}
	if res.GetBal().Balance != 499 {
		t.Errorf("GetBalance failure, expected balance %v, got %v\n", 499, res.GetBal().Balance)
	}
}
