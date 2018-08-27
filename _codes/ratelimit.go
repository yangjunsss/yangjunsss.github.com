package main

import (
  "fmt"
  "time"
)

func main(){
    rate := time.Second/10
    throttle := time.Tick(rate)
    requests := []int{0,1,2,3}
    for r := range requests{
      <- throttle
      go func(n int){
        fmt.Println(n)
      }(r)
    }

    burstLimit := 1
    tick := time.NewTicker(rate)
    defer tick.Stop()
    throttle1 := make(chan time.Time, burstLimit)
    go func(){
      for t := range tick.C {
        select {
        case throttle1 <- t:
          fmt.Println("tick")
        default:
          fmt.Println("drop")
        }
      }
    }()
    for r:= range requests {
      time.Sleep(1*time.Second)
      <- throttle1
      go func(n int){
        fmt.Println(n)
      }(r)
    }
}
