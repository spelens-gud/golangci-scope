package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 100; i++ {
		fmt.Println(i)
		fmt.Println("hello world")
		time.Sleep(10 * time.Second)
	}

	return
}
