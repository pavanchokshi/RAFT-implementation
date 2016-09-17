package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	//"bytes"
	//"encoding/binary"
	//"httprouter"
	"log"
	"net/http"

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


var postString string
var q Query

func main() {

	fmt.Println("Starting server on Port : 3003")
	router := httprouter.New()
	router.Handle("POST", "/queries", TestFunc)
	router.Handle("POST", "/commit", ReceiveCommit)
	log.Fatal(http.ListenAndServe(":3003", router))
}

func TestFunc(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("in server4")
	//q := Query{}
	ok := acknowledgeQuery{}
	ok.acknowledgeQuery = "OK from server 4"
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&q)

	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}
	//fmt.Println(q)
	postString = q.Query
	acknowledgeLeader(ok)
	//fmt.Println(ok)

}

func ConnectToMongo() {   //CONVERT INTO BOOLEAN

	//Name1 := "Nipun"
	//Email1 := "abcd@gmail.com"

	//fmt.Println("BEFORE URI")
	//fmt.Println("A")
	fmt.Println(postString)
	fmt.Println("B")
	uri := "mongodb://nipun:nipun@ds059804.mongolab.com:59804/raft"
   if uri == "" {
                fmt.Println("no connection string provided")
               
        }
 		fmt.Println("Before dial")
        sess, err := mgo.Dial(uri)
        if err != nil {
                fmt.Printf("Can't connect to mongo, go error %v\n", err)
    
        }
        defer sess.Close()
        fmt.Println("before setsafe")
        sess.SetSafe(&mgo.Safe{})
        fmt.Println("before collection")

      	collection := sess.DB("raft").C("raftCollection4")
      	fmt.Println("ENTRY COMMITED")

        fmt.Println("Before")
        err = collection.Insert(&Query{postString})
        if err != nil {
                log.Fatal("Problem inserting data: ", err)
                return
        }
        fmt.Println("After")



	
}



func ReceiveCommit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("in server4 while commiting")
	Q := CommitQuery{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&Q)
	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}
	fmt.Println(Q)
	ConnectToMongo()

}

func acknowledgeLeader(OK acknowledgeQuery) {

	url := fmt.Sprintf("http://localhost:3000/acknowledge/true")
	fmt.Println(OK)
	client := http.Client{}
	//b, _ := json.Marshal(OK)
	//buf := new(bytes.Buffer)
	//err := binary.Write(buf, binary.BigEndian, &b)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Errorf("Error")
	}
	//req.Header.Set("Content-Type", "application/json")
	res, _ := client.Do(req)
	res.Body.Close()
	if err != nil {
		fmt.Errorf("Error")
	}

}

