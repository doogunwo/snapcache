package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	// 트랜잭션을 저장할 파일 생성
	file, err := os.Create("transactions.txt")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// 블록 10부터 163654까지 반복
	for i := uint64(0); i <= 163654; i++ {
		block, err := ec.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Fatalf("Failed to retrieve block %d: %v", i, err)
		}

		// 블록의 트랜잭션을 파일에 저장
		for _, tx := range block.Transactions() {
			_, err := file.WriteString(tx.Hash().Hex() + "\n")
			if err != nil {
				log.Fatalf("Failed to write transaction to file: %v", err)
			}
		}
		fmt.Printf("Processed block %d\n", i)
		
	}

	fmt.Println("Finished writing transactions to file.")
}