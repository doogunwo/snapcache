package main

import (
	"encoding/csv"
  "bufio"
  "fmt"
  "os"
  "strconv"

  "github.com/VictoriaMetrics/fastcache"
)

//목적 : account_frquncie.csv에

func main(){

  file , err := os.Open("./account_frequencies.csv")
  if err != nil {
    fmt.Println("Failed to open csv file : %v", err)
  }
  defer file.Close()

  reader := csv.NewReader(bufio.NewReader(file))
  
  records, _ := reader.ReadAll()

  cache := fastcache.New(100 * 1024 * 1024)

  for _ , record := range records[:1]{
    if len(record) < 2 {
      continue
    }
    address := record[0]
    balance, err := strconv.Atoi(record[1])
    if err != nil {
      fmt.Println("err : %s , %v", address ,  err)
      continue
    }

    if balance > 100 {
      fmt.Println([]byte(address) , []byte(record[1]))
      cache.Set([]byte(address), []byte(record[1]))
    }
  }

  err = cache.SaveToFile("account_cache.dat")
  if err != nil {
    fmt.Printf("Failed to save cache to fill : %v\n",err)
    return
  }
  
  fmt.Println("cache successfully wirtten to account cache.data")

}
