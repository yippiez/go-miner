package main

/*
.TODO

*- Replace ReadString with ReadUntil
*- Unite all error reading to one function
*- Use flag to get arguments from command line

*/

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
	"bufio"
)


func checkErr(x error)(bool){
	if x == nil{
		return true
	}
	log.Print(x)
	log.Fatalf("Error quiting.")
	return false
}


func getServerInfo()(string){

	/* Gets the server info in the format of host:port */

	resp, err := http.Get("https://raw.githubusercontent.com/revoxhere/duino-coin/gh-pages/serverip.txt") // gets the response
	checkErr(err)
	
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
	
	reader := bufio.NewReader(conn) //reads from tcp connection

	for{ //while loop
		
		_,err := conn.Write([]byte("JOB,"+username)) // Asking for a job
		
		hash, err := reader.ReadString(',')
		hash = strings.TrimSuffix(hash, ",")
		
		if err != nil{
			log.Println("Error getting the job. Reconnecting to server in 15 seconds")
			time.Sleep(15 * time.Second)
			work(conn)
		}

		job, err := reader.ReadString(',')
		job = strings.TrimSuffix(job, ",")

		if err != nil{
			log.Println("Error getting the job. Reconnecting to server in 15 seconds")
			time.Sleep(15 * time.Second)
			work(conn)
		}

		diff := 3500 // fixed causes unnecesary lag


		for i := 0; i <= (diff * 100); i++ {
			hashes++ // add to hash counter
			h := sha1.New() //hashing object
			h.Write( []byte(hash + strconv.Itoa(i)) )
			newhash := hex.EncodeToString(h.Sum(nil))

			if (newhash) == job{ //if the result is the same with the job

				_,err = conn.Write( []byte(strconv.Itoa(i)) ) //sends the result of hash algorithm to the pool
				
				s, err := reader.ReadString('D')
				checkErr(err)		
				
				if strings.Contains(s, "GOOD"){
					accepted++
				}
			}
		}
	}
}


func workers(username string, addr string){
	conn, err := net.Dial("tcp", addr)
	buff := make([]byte, 3)
	_,_ = conn.Read(buff)
	log.Println("Server is on version:" + string(buff))

	checkErr(err)
	log.Println("Worker created")
	work(conn)
}

var username string = ""
var x int = 0

var hashes int = 0
var accepted int = 0
var balance float64 = 0
var balanceNew float64 = 0

func profit(){
	balanceNew := getBalance()
	log.Print("PROFIT: ",balanceNew-balance)
	balance = balanceNew
}

func calcHash(){
	totalKhashes := hashes / 1000
	hashes = 0
	log.Print("TOTAL",totalKhashes,"K/Hs")
	log.Print("Accepted Shares:",accepted)
}

func getBalance()(float64){
	
	conn, err := net.Dial("tcp", getServerInfo())
	checkErr(err)

	_,err = conn.Write([]byte("BALA," + username))
	checkErr(err)
	
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
		log.Println("How many goroutine you want?")
		fmt.Scan(&x)

	}else if len(argsWithoutProg) > 0{

		username = os.Args[1]
		x,_ = strconv.Atoi(os.Args[2])
	
	}

	balance = getBalance()

	for i:=0;i<x;i++{
		go workers(username, addr)
		time.Sleep(2*time.Second)
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

