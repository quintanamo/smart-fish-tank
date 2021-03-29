package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "io"
    "io/ioutil"
    "os"
    "math/rand"
    "strconv"
    "github.com/jacobsa/go-serial/serial"
)

// globals
const CLEAR_LCD = ("\xFE\x01                                \xFE\x01")

// function called when the ticker elapses
func AddCurrentTemperature(getCurrentTemperatureInterval time.Duration, db *sql.DB) {
	ticker := time.NewTicker(getCurrentTemperatureInterval)
    randomTemp := rand.Float64()*20+40
    tempAsStr := strconv.FormatFloat(randomTemp, 'E', -1, 64)
	for _ = range ticker.C {
		insert, err := db.Query("INSERT INTO temperature_data (temperature, recorded) VALUES ("+tempAsStr+", NOW())")
        // if there is an error inserting, handle it
        if err != nil {
            fmt.Println("\033[31m"+err.Error()+"\033[0m")
            fmt.Println("\033[31mCannot add temperature data to database \"smart-fish-tank\".\033[0m")
        } else {
            fmt.Println("\033[32mTemperature data added to database \"smart-fish-tank\".\033[0m")
        }
        // be careful deferring Queries if you are using transactions
        insert.Close()
	}
}

func DisplayLCDMessage(message string, port io.ReadWriteCloser) {
    port.Write([]byte(CLEAR_LCD))
    port.Write([]byte(message))
}

func main() {
    options := serial.OpenOptions{
        PortName: "/dev/serial0",
        BaudRate: 9600,
        DataBits: 8,
        StopBits: 1,
        MinimumReadSize: 4,
    }
    port, err := serial.Open(options)
    if err != nil {
        fmt.Println(err)
    }
    defer port.Close()
    DisplayLCDMessage("Starting smart  fish tank...", port)

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
        fmt.Println("\033[32mSuccessfully loaded config values!\033[0m")
    }
    defer configJSON.Close()

    // connect to mysql database
    db, err := sql.Open("mysql", dbUsername+":"+dbPassword+"@tcp("+dbAddress+")/smart_fish_tank")
    // if there is an error opening the connection, handle it
    if err != nil {
        fmt.Println("\033[31m"+err.Error()+"\033[0m")
        fmt.Println("\033[31mRunning without MySQL connection!  Data will not be saved!\033[0m")
    } else {
        fmt.Println("\033[32mSuccessfully connected to MySQL.\033[0m")
    }
    // close db when main finishes executing
    defer db.Close()

    // set defined temperature request interval
    getCurrentTemperatureInterval := 10 * time.Minute
    go AddCurrentTemperature(getCurrentTemperatureInterval, db)

    // serve the static folder
    fileServer := http.FileServer(http.Dir("./static"))
    http.Handle("/", fileServer)
    // serve content on port 8080
    fmt.Println("\033[36mRunning server on port 8080...\n\033[0m")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }

}