package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
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

	// transactions.txt 파일 열기
	txFile, err := os.Open("transactions.txt")
	if err != nil {
		log.Fatalf("Failed to open transactions file: %v", err)
	}
	defer txFile.Close()

	
	cnt := 1
	// 파일에서 트랜잭션 해시를 한 줄씩 읽기
	scanner := bufio.NewScanner(txFile)
	for scanner.Scan() {
		txHash := common.HexToHash(scanner.Text())

		// 트랜잭션 조회
		tx, _, err := ec.TransactionByHash(context.Background(), txHash)
		if err != nil {
			log.Printf("Failed to retrieve transaction %s: %v", txHash.Hex(), err)
			continue
		}

		cnt = cnt + 1
		fmt.Println(cnt, " ",tx)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading transactions file: %v", err)
	}

	fmt.Println("Finished writing accounts to file.")
}