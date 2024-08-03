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

    // accounts.txt 파일 열기
    accFile, err := os.Open("accounts.txt")
    if err != nil {
        log.Fatalf("Failed to open accounts file: %v", err)
    }
    defer accFile.Close()

    // 파일에서 계정 주소를 한 줄씩 읽기
    scanner := bufio.NewScanner(accFile)
    for scanner.Scan() {
        account := common.HexToAddress(scanner.Text())

        // 계정 잔액 조회
        balance, err := ec.BalanceAt(context.Background(), account, nil)
        if err != nil {
            log.Printf("Failed to retrieve balance for account %s: %v", account.Hex(), err)
            continue
        }

        fmt.Printf("Account %s balance: %d\n", account.Hex(), balance)
    }

    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading accounts file: %v", err)
    }

    fmt.Println("Finished retrieving balances.")
}