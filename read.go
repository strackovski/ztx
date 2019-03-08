// Author: Vladimir Strackovski <vladimir.strackovski@dlabs.si>
// 08/03/2019
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "runtime"
    "sync"
    "time"
)

// Constants
// If another currency is added, add address data, API endpoint URLs and slice sizes here.
// URL is a paremetrised string to be used with 'printf' to replace the string placeholder (%s).
const (
    ethDataFile    = "data/ethereum.json"
    ethEndpointUrl = "http://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=asc&apikey=YourApiKeyToken"
    chunkSize      = 50
)

// eth is a slice of all Ethereum addresses to check
var eth []string
// ethChunks is a slice of chunked Ethereum address slices
var ethChunks [][]string

var failure int

// init runs first to initialise raw data
func init() {
    log.Printf("Reading Ethereum addresses from data file '%s'...\n", ethDataFile)
    ethRaw, err := ioutil.ReadFile(ethDataFile)
    if err != nil {
        log.Printf("Error reading file '%s': '%s\n'", ethDataFile, err)
    }

    log.Println("Unmarshalling raw data to slices...")
    err = json.Unmarshal(ethRaw, &eth)
    if err != nil {
        log.Printf("Error unmarshalling data from file '%s': '%s'\n", ethDataFile, err)
    }

    log.Printf("Unmarshalled data from file '%s'\n", ethDataFile)
    log.Println("Chunking...")
    for i := 0; i < len(eth); i += chunkSize {
        end := i + chunkSize

        if end > len(eth) {
            end = len(eth)
        }

        ethChunks = append(ethChunks, eth[i:end])
    }
    log.Println("Init done.")
}

// call is the function that executes the request to the given URL, passing the data
// value in the url parameter.
func call(url, data string) string {
    resp, err := http.Get(fmt.Sprintf(url, data))

    if err != nil {
        failure++
        log.Printf("An error occured while calling URL '%s': %s\n", url, err)
    }

    if resp.StatusCode != 200 {
        failure++
    }

    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()

    log.Println(string(body))
    return string(body)
}

// main function
func main() {
    // Set max proc (arbitrary atm), and initial address count.
    runtime.GOMAXPROCS(10)
    ethAddressCount := 0
    failure = 0

    log.Println("Starting timer and wait group...")
    wg := sync.WaitGroup{} // Make sure to wait for all routines to finish
    start := time.Now()    // Start the timer

    // Iterate over aggregate address slice
    for _, ethChunk := range ethChunks {
        // Iterate over the slice of chunks
        for _, ethAddr := range ethChunk {
            ethAddressCount++
            wg.Add(1) // Add wait group
            go func() {
                defer wg.Done()               // Wait until all groups finished
                call(ethEndpointUrl, ethAddr) // Make the call to the API
            }()
        }
    }

    wg.Wait()                    // Ensure all routines finish before returning
    elapsed := time.Since(start) // Measure elapsed time since starting the timer

    log.Println("================")
    log.Printf("Success / failure (rate): %d / %d (%d)", 200 - failure, failure, 200 / failure)
    log.Printf("Reading %d ETH wallets took %s", ethAddressCount, elapsed)
}
