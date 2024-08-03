package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/VictoriaMetrics/fastcache"
)


func main(){
  cache , err := fastcache.LoadFromFile("./account_cache.dat/")
  if err != nil {
    fmt.Printf("Failed to load cache from file : %v\n",err)
    return
  }

  file ,err := os.Open("./account_frequencies.csv")
  if err != nil {
    fmt.Printf("Failed to open csv file : %v\n",err)
    return 
  }
  defer file.Close()

  reader := csv.NewReader(file)
  records, err := reader.ReadAll()
  if err != nil {
    fmt.Printf("Failed to open csv file :%v\n",err)
    return
  }
  cnt := 1
  for _, keys := range records[1:] {
    if len(keys) < 2{
      continue
    }

    if cnt > 800 {
      break
    }
    key := keys[0]
    balance := cache.Get(nil,[]byte(key))
    if len(balance) >0 {
      fmt.Println("addr : %s , balance : %s \n",key , balance)
    }else {
      fmt.Println("addr : %s , not found in cache \n",key )
    }
    cnt = cnt +1
  }


}
