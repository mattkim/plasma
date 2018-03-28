package tester

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/plasma/contracts/gen/contracts"
)

func DepositFilter(plasma *contracts.Plasma) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
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
			length++
		}
		next = itr.Next()
	}

	fmt.Printf("%d Deposit logs found \n", length)
}

func SubmitBlockFilter(plasma *contracts.Plasma) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := plasma.FilterSubmitBlock(&opts)

	if err != nil {
		panic(err)
	}

	next := true
	length := 0

	for next {
		if itr.Event != nil {
			fmt.Println(itr.Event.Sender.Hex())
			fmt.Println(itr.Event.Root)
			length++
		}
		next = itr.Next()
	}

	fmt.Printf("%d Submit logs found \n", length)
}

func ExitStartedFilter(plasma *contracts.Plasma) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := plasma.FilterExitStarted(&opts)

	if err != nil {
		panic(err)
	}

	next := true
	length := 0

	for next {
		if itr.Event != nil {
			fmt.Println(itr.Event.Sender.Hex())
			fmt.Println(itr.Event.Blocknum)
			fmt.Println(itr.Event.Txindex)
			fmt.Println(itr.Event.Oindex)
			length++
		}
		next = itr.Next()
	}

	fmt.Printf("%d Exit started logs found \n", length)
}

func DepositWatchFilter(plasma *contracts.Plasma) {
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
