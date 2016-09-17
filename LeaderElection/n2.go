package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type MyDetails struct {
	IsLeader bool   `json:"IsLeader"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
}

type NodeList []MyDetails

var timeout chan bool
var ch chan int
var i int = 1
var vote string
var voteGiven bool = false
var voteCount int = 1 //DK
var candidate string
var total int //DK
var state string = "follower"
var timeup = time.Duration(5 * time.Second)
var retryCount = 3

func initialize() {
	myDetails := MyDetails{}
	myDetails.IsLeader = false
	myDetails.IP = "localhost" //enter the IP address
	myDetails.Port = "3001"    //enter the port number
	myDeatilsInJSON, _ := json.Marshal(myDetails)
	myDetailsBuff := new(bytes.Buffer)
	err := binary.Write(myDetailsBuff, binary.BigEndian, &myDeatilsInJSON)
	if err != nil {
		fmt.Errorf("Error in request API: %v", err)
	}
	url := fmt.Sprintf("http://localhost:9999/addserver")
	client := http.Client{}
	req, _ := http.NewRequest("POST", url, myDetailsBuff)
	res, _ := client.Do(req)
	res.Body.Close()
	fmt.Println("inside init")
}

func getNodes() NodeList {
	nodeList := NodeList{}
	resp, _ := http.Get("http://localhost:9999/getallservers")
	jsonDecoder := json.NewDecoder(resp.Body)
	err := jsonDecoder.Decode(&nodeList)
	if err != nil {
		fmt.Println("Error in decoding json")
	}
	total = len(nodeList)
	return nodeList
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func NotifyFollowers() {
	nodeList := getNodes()
	for _, val := range nodeList {
		url := fmt.Sprintf("http://%s:%s/iamleader/3001", val.IP, val.Port)
		client := http.Client{}
		req, _ := http.NewRequest("PUT", url, nil)
		res, _ := client.Do(req)
		res.Body.Close()
	}
}

func WatchDog() {
	for i > 0 {
		select {
		case <-ch:
			fmt.Println("Inside CH-Watchdog")
			if !voteGiven {
				state = "candidate"
				go AskForVote()
			}
		case <-timeout:
			if state != "leader" && state != "candidate" {
				GrantVote()
			}
		}
	}
}

func IncomingVoteRequest(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	candidate = p.ByName("candidate")
	fmt.Println("Vote request from", candidate)
	timeout <- true
}

func GrantVoteResponse(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	vote = p.ByName("vote")
	//ch <- 1
	fmt.Println("Vote received:", vote)
	if vote == "true" {
		voteCount++
	}
	fmt.Println(voteCount)
	if voteCount > total/2 && state != "leader" {
		fmt.Println("Leader....")
		state = "leader"
		myDetails := MyDetails{}
		myDetails.IsLeader = true
		myDetails.IP = "localhost" //enter the IP address
		myDetails.Port = "3001"    //enter the port number
		myDeatilsInJSON, _ := json.Marshal(myDetails)
		myDetailsBuff := new(bytes.Buffer)
		err := binary.Write(myDetailsBuff, binary.BigEndian, &myDeatilsInJSON)
		if err != nil {
			fmt.Errorf("Error in request API: %v", err)
		}
		url := fmt.Sprintf("http://localhost:9999/setleader")
		client := http.Client{}
		req, _ := http.NewRequest("POST", url, myDetailsBuff)
		res, _ := client.Do(req)
		res.Body.Close()
		NotifyFollowers() //DK
	}
}

func GrantVote() {
	var url string
	if candidate != "3001" {
		if !voteGiven {
			voteGiven = true
			fmt.Println("Vote given to:", candidate)
			url = fmt.Sprintf("http://localhost:%s/grantvote/%t", candidate, voteGiven)
		} else {
			fmt.Println("Vote denied to:", candidate)
			url = fmt.Sprintf("http://localhost:%s/grantvote/%t", candidate, !voteGiven)
		}

		client := http.Client{}
		req, _ := http.NewRequest("POST", url, nil)
		res, _ := client.Do(req)
		res.Body.Close()
	}
}

func AskForVote() {
	fmt.Println("Sending Vote request...")
	nodeList := getNodes()
	for _, value := range nodeList {
		if !voteGiven {
			url := fmt.Sprintf("http://localhost:%s/voteforme/3001", value.Port)
			fmt.Println("URL:", url)
			_, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Println("Ask for vote completed")
	/*	myrand := random(1, 10)
		time.Sleep(time.Duration(1000*myrand) * time.Millisecond)
		if state != "leader" || !voteGiven {
			ch <- 1
		}*/
}

func LeaderNotification(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if state != "leader" {
		currentLeader := p.ByName("leader")
		fmt.Println("We have new leader:", p.ByName("leader"))
		state = "follower"
		myrand := random(5, 15)
		// //int64(random(1, 10))
		time.Sleep(time.Duration(1000*myrand) * time.Millisecond)
		go IsLeaderAlive(currentLeader)
	}
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeup)
}

func IsLeaderAlive(currentLeader string) {

	if retryCount > 0 {
		fmt.Println("Trying......", retryCount)
		transport := http.Transport{
			Dial: dialTimeout,
		}

		client := http.Client{
			Transport: &transport,
		}

		url := fmt.Sprintf("http://localhost:%s/isalive", currentLeader)
		res, err := client.Get(url)

		if err != nil {

			time.Sleep(5000 * time.Millisecond)
			retryCount--
			IsLeaderAlive(currentLeader)
		}

		fmt.Println("Response Status:", res.StatusCode)
		if res.StatusCode == 200 {
			time.Sleep(5000 * time.Millisecond)
			IsLeaderAlive(currentLeader)
		}
	} else {
		resp, _ := http.Get("http://localhost:9999/flushall")
		resp.Body.Close()
		initiateElection()
	}
}

func initiateElection() {
	voteCount = 1
	voteGiven = false
	state = "follower"
	initialize()
	myrand := random(5, 15)
	// //int64(random(1, 10))
	time.Sleep(time.Duration(1000*myrand) * time.Millisecond)
	time.Sleep(1500 * time.Millisecond)
	AskForVote()
}

func NotifyIamAlive(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println("I am Alive")
	rw.WriteHeader(http.StatusOK)
}

func IceBreaker() {

	myrand := random(5, 15)
	// //int64(random(1, 10))
	time.Sleep(time.Duration(1000*myrand) * time.Millisecond)
	time.Sleep(1500 * time.Millisecond)
	fmt.Println("Inside Icebreaker:", ch)
	ch <- 1

}

func main() {
	initialize()
	timeout = make(chan bool, 1)
	ch = make(chan int, 1)
	//voteReqst = make(chan bool, 1)
	mux := httprouter.New()
	//go SendVoteChannel()
	//go RecvVoteChanel()
	go WatchDog()

	fmt.Println("Server ready to listen...")
	mux.GET("/voteforme/:candidate", IncomingVoteRequest)
	mux.POST("/grantvote/:vote", GrantVoteResponse)
	mux.PUT("/iamleader/:leader", LeaderNotification)
	mux.GET("/isalive", NotifyIamAlive)
	server := http.Server{
		Addr:    "0.0.0.0:3001",
		Handler: mux,
	}

	go IceBreaker()
	server.ListenAndServe()

}
