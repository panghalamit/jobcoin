package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/panghalamit/jobcoin/pb"
	"google.golang.org/grpc"
)

type freshAccounts struct {
	accounts []string
}

func getAddrsFromJsonFile(filePath string) ([]string, error) {
	var addrs freshAccounts
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err1 := json.Unmarshal(jsonBytes, &addrs)
	if err1 != nil {
		return nil, err1
	}
	return addrs.accounts, nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ./client <MixerRpcAddr> <pathToFreshAddrJson>")
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
	// Create a KvStore client
	mxc := pb.NewMixerClient(conn)

	addrlist, err := getAddrsFromJsonFile(os.Args[2])
	addrs := make([]*pb.Address, len(addrlist))
	for i := 0; i < len(addrlist); i++ {
		addrs[i] = &pb.Address{Address: addrlist[i]}
	}

	res, err := mxc.Mix(context.Background(), &pb.QueryDepositAddress{Addrs: addrs, Value: int64(200)})
	if err != nil {
		log.Fatalf("%v/n", err)
	}
	log.Printf("depsit address received from mixer %v\n", res.GetDepositAddr().Address)
}
