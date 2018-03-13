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

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/contracts"
	"github.com/kyokan/plasma/util"
	"github.com/pborman/uuid"
	"github.com/urfave/cli"
)

type DepositEvent struct {
	Sender common.Address
	Value  *big.Int
}

func Main(c *cli.Context) {
	fmt.Println("Hello World")
	// contractAddress := c.GlobalString("contract-addr")
	// nodeUrl := c.GlobalString("node-url")

	// userAddress := "627306090abaB3A6e1400e9345bC60c78a8BEf57"
	// privateKey := "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"
	// depositFilter := "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	// depositDescription := `[{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Deposit","type":"event"}]`

	// makeKey()
	// getCurrentChildBlock(nodeUrl, contractAddress)
	// getPendingNonce(nodeUrl, userAddress)

	// filter(nodeUrl, contractAddress, depositFilter, depositDescription)
	// plasmaFilter(nodeUrl, contractAddress)
	// plasmaWatchFilter(nodeUrl, contractAddress)
	getFromKeyStore()
	// for i := 0; i < 3; i++ {
	// 	deposit(nodeUrl, contractAddress, userAddress, privateKey)
	// }
	select {}
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

	plasma.WatchDeposit(&opts, ch)

	go func() {
		for true {
			select {
			// case err := <-s.Err():
			// 	fmt.Println("**** found error!")
			// 	fmt.Println(err)
			case event := <-ch:
				fmt.Println("**** found event!")
				fmt.Println(event)
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

func deposit(nodeUrl string, contractAddress string, userAddress string, privateKey string) {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	// Instantiate the contract and display its name
	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	privateKeyECDSA := getPrivateKeyECDSA(privateKey)
	auth := bind.NewKeyedTransactor(privateKeyECDSA)

	if err != nil {
		panic(err)
	}

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

const (
	KEYJSON_FILEDIR   = `/var/folders/vl/q9_v72n10cnd5tp0_014p4zr0000gn/T/go-ethereum-keystore061436836/UTC--2018-03-13T06-42-13.124847499Z--939597e3bc2bdb25e0a0bbe009ad1aeaae901400`
	SIGN_PASSPHRASE   = ""
	KEYSTORE_DIR      = `/var/folders/vl/q9_v72n10cnd5tp0_014p4zr0000gn/T/go-ethereum-keystore061436836`
	COINBASE_ADDR_HEX = `0x939597e3bc2bdb25e0a0bbe009ad1aeaae901400`
)

func getFromKeyStore() {
	// Init a keystore
	ks := keystore.NewKeyStore(
		KEYSTORE_DIR,
		keystore.LightScryptN,
		keystore.LightScryptP)
	fmt.Println()

	// Create account definitions
	fromAccDef := accounts.Account{
		Address: common.HexToAddress(COINBASE_ADDR_HEX),
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
	keyJson, readErr := ioutil.ReadFile(KEYJSON_FILEDIR)
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
	fmt.Printf("key extracted: addr=%s", keyWrapper.Address.String())

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
