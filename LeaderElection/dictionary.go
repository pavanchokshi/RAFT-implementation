package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Servers struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	IsLeader bool   `json:"IsLeader"`
}

var ServerList = make(map[string]Servers)
var Flag = false

func main() {
	fmt.Println("Inside dictionary on port :9999")

	http.HandleFunc("/addserver", AddServer)
	http.HandleFunc("/getallservers", GetAllServersList)
	http.HandleFunc("/setleader", SetLeader)
	http.HandleFunc("/getleader", GetLeader)
	http.HandleFunc("/flushall", FlushAll)
	http.ListenAndServe(":9999", nil)
}

func FlushAll(w http.ResponseWriter, r *http.Request) {

	if Flag == false {

		for k := range ServerList {
			delete(ServerList, k)
		}
		Flag = true
	}

}

func AddServer(w http.ResponseWriter, r *http.Request) {

	server := Servers{}

	//Decode JSON to struct
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&server)
	if err != nil {
		fmt.Println("Error in decoding json")
	}
	key := server.IP + server.Port
	ServerList[key] = server
	json.NewEncoder(w).Encode(ServerList[key])
}

func GetAllServersList(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Get All Servers")
	list := make([]Servers, len(ServerList))
	index := 0
	for k := range ServerList {
		list[index] = ServerList[k]
		index++
	}
	json.NewEncoder(w).Encode(list)
}

func SetLeader(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Inside Set Leader")

	server := Servers{}
	//Decode JSON to struct
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&server)
	if err != nil {
		fmt.Println("Error in decoding json")
	}

	key := server.IP + server.Port
	if ServerList[key].IP == "" {
		fmt.Println("Server doesnot exist")
	} else {
		ServerList[key] = server
		Flag = false
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ServerList[key])
}

func GetLeader(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Get Leader")
	leader := Servers{}
	for k, _ := range ServerList {
		if ServerList[k].IsLeader == true {
			leader = ServerList[k]
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(leader)
}
