
package main

import (
	"log"
	"fmt"
	"os"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"time"
	"strings"
	"net"
	"bytes"
)

var username string = ""
var x int = 1 // goroutine count
var addr string = "51.15.127.80:2811"
var accepted int = 0 // accepted shares
var rejected int = 0 // decliend shares

func work(){

	conn, _ := net.Dial("tcp", addr)

	buffer := make([]byte, 3)
	_,err := conn.Read(buffer)
	log.Println("Server is on version:" + string(buffer))

	if(err != nil){
		log.Fatal("Servers might be down quitting")
	}

	for{

		// requesting a job
		_,err := conn.Write([]byte("JOB," + username))
		if(err != nil){
			log.Fatal("Error requesting job")
		}

		// making a buffer for the job
		buffer := make([]byte, 1024)
		_,err = conn.Read(buffer) // Getting the job
		if(err != nil){
			log.Fatal("Error getting the job")
		}

		job := strings.Split(string(buffer), ",") // parsing the job

		hash := job[0]
		goal := job[1]
		diff, _ := strconv.Atoi( strings.Replace(job[2],"\x00", "", -1) ) //Removes null bytes from job then converts it to an int

		// log.Println("Got a job DIF:" + strconv.Itoa(diff) + " HASH:" + hash + " GOAL:" + goal)
		
		for i := 0; i <= diff*100; i++{

			h := sha1.New()
			h.Write([]byte( hash + strconv.Itoa(i) )) // hash
			nh := hex.EncodeToString(h.Sum(nil))

			if nh == goal{
				_,err = conn.Write( []byte(strconv.Itoa(i)) ) //sends the result of hash algorithm to the pool
				
				if err != nil{
					log.Println("Error writing hash result")
					break
				}
				
				feedback_buffer := make([]byte, 6)
				_,err = conn.Read(feedback_buffer) //reads response

				feedback_buffer = bytes.Trim(feedback_buffer, "\x00")
				feedback := string(feedback_buffer)

				if feedback == "GOOD" || feedback == "BLOCK"{
					accepted++
				}else if feedback == "BAD"{
					rejected++
				}else if feedback == "INVU"{
					log.Fatal("Invalid username received in feedback")
				}

			}

		}

	}
}

func main(){

	argsWithoutProg := os.Args[1:]
	
	log.Println("GO miner started... \n")

	if len(argsWithoutProg) == 0 {

		log.Println("Enter Username:")
		fmt.Scan(&username)
		log.Println("How many goroutine you want?")
		fmt.Scan(&x)

	}else if len(argsWithoutProg) > 0{

		username = os.Args[1]
		x, _ = strconv.Atoi(os.Args[2])
	
	}

	string_count := strconv.Itoa(x);

	log.Println("USERNAME:" + username)
	log.Println("GOROUTINE COUNT:" + string_count)



	for i:=0; i<x; i++ {
		go work()
		time.Sleep(1 * time.Second)
	}

	for{

		log.Printf("Accepted shares :%d Rejected shares:%d\n", accepted, rejected)
		time.Sleep(10*time.Second)

	}
}
