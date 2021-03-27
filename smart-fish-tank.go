package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
)

func AddCurrentTemperature() {
	ticker := time.NewTicker(1 * time.Minute)
	for _ = range ticker.C {
		fmt.Println("Tock")
	}
}

func main() {
    go AddCurrentTemperature()

    fileServer := http.FileServer(http.Dir("./static"))
    http.Handle("/", fileServer)

    fmt.Printf("Starting server at port 8080\n")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }

}