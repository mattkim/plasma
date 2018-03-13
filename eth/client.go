package eth

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/util"
)

const depositFilter = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
const depositDescription = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Deposit","type":"event"}]`

type DepositEvent struct {
	Sender common.Address
	Value  *big.Int
}

type Client struct {
	typedClient *ethclient.Client
	rpcClient   *rpc.Client
}

func NewClient(url string) (*Client, error) {
	fmt.Println(url)
	c, err := rpc.Dial(url)

	if err != nil {
		return nil, err
	}

	return &Client{typedClient: ethclient.NewClient(c), rpcClient: c}, nil
}

// TODO: test if i can sign stuff with this method.
func (c *Client) SignData(addr *common.Address, data []byte) ([]byte, error) {
	log.Printf("Attempting to sign data on behalf of %s", util.AddressToHex(addr))
	var res []byte
	err := c.rpcClient.Call(&res, "eth_sign", util.AddressToHex(addr), common.ToHex(data))
	log.Printf("Received signature on behalf of %s.", util.AddressToHex(addr))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) Subscribe(address common.Address) {
	filter := ethereum.FilterQuery{}
	filter.Addresses = make([]common.Address, 0)
	filter.Addresses = append(filter.Addresses, address)
	filter.FromBlock = big.NewInt(0)
	filter.Topics = [][]common.Hash{{common.HexToHash(depositFilter)}}

	// // Is this why it doesn't work?
	// ctx := context.TODO()

	// logs, _ := c.typedClient.FilterLogs(ctx, filter)
	// fmt.Println(len(logs)) // Ouptuts '143'

	ctx := context.Background()
	ch := make(chan types.Log)
	c.typedClient.SubscribeFilterLogs(ctx, filter, ch)

	for true {
		log := <-ch
		fmt.Println("Matching log encountered")
		fmt.Println(log)
	}
}

func (c *Client) SubscribeDeposits(address common.Address, resChan chan<- DepositEvent) error {
	query := ethereum.FilterQuery{
		FromBlock: nil,
		ToBlock:   nil,
		Topics:    [][]common.Hash{{common.HexToHash(depositFilter)}},
		Addresses: []common.Address{address},
	}

	// ch := make(chan types.Log)
	ch := make(chan types.Log, 2)

	s, err := c.typedClient.SubscribeFilterLogs(context.TODO(), query, ch)

	if err != nil {
		return err
	}

	log.Printf("Watching for deposits on address %s.", util.AddressToHex(&address))

	depositAbi, err := abi.JSON(strings.NewReader(depositDescription))

	if err != nil {
		return err
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
				parseDepositEvent(&depositAbi, resChan, &event)
			}
		}
	}()

	return nil
}

func parseDepositEvent(depositAbi *abi.ABI, resChan chan<- DepositEvent, raw *types.Log) {
	event := DepositEvent{}
	err := depositAbi.Unpack(&event, "Deposit", raw.Data)

	if err != nil {
		log.Print("Failed to unpack deposit: ", err)
		return
	}

	log.Printf("Received %s wei deposit from %s.", event.Value.String(), util.AddressToHex(&event.Sender))
	resChan <- event
}
