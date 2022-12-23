package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// set the following variables manually
	hitCount := 300
	hitUrl := "http://localhost:3000/api"

	for i := 1; i <= hitCount; i++ {
		fmt.Printf("Hit Count: %d\n", i)
		res, err := http.Get(hitUrl)
		if err != nil {
			log.Fatalln(err)
			return
		}
		if res.StatusCode != 200 {
			log.Printf("The status is wrong. Status: %d\n", res.StatusCode)
			break
		}
	}

	fmt.Printf("Successfully run the hitter with %d times \n", hitCount)
}
