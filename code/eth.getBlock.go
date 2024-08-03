package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	// geth.ipc에 연결
	client, err := rpc.Dial("../../network1/peer1/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// ethclient 인스턴스 생성
	ec := ethclient.NewClient(client)


	// 블록 0부터 163654까지 반복
	for i := uint64(0); i <= 163654; i++ {
		block, err := ec.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Fatalf("Failed to retrieve block %d: %v", i, err)
		}

	
		fmt.Println("Processed block %d\n", i)
		fmt.Println("Processed block %d\n", block)

		
	}

	fmt.Println("Finished writing transactions to file.")
}