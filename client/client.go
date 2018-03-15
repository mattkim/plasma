package client

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/contracts"
	"github.com/kyokan/plasma/util"
	"github.com/pborman/uuid"
	"github.com/urfave/cli"
)

const (
	SIGN_PASSPHRASE     = "test"
	DEPOSIT_FILTER      = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	DEPOSIT_DESCRIPTION = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Deposit","type":"event"}]`
)

type DepositEvent struct {
	Sender common.Address
	Value  *big.Int
}

func Main(c *cli.Context) {
	fmt.Println("Hello World")

	contractAddress := c.GlobalString("contract-addr")
	nodeUrl := c.GlobalString("node-url")

	keystoreDir := "/Users/mattkim/geth/chain/keystore"
	keystoreFile := "/Users/mattkim/geth/chain/keystore/UTC--2018-03-13T17-33-34.839516799Z--44a5cae1ebd47c415630da1e2131b71d1f2f5803"
	userAddress := "44a5cae1ebd47c415630da1e2131b71d1f2f5803"
	keyWrapper := getFromKeyStore(userAddress, keystoreDir, keystoreFile)

	// makeKey()
	// getCurrentChildBlock(nodeUrl, contractAddress)
	// getPendingNonce(nodeUrl, userAddress)
	// filter(nodeUrl, contractAddress, depositFilter, depositDescription)
	// plasmaFilter(nodeUrl, contractAddress)

	for i := uint64(0); i < uint64(1); i++ {
		deposit(nodeUrl, contractAddress, userAddress, keyWrapper.PrivateKey, i)
		time.Sleep(2000 * time.Millisecond)
	}

	// slipperWatchFilter(nodeUrl, contractAddress)
	plasmaWatchFilter(nodeUrl, contractAddress)

	for i := uint64(1); i < uint64(2); i++ {
		deposit(nodeUrl, contractAddress, userAddress, keyWrapper.PrivateKey, i)
		time.Sleep(2000 * time.Millisecond)
	}

	select {}
}

func slipperWatchFilter(nodeUrl string, contractAddress string) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	query := ethereum.FilterQuery{
		FromBlock: nil,
		ToBlock:   nil,
		Topics:    [][]common.Hash{{common.HexToHash(DEPOSIT_FILTER)}},
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	// ch := make(chan types.Log)
	ch := make(chan types.Log, 2)

	s, err := conn.SubscribeFilterLogs(context.Background(), query, ch)

	if err != nil {
		panic(err)
	}

	// Errors on the channel.
	errChan := s.Err()

	go func() {
		for {
			select {
			case err := <-errChan:
				// TODO: log error for real.
				log.Printf("Logs subscription error: %v", err)
				break
			case event := <-ch:
				fmt.Println("get something on the channel")
				fmt.Println(event)
			}
		}
	}()
}

func plasmaWatchFilter(nodeUrl string, contractAddress string) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	// Instantiate the contract and display its name
	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	opts := bind.WatchOpts{
		Start:   nil, // latest block
		Context: context.Background(),
	}

	ch := make(chan *contracts.PlasmaDeposit)

	s, err := plasma.WatchDeposit(&opts, ch)

	if err != nil {
		panic(err)
	}

	go func() {
		for true {
			select {
			case err := <-s.Err():
				fmt.Println("**** found error!")
				fmt.Println(err)
			case event := <-ch:
				fmt.Println("**** found event!")
				fmt.Println(event.Sender.Hex())
				fmt.Println(event.Value)
				break
			}
		}
	}()
}

func plasmaFilter(nodeUrl string, contractAddress string) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	// Instantiate the contract and display its name
	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	opts := bind.FilterOpts{
		Start:   0x0B,
		End:     nil,
		Context: context.Background(),
	}

	itr, err := plasma.FilterDeposit(&opts)

	if err != nil {
		panic(err)
	}

	next := true
	length := 0

	for next {
		if itr.Event != nil {
			fmt.Println(itr.Event.Sender.Hex())
			fmt.Println(itr.Event.Value)
		}
		next = itr.Next()
		length++
	}

	fmt.Printf("%d logs found \n", length)
}

func filter(nodeUrl string, contractAddress string, depositFilter string, depositDescription string) {
	client, err := ethclient.Dial(nodeUrl)

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(1),
		ToBlock:   nil,
		Topics:    [][]common.Hash{{common.HexToHash(depositFilter)}},
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	logs, err := client.FilterLogs(context.Background(), query)

	if err != nil {
		panic(err)
	}

	depositAbi, err := abi.JSON(strings.NewReader(depositDescription))

	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d logs\n", len(logs))

	for _, log := range logs {
		event := DepositEvent{}
		err = depositAbi.Unpack(&event, "Deposit", log.Data)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Received %s wei deposit from %s.\n",
			event.Value.String(),
			util.AddressToHex(&event.Sender))
	}
}

func deposit(nodeUrl string, contractAddress string, userAddress string, privateKeyECDSA *ecdsa.PrivateKey, value uint64) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	// Instantiate the contract and display its name
	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	auth := bind.NewKeyedTransactor(privateKeyECDSA)

	if err != nil {
		panic(err)
	}

	auth.Value = new(big.Int).SetUint64(value)

	tx, err := plasma.Deposit(auth)

	if err != nil {
		log.Fatalf("Failed to deposit: %v", err)
	}

	fmt.Printf("Deposit pending: 0x%x\n", tx.Hash())
}

func getCurrentChildBlock(nodeUrl string, contractAddress string) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	// Instantiate the contract and display its name
	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	fmt.Println(plasma)

	res, err := plasma.CurrentChildBlock(nil)

	if err != nil {
		log.Fatalf("Failed to get current child block: %v", err)
	}

	fmt.Println("Deposit result:", res)
}

func getKey(privateKey string) *keystore.Key {
	// TODO: is random id okay?
	privateKeyECDSA := getPrivateKeyECDSA(privateKey)

	id := uuid.NewRandom()

	return &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
}

func getPrivateKeyECDSA(privateKey string) *ecdsa.PrivateKey {
	key, err := hex.DecodeString(privateKey)
	if err != nil {
		panic(err)
	}

	privateKeyECDSA, err := crypto.ToECDSA(key)

	if err != nil {
		panic(err)
	}

	return privateKeyECDSA
}

func makeKey() {
	privateKey := "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"

	privateKeyECDSA := getPrivateKeyECDSA(privateKey)

	fmt.Println(privateKeyECDSA)

	publicKey := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	fmt.Println(publicKey)
	fmt.Println(publicKey.Hex())
}

func getFromKeyStore(addr string, keystoreDir string, keystoreFile string) *keystore.Key {
	// Init a keystore
	ks := keystore.NewKeyStore(
		keystoreDir,
		keystore.LightScryptN,
		keystore.LightScryptP)

	// Create account definitions
	fromAccDef := accounts.Account{
		Address: common.HexToAddress(addr),
	}

	// Find the signing account
	signAcc, err := ks.Find(fromAccDef)
	if err != nil {
		fmt.Println("account keystore find error:")
		panic(err)
	}

	// Unlock the signing account
	errUnlock := ks.Unlock(signAcc, SIGN_PASSPHRASE)
	if errUnlock != nil {
		fmt.Println("account unlock error:")
		panic(err)
	}

	// Open the account key file
	keyJson, readErr := ioutil.ReadFile(keystoreFile)
	if readErr != nil {
		fmt.Println("key json read error:")
		panic(readErr)
	}

	// Get the private key
	keyWrapper, keyErr := keystore.DecryptKey(keyJson, SIGN_PASSPHRASE)
	if keyErr != nil {
		fmt.Println("key decrypt error:")
		panic(keyErr)
	}

	return keyWrapper
}

func getPendingNonce(nodeUrl string, userAddress string) uint64 {
	client, err := rpc.Dial(nodeUrl)
	if err != nil {
		panic(err)
	}

	var result string

	err = client.CallContext(
		context.Background(),
		&result,
		"eth_getTransactionCount",
		common.HexToAddress(userAddress),
		"pending",
	)

	if err != nil {
		panic(err)
	}

	return hexToUint64(result)
}

func hexToUint64(hexStr string) uint64 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 64)
	return uint64(result)
}
