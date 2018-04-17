package node

import (
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

func StartDepositListener(level *db.Database, sink *TransactionSink, plasma *eth.PlasmaClient) {
	ch := make(chan eth.DepositEvent)
	sink.AcceptDepositEvents(ch)

	for {
		// TODO: change name to block number
		idx, err := level.DepositDao.LastDepositEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			panic(err)
		}

		log.Printf("Looking for deposit events at block number: %d\n", idx)

		events, lastIdx := plasma.DepositFilter(idx)

		if len(events) > 0 {
			count := uint64(0)

			for _, event := range events {
				ch <- eth.DepositEvent{
					Sender: event.Sender,
					Value:  event.Value,
				}

				count += 1

				// It's not synchronized right now...
				time.Sleep(time.Second * 3)
			}

			log.Printf("Found %d deposit events at from blocks %d to %d.\n", count, idx, lastIdx)

			// update deposits for the next round.
			level.DepositDao.SaveDepositEventIdx(lastIdx + 1)
		} else {
			log.Printf("No deposit events at block %d.\n", idx)
		}

		// Every 10 seconds look for deposits
		time.Sleep(time.Second * 10)
	}
}
