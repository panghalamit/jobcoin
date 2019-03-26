package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/panghalamit/jobcoin/pb"
	context "golang.org/x/net/context"
)

type Jobcoin struct {
	rwLock   sync.RWMutex
	balances map[string]int64
	fee      int64
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var randomS rand.Source

// initializes jobcoin
func InitJobcoin(txFee int64) (*Jobcoin, error) {
	balances := make(map[string]int64)
	randomS = rand.NewSource(time.Now().UnixNano())

	return &Jobcoin{sync.RWMutex{}, balances, txFee}, nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[randomS.Int63()%int64(len(letters))]
	}
	return string(b), nil
}
func (blockchain *Jobcoin) GetBalance(ctx context.Context, addr *pb.Address) (*pb.Result, error) {
	blockchain.rwLock.RLock()
	defer blockchain.rwLock.RUnlock()
	v, ok := blockchain.balances[addr.Address]
	if !ok {
		return &pb.Result{Result: &pb.Result_Bal{Bal: &pb.Balance{Balance: 0}}}, errors.New("Jobcoin: Address doesn't exist")
	}
	return &pb.Result{Result: &pb.Result_Bal{Bal: &pb.Balance{Balance: v}}}, nil
}

// Transfer method transfer funds from an sender account to receiver account
func (blockchain *Jobcoin) Transfer(ctx context.Context, transferArgs *pb.TransferBalance) (*pb.Result, error) {
	from := transferArgs.From
	to := transferArgs.To
	value := transferArgs.Balance
	blockchain.rwLock.Lock()
	defer blockchain.rwLock.Unlock()
	if _, ok := blockchain.balances[from]; !ok {
		return &pb.Result{Result: &pb.Result_S{S: &pb.Success{}}}, errors.New("Jobcoin Transfer failure: sender address doesn't exist")
	}
	if _, ok := blockchain.balances[to]; !ok {
		return &pb.Result{Result: &pb.Result_S{S: &pb.Success{}}}, errors.New("Jobcoin Transfer failure: receiver address doesn't exist")
	}
	if blockchain.balances[from] < blockchain.fee+value {
		return &pb.Result{Result: &pb.Result_S{S: &pb.Success{}}}, errors.New("Jobcoin Transfer failure: sender doesn't have enough funds")
	}
	blockchain.balances[from] -= (blockchain.fee + value)
	blockchain.balances[to] += value
	return &pb.Result{Result: &pb.Result_S{S: &pb.Success{}}}, nil
}

// GenerateNewAddress generates a fresh address for Jobcoin
func (blockchain *Jobcoin) GetNewAddress(ctx context.Context, in *pb.Empty) (*pb.Result, error) {
	addr, err := randomString(20)
	if err != nil {
		return &pb.Result{Result: &pb.Result_Addr{Addr: &pb.Address{Address: ""}}}, errors.New("Jobcoin GetNewAddress error: err while reading pseudorandom bytes")
	}
	blockchain.rwLock.Lock()
	defer blockchain.rwLock.Unlock()
	blockchain.balances[addr] = 0
	return &pb.Result{Result: &pb.Result_Addr{Addr: &pb.Address{Address: addr}}}, nil
}

// FundAddress funds a given blockchain address
func (blockchain *Jobcoin) FundAddress(ctx context.Context, fundAddrArg *pb.AddressBalance) (*pb.Result, error) {
	addr := fundAddrArg.Address
	value := fundAddrArg.Balance
	blockchain.rwLock.Lock()
	defer blockchain.rwLock.Unlock()
	if _, ok := blockchain.balances[addr]; !ok {
		return &pb.Result{Result: &pb.Result_Bal{Bal: &pb.Balance{Balance: 0}}}, errors.New("Jobcoin FundAddress failure: address doesn't exist")
	}
	if value < 0 {
		return &pb.Result{Result: &pb.Result_Bal{Bal: &pb.Balance{Balance: blockchain.balances[addr]}}}, errors.New("Jobcoin FundAddress failure: value < 0")
	}// initializes jobcoin
func InitJobcoin(txFee int64) (*Jobcoin, error) {
	balances := make(map[string]int64)
	randomS = rand.NewSource(time.Now().UnixNano())

	return &Jobcoin{sync.RWMutex{}, balances, txFee}, nil
}
	blockchain.balances[addr] += value
	return &pb.Result{Result: &pb.Result_Bal{Bal: &pb.Balance{Balance: blockchain.balances[addr]}}}, nil
}
