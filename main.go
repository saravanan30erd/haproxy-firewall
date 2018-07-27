package main

import (
	"flag"
	"fmt"
	"github.com/bcicen/go-haproxy"
	"github.com/saravanan30erd/haproxy-dos-monitor/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	secondsToCheckHA int   = 10
	secondsTocacheIP int64 = 60
	requestLimit     int   = 5
	secToRunCleanFN  int   = 3600
)

var (
	logpath = flag.String("logpath", "/var/log/haproxy-dos-monitor.log", "Log Path")
)

type IPStore struct {
	sync.Mutex // this mutex protects the map(cache) below
	cache      map[string]int64
}

func New() *IPStore {
	return &IPStore{
		cache: make(map[string]int64),
	}
}
func (ips *IPStore) set(key string, value int64) {
	ips.Lock()
	ips.cache[key] = value
	ips.Unlock()
}
func (ips *IPStore) get(key string) (int64, bool) {
	ips.Lock()
	defer ips.Unlock()
	timeStamp, ok := ips.cache[key]
	return timeStamp, ok
}
func (ips *IPStore) del(key string) {
	ips.Lock()
	delete(ips.cache, key)
	ips.Unlock()
}

func getStat(quit chan bool, store *IPStore, ipBlockChan chan string) {
	client := &haproxy.HAProxyClient{
		Addr: "unix:///var/run/haproxy.sock",
	}
	//Check the haproxy table named "http" (haproxy frontend config name)
	result, err := client.RunCommand("show table http")
	if err != nil {
		utils.Errorlog.Println(err)
		quit <- false
	}
	output := strings.Split(result.String(), "\n")
	for _, line := range output {
		if strings.Contains(line, "key") {
			ip := strings.Split(strings.Fields(line)[1], "=")[1]
			//utils.Infolog.Printf("IP : %s\n", ip)
			countString := strings.Split(strings.Fields(line)[4], "=")[1]
			countInt, err := strconv.Atoi(countString)
			if err != nil {
				utils.Errorlog.Println(err)
			}
			//utils.Infolog.Printf("Count : %d\n", countInt)
			timeStamp, ok := store.get(ip)
			if ok {
				if time.Now().Unix()-timeStamp >= secondsTocacheIP {
					store.del(ip)
				}
			} else if countInt > requestLimit {
				store.set(ip, time.Now().Unix())
				ipBlockChan <- ip
				slackMessage := fmt.Sprintf("Got %s requests from %s at %s",
					countString, ip, time.Now().Format("Mon Jan _2 15:04:05 2006"))
				utils.SlackNotify(slackMessage)
			}
		}
	}
}

func main() {
	// Log initiate
	flag.Parse()
	utils.NewLog(*logpath)
	utils.Infolog.Println("haproxy-dos-monitor started")

	// channel used to terminate the app if haproxy socket fails
	quit := make(chan bool, 1)

	// channel to pass the IPs to blacklist
	ipBlockChan := make(chan string, 500)

	// Initiate the BlockIP func, it will keep listen ipblock channel and blacklist IPs
	go utils.BlockIP(ipBlockChan)

	store := New()
	ticker := time.NewTicker(time.Duration(secondsToCheckHA) * time.Second)
	tickerForIPExpire := time.NewTicker(time.Duration(secToRunCleanFN) * time.Second)
	for {
		select {
		case <-ticker.C:
			go getStat(quit, store, ipBlockChan)
		case <-tickerForIPExpire.C:
			go utils.CleanExpireIPs()
		case <-quit:
			message := "haproxy-dos-monitor DOWN"
			utils.Errorlog.Println(message)
			utils.SlackNotify(message)
			return
		}
	}
}
