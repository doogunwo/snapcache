package main

import (
        "context"
        "fmt"
        "log"

        "github.com/ethereum/go-ethereum/core/rawdb"
        "github.com/ethereum/go-ethereum/core/state"
        ec "github.com/ethereum/go-ethereum/ethclient"

)


func main(){

  ipcpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"
  dbpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth/chaindata"
  client, err := ec.Dial(ipcpath)
  if err != nil {
    log.Fatalf("Failed to connect geth ipc ",err)
  }

  block, err := client.BlockByNumber(context.Background(), nil)
  if err != nil {
    log.Fatalf("Failed to get the latest block : %v", err)
  }

  stateRoot := block.Root()
  
  fmt.Println(stateRoot)
  
  db, err := rawdb.NewLevelDBDatabase(dbpath,2048,128*4,"",false) 
  if err != nil {
    log.Fatalf("Failed to open the database :%v", err)
  }
  defer db.Close()
  
  stateDB := state.NewDatabase(db)
  stateTrie, err := stateDB.OpenTrie(stateRoot)
  if err != nil {
    log.Fatalf("Failed to open state trie : %v ", err)
  }
  
  it,err := stateTrie.NodeIterator(nil)
  if err != nil {
    log.Fatalf("stateTrie open error :%v",err)
  }
  for it.Next(true){
    fmt.Println("Account : %s", it)
  }

}
