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
    fmt.Println("\nStarting Smart Fish Tank...")
    // default config values
    var dbUsername string = "root"
    var dbPassword string = ""
    var dbAddress string = "127.0.0.1:3306"
    // read from config file
    configJSON, err := os.Open("config.json")
    if err != nil {
        fmt.Println("\033[31mError reading \"config.json\", running with default config.\033[0m")
    } else {
        fmt.Println("Loading config values from \"config.json\"...")
        byteValue, _ := ioutil.ReadAll(configJSON)
        var configValues map[string]interface{}
        json.Unmarshal([]byte(byteValue), &configValues)
        dbUsername = configValues["database"].(map[string]interface{})["username"].(string)
        dbPassword = configValues["database"].(map[string]interface{})["password"].(string)
        dbAddress = configValues["database"].(map[string]interface{})["address"].(string)
        fmt.Println("\033[32mSuccessfully loading config values!\033[0m")
    }
    defer configJSON.Close()

    // connect to mysql database
    db, err := sql.Open("mysql", dbUsername+":"+dbPassword+"@tcp("+dbAddress+")/")
    // if there is an error opening the connection, handle it
    if err != nil {
        fmt.Println("\033[31m"+err.Error()+"\033[0m")
    } else {
        fmt.Println("\033[32mSuccessfully connected to MySQL.\033[0m")
    }
    // close db when main finishes executing
    defer db.Close()

    insert, err := db.Query("CREATE DATABASE IF NOT EXISTS `smart-fish-tank`;")
    // if there is an error inserting, handle it
    if err != nil {
        fmt.Println("\033[31m"+err.Error()+"\033[0m")
        fmt.Println("\033[31mContinuing without database \"smart-fish-tank\".\033[0m")
    } else {
        fmt.Println("\033[32mUsing database \"smart-fish-tank\".\033[0m")
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
    fmt.Println("\033[36mRunning server on port 8080...\n\033[0m")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }

}