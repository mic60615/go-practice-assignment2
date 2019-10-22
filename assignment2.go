package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"
)

type timeAPIResponse struct {
	Abbrev      string `json: "abbreviation"`
	Datetime    string `json: "datetime"`
	DayOfWeek   int    `json: "day_of_week"`
	DayOfYear   int    `json: "day_of_year"`
	DST         bool   `json: "dst"`
	DSTFrom     string `json: "dst_from"`
	DSTOffset   int    `json: "dst_offset`
	DSTUntil    string `json: dst_until`
	RawOffset   int    `json: "raw_offset"`
	TimeZone    string `json: "timezone"`
	Unixtime    int64  `json: "unixtime`
	UTCDatetime string `json: "utc_datetime"`
	UTCOffset   string `json: "utc_offset"`
	WeekNumber  int    `json: "week_number"`
}

type TimeData struct {
	StartTime       int64
	LastFetchedTime int64
	NumRequests     uint32
}

type ChannelData struct {
	IP              string
	LastFetchedTime int64
	RequestTime     int64
}

// func FloatToString(input float64) string {
// 	return strconv.FormatFloat(input, 'E', -1, 64)
// }

var timeData TimeData
var channelData ChannelData
var logdatachan chan ChannelData

func main() {

	logdatachan = make(chan ChannelData, 1)
	timeData.StartTime = time.Now().UnixNano()
	duration, _ := time.ParseDuration(fmt.Sprintf("%f", math.E) + "s")

	getTimeTicker := time.NewTicker(duration)
	quit := make(chan struct{})

	//background goroutine which fetches data every E seconds
	go func() {

		for {

			select {
			case <-getTimeTicker.C:

				timeData.NumRequests++
				response, err := http.Get("http://worldtimeapi.org/api/ip")
				if err != nil {

				}
				body, _ := ioutil.ReadAll(response.Body)
				var resp timeAPIResponse
				json.Unmarshal(body, &resp)
				timeData.LastFetchedTime = resp.Unixtime

			case <-quit:
				return

			}
		}

	}()

	//goroutine which logs data to a file every time someone fires a GET request
	go func() {
		for {
			select {
			case channelData := <-logdatachan:
				//fmt.Println(3)
				//fmt.Println(channelData)
				//channelData = ChannelData{}
				file, _ := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

				fmt.Fprintf(file, "%s-%d-%d\n", channelData.IP, channelData.RequestTime, channelData.LastFetchedTime)

			case <-quit:
				return
			}
		}
	}()

	http.HandleFunc("/", timeFunc)

	http.ListenAndServe(":12345", nil)

	time.Sleep(time.Hour)

}

//runs on every GET request
//prints out background data (i.e. the data updated every E seconds)
//updates data to be logged to file and sent to channel
func timeFunc(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		timeStr, err := json.Marshal(timeData)
		if err == nil {
			fmt.Fprintf(w, "%s", timeStr)

		} else {
			fmt.Fprintf(w, "error occurred while marshalling timedata")
		}

		//////////////////////////

		channelData.LastFetchedTime = timeData.LastFetchedTime
		channelData.RequestTime = time.Now().UnixNano()
		channelData.IP = req.RemoteAddr
		logdatachan <- channelData

	}
}
