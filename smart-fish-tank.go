package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "io/ioutil"
    "os"
)

// function called when the ticker elapses
func AddCurrentTemperature(getCurrentTemperatureInterval time.Duration) {
	ticker := time.NewTicker(getCurrentTemperatureInterval)
	for _ = range ticker.C {
		fmt.Println("Tock")
	}
}

func main() {
    // default config values
    var dbUsername string = "root"
    var dbPassword string = ""
    var dbAddress string = "127.0.0.1:3306"
    // read from config file
    configJSON, err := os.Open("config.json")
    if err != nil {
        fmt.Println("Error reading \"config.json\", running with default config.")
    } else {
        fmt.Println("Loading config values from \"config.json\".")
        byteValue, _ := ioutil.ReadAll(configJSON)
        var configValues map[string]interface{}
        json.Unmarshal([]byte(byteValue), &configValues)
        dbUsername = configValues["database"].(map[string]interface{})["username"].(string)
        dbPassword = configValues["database"].(map[string]interface{})["password"].(string)
        dbAddress = configValues["database"].(map[string]interface{})["address"].(string)
    }
    defer configJSON.Close()

    // connect to mysql database
    db, err := sql.Open("mysql", dbUsername+":"+dbPassword+"@tcp("+dbAddress+")/")
    // if there is an error opening the connection, handle it
    if err != nil {
        panic(err.Error())
    } else {
        fmt.Println("Successfully connected to MySQL.")
    }
    // close db when main finishes executing
    defer db.Close()

    insert, err := db.Query("CREATE DATABASE IF NOT EXISTS `smart-fish-tank`;")
    // if there is an error inserting, handle it
    if err != nil {
        fmt.Println(err.Error())
        fmt.Println("Continuing without database \"smart-fish-tank\".")
    } else {
        fmt.Println("Using database \"smart-fish-tank\".")
    }
    // be careful deferring Queries if you are using transactions
    defer insert.Close()

    // set defined temperature request interval
    getCurrentTemperatureInterval := 10 * time.Minute
    go AddCurrentTemperature(getCurrentTemperatureInterval)

    // serve the static folder
    fileServer := http.FileServer(http.Dir("./static"))
    http.Handle("/", fileServer)
    // serve content on port 8080
    fmt.Printf("Running server on port 8080...\n")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }

}