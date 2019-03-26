package main

import (
	"flag"
)

func main() {
	var clientPort int
	var mixerFee int64
	var jobCoinAddr string
	flag.IntVar(&clientPort, "port", 3001, "Port on which jobcoin server listen to requests")
	flag.Int64Var(&mixerFee, "mixerFee", 1, "transaction fee for jobcoin miners")
	flag.StringVar(&jobCoinAddr, "jobcoinAddr", "http://localhost:3000", "rpc addr for jobcoin for querying")
	flag.Parse()
	serve(jobCoinAddr, mixerFee, clientPort)
}
