## Implementation of mixer in golang

There are two major components

1. jobcoin-server: blockchain rpc server, that can generate new addresses, fund existing addresses, transfer funds between addresses and can check balance of a given address

2. mixer: mixer rpc server, that listens for mix requests from clients, returns a deposit address.

Mixer uses jobcoin rpc apis to poll jobcoin if funds to deposit addresses are transferred. If so moves funds to house address. payouts are made once there is good mix of users with similar payout amounts. 

3. clients : jobcoin and mixer clients to interact with server apis


## Running instructions

### build jobcoin

`$ cd go/src/panghalamit/jobcoin/jobcoin-server`

`$ go build .`

### test jobcoin
run `go test -v`  inside jobcoin-server 

### run jobcoin
example: 

`$ ./jobcoin-server --port 3000 --txFee 1`

### build mixer

`$ cd go/src/panghalamit/jobcoin/mixer`

`go build .`

### test mixer
make sure jobcoin is running on "0:3000" or pass appropriate address to function connectToJobcoin in TestMix

run  `go test -v` inside mixer 


### run mixer 

example:

`$ ./mixer --mixerFee 1 --port 3001 --jobcoinAddr "0:3000"`

## Using clients

### build clients

#### build jobcoin client

`$ cd go/src/panghalamit/jobcoin/client/jobcoin`

`$ go build .`

#### run jobcoin client

1. To query balance balance of an address 

`./jobcoin <jobcoinRPCAddr> <cmd> <addr>`

example :

`$ ./jobcoin "0:3000" 1  "qL3UDfyEcQAaUJ1lDSIZ" `

2. To generate new address

`./jobcoin <jobcoinRPCAddr> <cmd>`

example:

`$ ./jobcoin "0:3000" 2`

3. To fund a existing address

`./jobcoin <jobcoinRPCAddr> <cmd> <addr> <val>`

example:

`$ ./jobcoin "0:3000" 3 "KC4yYEwe5YJPuNAPeeMq" 1000`

4. To transfer funds

`./jobcoin <jobcoinRPCAddr <cmd> <from> <to> <val>`

example:

`$ ./jobcoin "0:3000" 4 "KC4yYEwe5YJPuNAPeeMq" "qL3UDfyEcQAaUJ1lDSIZ" 500`

#### build mixer client

`$ cd go/src/panghalamit/jobcoin/client/mixer`

`$ go build .`

#### run mixer client

send mix request and get a deposit address. fresh addresses by client are put in a json file.

`./mixer <mixerRpcAddr> <pathToFreshAddrJsonfile>`

example

`$ ./mixer "0:3000"  fresh-addresses.json `


