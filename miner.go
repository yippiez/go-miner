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
	"sync"
	"fmt"
)

func getServerInfo()(string){
	/*
		Gets the server info in the format of host:port
	*/

	resp, err := http.Get("https://raw.githubusercontent.com/revoxhere/duino-coin/gh-pages/serverip.txt") // gets the response

	if err != nil{
		log.Fatalln(err) // prints error and then exits
	}
	
	log.Println("Got the server info!")
	
	defer resp.Body.Close() // waits for the functions end to execute

	body, err := ioutil.ReadAll(resp.Body) // reads all data
	if err != nil{log.Fatalln(err)} // checks error

	content := strings.Split(string(body),"\n") // converts string into array
	host := content[0:2][0] // parses host value
	port := content[0:2][1]	// parses port value

	return (host + ":" + port) 
}

func work(conn net.Conn){
for{ //while loop
	_,err := conn.Write([]byte("JOB")) // Asking for a job
	
	if err != nil{
		log.Fatalln(err)
	}

	buffer := make([]byte, 1024)
	_,err = conn.Read(buffer) // Getting the job
	if err!=nil{
		log.Println("Error getting the job. Reconnecting to server")
		work(connect(username, password))
	
	}

	job := strings.Split(string(buffer), ",") // parsing the job
	buffer = make([]byte, 1024) // buffer for receiving
	diff, _ := strconv.Atoi( strings.Replace(job[2],"\x00", "", -1) ) //Removes null bytes from job then converts it to an int
	


	for i := 0; i <= (diff * 100); i++ {
		h := sha1.New() //hashing object
		h.Write( []byte(job[0] + strconv.Itoa(i)) )
		nh := hex.EncodeToString(h.Sum(nil))

		if (nh) == job[1]{ //if the result is even with the job

			_,err = conn.Write( []byte(strconv.Itoa(i)) ) //sends the result of hash algorithm to the pool
			
			if err != nil{
				log.Fatalln(err)
			}

			_,err = conn.Read(buffer) //reads response
			
			if err != nil{
				log.Fatalln(err)
			}
			

			if strings.Replace(string(buffer),"\x00", "", -1) == "GOOD"{
				log.Printf("Accepted share %d Difficulty %d\n",i,diff)
			}else if strings.Replace(string(buffer),"\x00", "", -1) == "BAD"{
				log.Printf("Rejected share %d Difficulty %d\n",i,diff)
			}
		}
	}
}
}

func connect(username string, password string) net.Conn{

	addr := getServerInfo()
	conn, err := net.Dial("tcp", addr)
	if(err!=nil){log.Fatalln(err)}

	// Get the current server version
	buffer := make([]byte, 3)
	_,err = conn.Read(buffer)
	log.Println("Server is on version:" + string(buffer))
	if(err!=nil){
		log.Println("Servers might be down.")
		log.Fatalln(err)
	}

	// Login to server
	buffer = make([]byte, 3)
	loginString := "LOGI," + strings.Replace(username,"\n", "", -1) + "," + strings.Replace(password,"\n", "", -1)
	conn.Write([]byte(loginString))

	// Feedback
	_,err = conn.Read(buffer)
	log.Println("Login feedback: " + string(buffer))
	if err != nil {log.Fatalln(err)}

	if string(buffer) == "NO,"{
		log.Fatalln("Wrong username or password.")
	}

	return conn
}


func workers(username string, password string){
	conn := connect(username,password)
	defer conn.Close()
	work(conn)
}

var username string = ""
var password string = ""

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	
	log.Println("Enter Username:")
	fmt.Scan(&username)
	log.Println("Enter password:")
	fmt.Scan(&password)

	x := 0
	log.Println("How many goroutine you want?")
	fmt.Scan(&x)

	for i:=0;i<x;i++{
		go workers(username, password)
	}
	
	wg.Wait()
}

