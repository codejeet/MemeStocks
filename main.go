package main

import (
    "fmt"
    "sort"
    "context"
    "regexp"
    "strings"
    "encoding/json"
    "os"
    "io/ioutil"
    "time"
    "net/http"
    twitterscraper "github.com/n0madic/twitter-scraper"
)




func BytesToString(data []byte) string {
	return string(data[:])
}

func GetTweets(user string) []string {
    var tweets[]string
    scraper := twitterscraper.New()
    
    //Store tweets into tweet araray
    
    for tweet := range scraper.GetTweets(context.Background(), user, 50) {
        if tweet.Error != nil {
            panic(tweet.Error)
        }
        tweets = append(tweets, tweet.Text)
    }
    return tweets
}

//Key Value Funcs

type kv struct {
    Key   string
    Value int
}

func rankMapStringInt(values map[string]int) []string {
    var ss []kv
    for k, v := range values {
        ss = append(ss, kv{k, v})
    }
    sort.Slice(ss, func(i, j int) bool {
        return ss[i].Value > ss[j].Value
    })
    ranked := make([]string, len(values))
    for i, kv := range ss {
        ranked[i] = kv.Key
    }
    return ranked
}


func WriteMapToJson(mapping map[string]int){
    b, _ := json.Marshal(mapping)
    path, _ := os.Getwd()
    filename := fmt.Sprintf("%s/%s.json", path, time.Now())
    f1, _ := os.Create(filename)
    f1.Write(b)
    f1.Sync()
    fmt.Println(b)
}

func LoadMapFromJson(filePath string) map[string]int{
    dat, _ := ioutil.ReadFile(filePath)

    var ret map[string]int
    if err := json.Unmarshal(dat, &ret); err != nil {
        panic(err)
    }

    return ret
    
}


func StocksFromUserArray(users []string) map[string]int{
    var tweets[]string
    for _, u := range users {
        for _, t := range GetTweets(u) {
            tweets = append(tweets, t) 
        }
    }
    
    // Mapping of stock names to amount of mentions
    mentions := make(map[string]int)

    for _, g := range tweets{
        	re := regexp.MustCompile(`\$[A-Za-z]+`)
            //Byte array array
            matches :=  re.FindAll([]byte(g), -1)
            //String array 
            for _, j := range matches{
                ret := string(j[:])
                mentions[ret]++
            }
    }
    return mentions
}


func GetCBOE(ticker string) []byte {
        resp, err := http.Get(fmt.Sprintf("https://cdn.cboe.com/api/global/delayed_quotes/options/%s.json", ticker))
        if err != nil {
                panic(err)
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                panic(err)
        }
        return body
}

func decodeInterface(b []byte)  map[string]interface{}{
    jsonMap := make(map[string](interface{}))
	err := json.Unmarshal([]byte(b), &jsonMap)
	if err != nil {
		fmt.Printf("ERROR: fail to unmarshall json, %s", err.Error())
	}

	// get response as a map with interface{}
	respMap := jsonMap["data"].(map[string]interface{})
    return respMap
}

func printCallPut(stockData map[string]interface{}) {
    var totalCalls int
    var totalPuts int
    for _, index := range stockData["options"].([]interface{}) {
        //list properties
        // for j, inner := range index.(map[string]interface{}) {
        //     fmt.Printf("%s -> %d \n", j, inner)
        // }

        isCall := strings.Contains(index.(map[string]interface{})["option"].(string), "C00")
        openInterest := int(index.(map[string]interface{})["open_interest"].(float64))

        if isCall {
           totalCalls += openInterest 
        } else {
           totalPuts += openInterest
        }
    }
    
    ratio := float64(totalPuts)/float64(totalCalls)

    fmt.Printf("Total Calls: %d \n", totalCalls)
    fmt.Printf("Total Puts: %d \n", totalPuts)
    fmt.Printf("Ratio (lower is better): %f \n", ratio)
}

func main() {
    users := []string{"list", "of", "twitter", "usersnames", "here"}

    mentions := StocksFromUserArray(users)
    WriteMapToJson(mentions)

    // Load previous results from a json file
    // mentions := LoadMapFromJson("2021-02-10 17:54:04.095373144 -0800 PST m=+9.448354701.json")

    for i, index := range rankMapStringInt(mentions) {
        fmt.Printf("%3d: %s -> %d \n", i, index, mentions[index])
    }

    // print the put call ratio of a particular stock
    // stockData := decodeInterface(GetCBOE("VGAC"))
    // printCallPut(stockData)
}


type tweetinfo struct{
    amount  int
    calls   int
    puts    int
    cpRatio float32
}
