package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/panghalamit/jobcoin/pb"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

const POLLING_INTERVAL time.Duration = time.Duration(100) * time.Millisecond

const MIN_ACCOUNT_MIX = 4

type PollResultType struct {
	depositAddress  string
	fundTransferred bool
}

type SendPayoutType struct {
	payoutAddr string
	funds      int64
}

type MixRequestInput struct {
	args     *pb.QueryDepositAddress
	response chan pb.DepositAddress
}

type MixerService struct {
	houseAddress           string
	depositToPayoutAddrs   map[string][]string
	depositToPayoutBalance map[string]int64
	depositAddressesToPoll map[string]int64
	requestCh              chan MixRequestInput
}

type PendingMix struct {
	amountsToAddrs map[int64][]string
}

func (mixer *MixerService) Mix(ctx context.Context, queryDepositArgs *pb.QueryDepositAddress) (*pb.DepositAddress, error) {
	c := make(chan pb.DepositAddress)
	mixer.requestCh <- MixRequestInput{args: queryDepositArgs, response: c}
	result := <-c
	return &result, nil
}

func RunMixerServer(mixer *MixerService, port int) {
	// Convert port to a string form
	portString := fmt.Sprintf(":%d", port)
	// Create socket that listens on the supplied port
	c, err := net.Listen("tcp", portString)
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()

	pb.RegisterMixerServer(s, mixer)
	log.Printf("Going to listen on port %v", port)

	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func InitMixer(cl pb.JobcoinClient) (*MixerService, error) {
	// generate a house address
	result, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
	if err != nil {
		return &MixerService{}, nil
	}
	houseAddress := result.GetAddr().Address

	depositToPayoutAddrs := make(map[string][]string)
	depositToPayoutBalance := make(map[string]int64)
	depositAddressesToPoll := make(map[string]int64)
	requestCh := make(chan MixRequestInput)

	return &MixerService{houseAddress, depositToPayoutAddrs, depositToPayoutBalance, depositAddressesToPoll, requestCh}, nil
}

func connectToBlockchain(jobcoinAddr string) (pb.JobcoinClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 5000 * time.Millisecond
	conn, err := grpc.Dial(jobcoinAddr, grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewJobcoinClient(nil), err
	}
	return pb.NewJobcoinClient(conn), nil
}

func serve(blockchainAddr string, mixerFee int64, clientPort int) {

	//connect to blockchain client
	blockchainCl, err := connectToBlockchain(blockchainAddr)
	if err != nil {
		log.Fatalf("Mixer failure, can't connect to blockchain api server %v", err)
	}
	mixer, err := InitMixer(blockchainCl)
	if err != nil {
		log.Fatalf("Mixer failure: failed to obtain house address while initializing %v", err)
	}
	pollResultChan := make(chan PollResultType, 100)
	mixEvent := make(chan int)
	go RunMixerServer(mixer, clientPort)
	pollingTimer := time.NewTimer(time.Second * 2)
	// event to transfer to house addresss
	type transferToHouseAddrEvent struct {
		depositAddr string
		value       int64
	}
	addToPendingMix := make(chan transferToHouseAddrEvent)
	payoutCh := make(chan SendPayoutType, 100)
	pendingMix := PendingMix{make(map[int64][]string)}
	mixAccounts := make(map[int64]int)
	for {
		select {
		case <-pollingTimer.C:
			// poll blockchain for deposit addresses
			// check if pending deposit addresses received funds from user
			log.Printf("Polling timeout fired, poll blockchain to see if any of deposit Address received funds\n")
			for k, v := range mixer.depositAddressesToPoll {
				log.Printf("Polling addr %v \n", k)
				go func(cl pb.JobcoinClient, addr string, val int64) {
					res, err := cl.GetBalance(context.Background(), &pb.Address{Address: addr})
					if err == nil {
						bal := res.GetBal().Balance
						if bal >= val {
							pollResultChan <- PollResultType{addr, true}
						} else {
							pollResultChan <- PollResultType{addr, false}
						}
					}
				}(blockchainCl, k, v)
			}
			pollingTimer.Reset(POLLING_INTERVAL)
		case mixReq := <-mixer.requestCh:
			// mix request from client

			log.Printf("Mix request received from client\n")
			freshAddrs := make([]string, len(mixReq.args.Addrs))
			for k, v := range mixReq.args.Addrs {
				freshAddrs[k] = v.Address
			}
			value := mixReq.args.Value
			c := make(chan *pb.Result)
			go func(cl pb.JobcoinClient, resp chan *pb.Result) {
				res, err := cl.GetNewAddress(context.Background(), &pb.Empty{})
				if err != nil {
					c <- &pb.Result{Result: &pb.Result_Addr{Addr: &pb.Address{Address: ""}}}
				} else {
					c <- res
				}
			}(blockchainCl, c)
			log.Printf("Waiting for get fresh deposit address from blockchain\n")
			resp := <-c
			depositAddr := resp.GetAddr().Address
			if depositAddr != "" {
				mixer.depositToPayoutAddrs[depositAddr] = freshAddrs
				mixer.depositAddressesToPoll[depositAddr] = value
			}
			pollingTimer.Reset(POLLING_INTERVAL)
			log.Printf("Sending deposit address %v back to client", depositAddr)
			mixReq.response <- pb.DepositAddress{DepositAddr: &pb.Address{Address: depositAddr}}
		case pollRes := <-pollResultChan:
			// result of blockchain polling
			log.Printf("Received a single result of polling blockchain to check funds of address %v \n", pollRes.depositAddress)
			if pollRes.fundTransferred {
				// user transferred the the funds, can include user in the mix now.
				mixer.depositToPayoutBalance[pollRes.depositAddress] = mixer.depositAddressesToPoll[pollRes.depositAddress]
				delete(mixer.depositAddressesToPoll, pollRes.depositAddress)
				go func(c chan int) {
					c <- 1
				}(mixEvent)
			}
		case <-mixEvent:
			// send payment to house address and add to pending Mix
			log.Printf("Transferring funds to house address\n")
			for k, v := range mixer.depositToPayoutBalance {
				//
				go func(cl pb.JobcoinClient, from string, to string, val int64) {
					_, err := cl.Transfer(context.Background(), &pb.TransferBalance{From: from, To: to, Balance: val})
					if err == nil {
						//transfer successful, add to pending Mix, value minus mixer fee
						addToPendingMix <- transferToHouseAddrEvent{from, val - mixerFee}
					}
				}(blockchainCl, k, mixer.houseAddress, v)
			}
		case e := <-addToPendingMix:
			log.Printf("Add to pending mix after tarnsfer to house address was successful\n")
			amount := int64(e.value / int64(len(mixer.depositToPayoutAddrs[e.depositAddr])))
			_, ok := pendingMix.amountsToAddrs[amount]
			if !ok {
				pendingMix.amountsToAddrs[amount] = mixer.depositToPayoutAddrs[e.depositAddr]
			} else {
				pendingMix.amountsToAddrs[amount] = append(pendingMix.amountsToAddrs[amount], mixer.depositToPayoutAddrs[e.depositAddr]...)
			}
			delete(mixer.depositToPayoutAddrs, e.depositAddr)
			delete(mixer.depositToPayoutBalance, e.depositAddr)
			mixAccounts[amount] += 1
			log.Printf("Check conditions for payout of %v and pay if matched\n", amount)
			if mixAccounts[amount] > MIN_ACCOUNT_MIX {
				for _, v := range pendingMix.amountsToAddrs[amount] {
					payoutCh <- SendPayoutType{payoutAddr: v, funds: amount}
				}
				delete(pendingMix.amountsToAddrs, amount)
			}

		case pay := <-payoutCh:
			log.Printf("Paying out at address %v, amount %v\n", pay.payoutAddr, pay.funds)
			go func(cl pb.JobcoinClient, from string, to string, val int64) {
				_, err := cl.Transfer(context.Background(), &pb.TransferBalance{From: from, To: to, Balance: val})
				//TODO handle transfer failures
				if err != nil {
					log.Fatalf("Mixer service failure, payout failed to %v with %v", to, err)
				}
			}(blockchainCl, mixer.houseAddress, pay.payoutAddr, pay.funds)
		}
	}
}
