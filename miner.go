package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"crypto/sha1"
	"strings"
	"strconv"
	"net"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)


func checkErr(x error)(bool){
	if x == nil{
		return true
	}

	log.Println(x)
	return false
}


func getServerInfo()(string){

	/*
		Gets the server info in the format of host:port
	*/

	resp, err := http.Get("https://raw.githubusercontent.com/revoxhere/duino-coin/gh-pages/serverip.txt") // gets the response

	if !checkErr(err){
		log.Println("Error getting server info trying again in 15 seconds")
		time.Sleep(15 * time.Second)
		return getServerInfo()
	}
	
	log.Println("Got the server info!")
	
	defer resp.Body.Close() // waits for the functions end to execute

	body, err := ioutil.ReadAll(resp.Body) // reads all data
	
	if !checkErr(err){
		log.Println("Error parsing the get body trying again")
		return getServerInfo()
	}

	content := strings.Split(string(body),"\n") // converts string into array
	host := content[0:2][0] // parses host value
	port := content[0:2][1]	// parses port value

	if len(host)>0 && len(port)>0{
		return (host + ":" + port) 
	}

	return getServerInfo()
}

func work(conn net.Conn){
	for{ //while loop
		_,err := conn.Write([]byte("JOB")) // Asking for a job

		buffer := make([]byte, 1024)
		_,err = conn.Read(buffer) // Getting the job
		
		if !checkErr(err){
			log.Println("Error getting the job. Reconnecting to server in 15 seconds")
			time.Sleep(15 * time.Second)
			work(connect(username, password, getServerInfo()))
		}

		job := strings.Split(string(buffer), ",") // parsing the job
		buffer = make([]byte, 4) // buffer for receiving
		diff, _ := strconv.Atoi( strings.Replace(job[2],"\x00", "", -1) ) //Removes null bytes from job then converts it to an int
		


		for i := 0; i <= (diff * 100); i++ {
			hashes++
			h := sha1.New() //hashing object
			h.Write( []byte(job[0] + strconv.Itoa(i)) )
			nh := hex.EncodeToString(h.Sum(nil))

			if (nh) == job[1]{ //if the result is even with the job

				_,err = conn.Write( []byte(strconv.Itoa(i)) ) //sends the result of hash algorithm to the pool

				_,err = conn.Read(buffer) //reads response
				
				if !checkErr(err){
					break
				}
				
				if string(buffer) == "GOOD"{
					accepted++
					//log.Printf("Accepted share %d Difficulty %d\n",i,diff)
				}
				/*
				else if string(buffer) == "BAD "{
					log.Printf("Rejected share %d Difficulty %d\n",i,diff)
				}
				*/
			}
		}
	}
}

func connect(username string, password string, addr string) net.Conn{

	
	conn, err := net.Dial("tcp", addr)
	
	if !checkErr(err){
		log.Println("Error creating connection trying again in 15 seconds")
		time.Sleep(15 * time.Second)
		return connect(username, password,addr)
	}

	// Get the current server version
	buffer := make([]byte, 3)
	_,err = conn.Read(buffer)
	log.Println("Server is on version:" + string(buffer))

	if(!checkErr(err)){
		log.Println("Servers might be down retry in 15 seconds.")
		time.Sleep(15 * time.Second)
		return connect(username, password,addr)	
	}

	// Login to server
	buffer = make([]byte, 3)
	loginString := "LOGI," + strings.Replace(username,"\n", "", -1) + "," + strings.Replace(password,"\n", "", -1)
	conn.Write([]byte(loginString))

	// Feedback
	_,err = conn.Read(buffer)
	log.Println("Login feedback: " + string(buffer))
	
	if(!checkErr(err)){
		log.Println("Cannot receive login feedback retry in 15 seconds.")
		time.Sleep(15 * time.Second)
		return connect(username, password,addr)	
	}

	if string(buffer) == "NO,"{
		log.Println("Wrong username or password.")
		return connect(username, password,addr)	
	}

	return conn
}


func workers(username string, password string, addr string){
	conn := connect(username,password,addr)
	defer conn.Close()
	work(conn)
}

var username string = ""
var password string = ""
var x int = 0

var hashes int = 0
var accepted int = 0
var balance float64 = 0
var balanceNew float64 = 0

func profit(){
	balanceNew := getBalance()
	log.Print("PROFIT: ",balanceNew-balance)
}

func calcHash(){
	totalKhashes := hashes / 1000
	hashes = 0
	log.Print("TOTAL",totalKhashes,"K/Hs")
	log.Print("Accepted Shares:",accepted)
}

func getBalance()(float64){
	
	conn := connect(username, password, getServerInfo())

	_,err := conn.Write([]byte("BALA"))
	
	if(!checkErr(err)){
		return -1
	}

	buffer := make([]byte, 100)
	conn.Read(buffer)
	ball, _ := strconv.ParseFloat( strings.Replace(string(buffer),"\x00", "", -1), 32)

	return ball
}

func main() {

	argsWithoutProg := os.Args[1:]
	addr := getServerInfo()
	

	if len(argsWithoutProg) == 0 {

		log.Println("Enter Username:")
		fmt.Scan(&username)
		log.Println("Enter password:")
		fmt.Scan(&password)
		log.Println("How many goroutine you want?")
		fmt.Scan(&x)

	}else if len(argsWithoutProg) > 0{

		username = os.Args[1]
		password = os.Args[2]
		x,_ = strconv.Atoi(os.Args[3])
	
	}

	balance = getBalance()

	for i:=0;i<x;i++{
		go workers(username, password,addr)
	}
	
	go func(){
		for{
			time.Sleep(60 * time.Second)
			profit()
		}
	}()

	for{
		time.Sleep(1 * time.Second)
		calcHash()
	}
}

