package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var CONCURRENT_JOB int = 10
var jobs chan int
var results chan int

const header_template = "HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n"

const numJobs = 5

func main() {
	//os.Args has arguemnts. "1:" and not "0:" as the first argument is the code executable.
	len := len(os.Args[1:])
	if len < 1 {
		log.Fatal("Port Number is Missing")
		return
	}
	//Checking if the specified port number is integer.
	i, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid Port Number")
		return
	}
	log.Println("Port Number: ", i)

	//formating port ":<port>"
	tcp_port := ":" + os.Args[1]

	//Listen on tcp_port
	if lst, l_err := net.Listen("tcp", tcp_port); l_err != nil {
		log.Fatal(l_err)
		return
	} else {
		//defer is executed only after the next command is executed.
		defer lst.Close()
		for {
			conn, a_err := lst.Accept()
			if a_err != nil {
				log.Fatal(a_err)
				// continue ?
			}
			handle_connection(conn)
			conn.Close()
		}
	}
}

func handle_connection(conn net.Conn) {
	log.Println("Client Socket: ", conn.RemoteAddr().String())

	//Buffer to store client data
	buf := make([]byte, 1024)

	//Read byte array data from connection
	len, r_err := conn.Read(buf)
	if r_err != nil {
		log.Fatal(r_err)
		//return ?
	}

	//To convert byte array data to httt.Request
	byte_nr := bytes.NewReader(buf[:len])    //1
	bufio_nr := bufio.NewReader(byte_nr)     //2
	req, r_err := http.ReadRequest(bufio_nr) //3
	if r_err != nil {
		log.Fatal(r_err)
		// return ?
	}
	//Check the http method
	log.Println("Request Method", req.Method)
	if req.Method != "GET" && req.Method != "POST" {
		dat, r_err := os.ReadFile("data/error/501.txt")
		if r_err != nil {
			log.Fatal(r_err)
			//return ?
		}
		conn.Write(dat)
	} else {
		//Removing '/' in url variable
		path := req.RequestURI[1:]
		log.Println("HTTP Request Path: ", path)
		if path != "html" && path != "txt" && path != "gif" && path != "jpeg" && path != "jpg" && path != "css" {
			dat, r_err := os.ReadFile("data/error/400.txt")
			if r_err != nil {
				log.Fatal(r_err)
				//return ?
			}
			conn.Write(dat)
		} else {
			uriPath := req.RequestURI[1:]
			file, r_err := os.Open("data/" + uriPath + "/" + "response." + uriPath)
			if r_err != nil {
				log.Fatal(r_err)
			}
			fileInfo, _ := file.Stat()

			var content_type string
			if path == "html" || path == "css" {
				content_type = "text/" + path
			} else if path == "jpg" || path == "jpeg" {
				content_type = "image/jpeg"
			} else if path == "gif" {
				content_type = "image/" + path
			} else {
				content_type = "text/plain"
			}

			rdat := fmt.Sprintf(header_template, content_type, fileInfo.Size())
			conn.Write([]byte(rdat))
			var err error
			_, err = io.Copy(conn, file)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}

func worker(id int, jobs <-chan int, results chan<- int) {

	for j := range jobs {
		log.Println("worker", id, "started  job", j)
		time.Sleep(time.Second)
		log.Println("worker", id, "finished job", j)
		results <- j * 2
	}

}

func newFunction() {
	jobs := make(chan int, 3)
	results := make(chan int, 3)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

//below code to get content type from []byte
//bs := make([]byte, fileInfo.Size())
//_, err = bufio.NewReader(file).Read(bs)
//if err != nil && err != io.EOF {
//	log.Fatal()
//	return
//}
//mimeType := http.DetectContentType(bs)
//log.Println("Detected Content Type", mimeType)
