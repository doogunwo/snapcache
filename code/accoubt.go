package main

import (
    "bufio"
    "fmt"
    "log"
    "os"

)

func main() {
   
    // accounts.txt 파일 열기
    accFile, err := os.Open("accounts.txt")
    if err != nil {
        log.Fatalf("Failed to open accounts file: %v", err)
    }
    defer accFile.Close()
	cnt := 1
    // 파일에서 계정 주소를 한 줄씩 읽기
    scanner := bufio.NewScanner(accFile)
    for scanner.Scan() {
        cnt = cnt + 1

    }

    fmt.Println("total account : ", cnt)
	
    fmt.Println("Finished retrieving balances.")
}