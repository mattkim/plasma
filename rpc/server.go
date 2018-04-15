package rpc

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func Start(
	port int,
	txService *TransactionService,
	blockService *BlockService,
) {
	log.Printf("Starting RPC server on port %d.", port)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(txService, "Transaction")
	s.RegisterService(blockService, "Block")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", port), r)
}
