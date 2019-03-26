package main

import (
	"log"
	"testing"
	"time"

	"github.com/panghalamit/jobcoin/pb"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

/*
func startJobcoin() {
	// Get hostname
	name, err := os.Hostname()
	if err != nil {
		// Without a host name we can't really get an ID, so die.
		log.Fatalf("Could not get hostname")
	}
	id := fmt.Sprintf("%s:%d", name, 3000)
	log.Printf("Starting jobcoin with ID %s", id) // Create socket that listens on the supplied port
	c, err := net.Listen("tcp", ":3000")
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()
	blockchain, err := jobcoin.InitJobcoin(1)
	if err != nil {
		log.Fatalf("failed to initialize Jobcoin %v", err)
	}
	pb.RegisterJobcoinServer(s, blockchain)
	log.Printf("Going to listen on port %v", "3000")
	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
	log.Printf("Done listening")
}*/

func connectToJobcoin(rpcAddr string) (pb.JobcoinClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 5000 * time.Millisecond
	conn, err := grpc.Dial(rpcAddr, grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewJobcoinClient(nil), err
	}
	return pb.NewJobcoinClient(conn), nil
}

func connectToMixer(rpcAddr string) (pb.MixerClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 5000 * time.Millisecond
	conn, err := grpc.Dial("0:3001", grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewMixerClient(nil), err
	}
	return pb.NewMixerClient(conn), nil
}

func getFreshAddresses(cl pb.JobcoinClient, n int) ([]*pb.Address, error) {
	addrs := make([]*pb.Address, n)
	for i := 0; i < n; i++ {
		res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
		if err == nil && res.GetAddr().Address != "" {
			addrs[i] = res.GetAddr()
		}
		log.Printf("%v\n", addrs[i])
	}
	return addrs, nil
}

func TestMix(t *testing.T) {
	// jobcoin Client
	jobcoinCl, err := connectToJobcoin("0:3000")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	// mixer Client
	mixerCl, err := connectToMixer("0:3001")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	toMix, err := getFreshAddresses(jobcoinCl, 5)
	if err != nil {
		t.Errorf("%v\n", err)
	}
	mixToFresh := make(map[string][]*pb.Address)
	for _, v := range toMix {
		_, err := jobcoinCl.FundAddress(context.Background(), &pb.AddressBalance{Address: v.Address, Balance: int64(1010)})
		if err != nil {
			t.Errorf("%v\n", err)
		}
		addrs, err := getFreshAddresses(jobcoinCl, 5)
		mixToFresh[v.Address] = addrs
		depositAddr, err := mixerCl.Mix(context.Background(), &pb.QueryDepositAddress{Addrs: addrs, Value: int64(1000)})
		if err != nil {
			t.Errorf("%v\n", err)
		}
		log.Printf("Transferring %v from %v to depositAddr: %v\n", 1001, v.Address, depositAddr.DepositAddr.Address)
		_, err = jobcoinCl.Transfer(context.Background(), &pb.TransferBalance{From: v.Address, To: depositAddr.DepositAddr.Address, Balance: int64(1001)})
		if err != nil {
			t.Errorf("%v\n", err)
		}
		time.Sleep(time.Duration(10) * time.Millisecond)

	}

	time.Sleep(time.Duration(10) * time.Second)

	for _, v := range mixToFresh {
		for _, k := range v {
			res, err := jobcoinCl.GetBalance(context.Background(), k)
			if err != nil {
				t.Errorf("%v\n", err)
			}
			if res.GetBal().Balance == 0 {
				t.Errorf("Expected mixer to send funds to fresh address %v, Got %v, Expected %v\n", k.Address, 0, 200)
			}
			log.Printf("Balance of fresh Address %v: %v\n", k.Address, res.GetBal().Balance)
		}
	}
}
