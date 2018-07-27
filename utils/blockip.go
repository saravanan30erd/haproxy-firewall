package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	SecondsToExpireIP int64  = 21600
	defaultChain      string = "ufw-user-input"
)

func getItem(l []string, i int) (s string, res error) {
	if i < len(l) {
		s = l[i]
		res = errors.New("pass")
	}
	return
}

func addRule(ip string, timeStamp string) (res bool) {
	cmd := "iptables -A " + defaultChain + " -s " + ip + " -j DROP -m comment --comment " + timeStamp
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err == nil {
		res = true
	} else {
		errOutput := strings.Split(string(out), "\n")
		Errorlog.Println("IPTABLES FAILED(Add Rule): ", errOutput)
		res = false
	}
	return
}

func removeRule(ip string, timeStamp string) (res bool) {
	cmd := "iptables -D " + defaultChain + " -s " + ip + " -j DROP -m comment --comment " + timeStamp
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err == nil {
		res = true
	} else {
		errOutput := strings.Split(string(out), "\n")
		Errorlog.Println("IPTABLES FAILED(Remove Rule): ", errOutput)
		res = false
	}
	return
}

func BlockIP(ipBlockChan chan string) {
	for {
		ipFromChan := <-ipBlockChan
		time.Sleep(1000 * time.Millisecond)
		//slackMessage := fmt.Sprintf("Got DOS IP %s at %s", ipFromChan,
		//	time.Now().Format("Mon Jan _2 15:04:05 2006"))
		//SlackNotify(slackMessage)
		cmd := "iptables -nL " + defaultChain
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			lines = append(lines[:0], lines[1:]...)
			lines = append(lines[:0], lines[1:]...)
			for _, line := range lines {
				if strings.Contains(line, "/*") {
					lineList := strings.Fields(line)
					ip, res := getItem(lineList, 3)
					if res != nil {
						timeString, res := getItem(lineList, 6)
						if res != nil {
							timeInt, _ := strconv.Atoi(timeString)
							if ip == ipFromChan {
								//fmt.Println("IP matched", ipFromChan)
								res := removeRule(ipFromChan, timeString)
								if !res {
									slackMessage := fmt.Sprintf("IPtables Failed(Remove Rule) for %s",
										ipFromChan)
									SlackNotify(slackMessage)
								}
							} else if time.Now().Unix()-int64(timeInt) >= SecondsToExpireIP {
								//fmt.Println("Rule expired", ip)
								res := removeRule(ip, timeString)
								if !res {
									slackMessage := fmt.Sprintf("IPtables Failed(Remove Rule) for %s",
										ip)
									SlackNotify(slackMessage)
								}
							}
						}
					}
				}
			}
			timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
			//fmt.Println("Rule added", ipFromChan)
			res := addRule(ipFromChan, timeStamp)
			if !res {
				slackMessage := fmt.Sprintf("IPtables Failed(Add Rule) for %s",
					ipFromChan)
				SlackNotify(slackMessage)
			} else {
				slackMessage := fmt.Sprintf("%s blocked in iptables at %s",
					ipFromChan, time.Now().Format("Mon Jan _2 15:04:05 2006"))
				SlackNotify(slackMessage)
			}
		}
	}

}

func CleanExpireIPs() {
	cmd := "iptables -nL " + defaultChain
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		lines = append(lines[:0], lines[1:]...)
		lines = append(lines[:0], lines[1:]...)
		for _, line := range lines {
			if strings.Contains(line, "/*") {
				lineList := strings.Fields(line)
				timeString, res := getItem(lineList, 6)
				if res != nil {
					timeInt, _ := strconv.Atoi(timeString)
					if time.Now().Unix()-int64(timeInt) >= SecondsToExpireIP {
						ip, res := getItem(lineList, 3)
						if res != nil {
							r := removeRule(ip, timeString)
							if !r {
								slackMessage := fmt.Sprintf("IPtables Failed(Remove Rule) for %s",
									ip)
								SlackNotify(slackMessage)
							}
						}

					}
				}
			}
		}
	}
}
