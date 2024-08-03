package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/rpc"
)

const (
	gethIPCPath       = "/mnt/nvme1n1/ethereum/node1500/geth.ipc"
)

func main() {

	// Geth 클라이언트와 연결합니다.
	client, err := rpc.Dial(gethIPCPath)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	



	err = client.Call(nil, "admin_exit")
	if err != nil {
		fmt.Printf("Failed to stop Geth: %v\n", err)
	} else {
		fmt.Println("Geth stopped successfully")
	}

    if err != nil {
        log.Fatalf("Failed to send 'exit' command: %v", err)
    }
    fmt.Println("'exit' command sent successfully.")
	

	
}


