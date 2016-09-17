package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	//"httprouter"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Query struct {
	Query string `json:"query"`
}

type acknowledgeQuery struct {
	acknowledgeQuery string `json:"acknowledgequery"`
}

type CommitQuery struct {
	CommitQuery string `json:"commitquery"`
}

var i int
var q Query
var postString string

func main() {

	fmt.Println("Starting server on Port : 3000")
	router := httprouter.New()
	router.Handle("POST", "/initiate", InitiateFunc)
	router.Handle("POST", "/acknowledge/:msg", ReceiveAcknowledgement)
	log.Fatal(http.ListenAndServe(":3000", router))
	fmt.Println("calculating quoram")
}

func InitiateFunc(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	q := Query{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&q)
	//fmt.Println(q.Query)
	//fmt.Println("Hi2")
	postString = q.Query
	//err := decoder.Decode(&q)
	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}
	BroadCast(q)
}

func BroadCast(query Query) {
	for j := 1; j < 5; j++ {
		url := fmt.Sprintf("http://localhost:300%d/queries", j)
		client := http.Client{}
		b, _ := json.Marshal(query)
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, &b)
		req, err := http.NewRequest("POST", url, buf)
		if err != nil {
			fmt.Errorf("Error")
		}
		req.Header.Set("Content-Type", "application/json")
		res, _ := client.Do(req)
		res.Body.Close()
		if err != nil {
			fmt.Errorf("Error")
		}

	}

}
func ConnectToMongo() { //CONVERT INTO BOOLEAN

	//Name1 := "Nipun"
	//Email1 := "abcd@gmail.com"

	//fmt.Println("BEFORE URI")
	//fmt.Println("A")
	fmt.Println("HEY FOLLOWERS YOU HAVE TO STORE: ")
	fmt.Println(postString)
	//fmt.Println("B")
	uri := "mongodb://nipun:nipun@ds059804.mongolab.com:59804/raft"
	if uri == "" {
		fmt.Println("no connection string provided")

	}
	//fmt.Println("Before dial")
	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)

	}
	defer sess.Close()
	//fmt.Println("before setsafe")
	sess.SetSafe(&mgo.Safe{})
	//fmt.Println("before collection")

	collection := sess.DB("raft").C("raftCollection")

	//fmt.Println("Before")
	err = collection.Insert(&Query{postString})
	if err != nil {
		log.Fatal("Problem inserting data: ", err)
		return
	}
	//fmt.Println("After")

}

func ReceiveAcknowledgement(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//fmt.Println("in server1 with acknowledgement")

	/*ok := acknowledgeQuery{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&ok)
	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}*/

	fmt.Println(ps.ByName("msg"))
	if ps.ByName("msg") == "true" {
		//fmt.Println("inside IF")
		i++
		fmt.Println("acknowledgement received")
	}
	//fmt.Println(ok)

	if i > 3 {

		fmt.Println("THERE IS A QUORAM")
		Quoram(i)
		i = 0

	}
}

func Quoram(i int) {
	//fmt.Println("inside Quoram")

	ack := fmt.Sprintf("Value of Follwers is %d", i)
	fmt.Println(ack)

	ConnectToMongo()
	fmt.Println("Entry commited")
	fmt.Println("Now asking Followers to commit")

	Q := CommitQuery{}
	Q.CommitQuery = "Hey Follwers please append the entry"

	//fmt.Println("SEND postString")

	InsertQuery(Q)
}

func InsertQuery(query CommitQuery) { //DECIDE FLOW HERE
	for i := 1; i < 5; i++ {
		url := fmt.Sprintf("http://localhost:300%d/commit", i)
		client := http.Client{}
		b, _ := json.Marshal(query)
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, &b)
		req, err := http.NewRequest("POST", url, buf)
		if err != nil {
			fmt.Errorf("Error")
		}
		req.Header.Set("Content-Type", "application/json")
		res, _ := client.Do(req)
		res.Body.Close()
		if err != nil {
			fmt.Errorf("Error")
		}
	}

}
