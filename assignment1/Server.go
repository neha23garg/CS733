package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const PORT = ":8080"

//struct of all needed file attributes
type FileAttributes struct {
	version, NumBytes, ExpTime int64
	data                       []byte    //will conatains actual content
	FutureExpDate              time.Time //timestamp till the file is valid and not expired

}

var (
	FileMap = make(map[string]FileAttributes)
)

//declaration of lock variable
var myLock = &sync.RWMutex{}

func serverMain() {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", PORT)

	//report error if connection is not established
	if err != nil {
		log.Printf("Error in listener: %v", err.Error())
	}

	// run loop forever (or until ctrl-c)  for handling requests continuously
	for {
		// accept connection on port
		conn, _ := ln.Accept()

		//handle multiple clients

		go processClient(conn)

	}
}

//Function for processing all client requests
func processClient(conn net.Conn) {
	for {
		dataBuffer := make([]byte, 1024)
		// will listen for message to process ending in newline (\n)
		size, err := conn.Read(dataBuffer[0:])
		if err != nil {
			log.Printf("Error in reading from client")
		}
		dataBuffer = dataBuffer[:size]
		//data1 will contain data staring from beginning to first new line
		data1 := strings.Split(string(dataBuffer), "\r\n")[0]
		firstIndexOfnewLine := bytes.Index(dataBuffer, []byte("\r\n"))
		//splitting on space
		command := strings.Split(data1, " ")
		argLen := len(command)
		if len(command) == 0 {
			errMsg := "ERR_CMD_ERR\r\n"
			conn.Write([]byte(errMsg))
		}
		errMsg := ""
		switch command[0] {

		case "read":
			if argLen != 2 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			//if file name is more than 250 bytes
			if len(command[1]) > 250 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			//calling read function
			ver, noBytes, exp, cont, errStr := ReadProcess(command)
			if errStr != "" {
				errMsg := "ERR_FILE_NOT_FOUND\r\n"
				conn.Write([]byte(errMsg))
			}
			//passing the result to client
			conn.Write([]byte("CONTENTS " + strconv.FormatInt(ver, 10) + " " + strconv.FormatInt(noBytes, 10) + " " + strconv.FormatInt(exp, 10) + "\r\n" + string(cont) + "\r\n"))

		case "write":
			if argLen < 3 || argLen > 4 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			//file name can't contains spaces and if file name is more than 250 bytes
			if strings.Contains(string(command[1]), " ") || len(command[1]) > 250 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			writeContent := bytes.Trim(dataBuffer[firstIndexOfnewLine:], "\r\n")
			numBytes, err := strconv.Atoi(command[2])
			if err != nil {
				errMsg = "ERR_INTERNAL1"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			if numBytes != len(writeContent) {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}

			ver, err := WriteProcess(command, writeContent)

			//file has already expired
			if ver == 1 {
				errMsg := "ERR_FILE_NOT_FOUND\r\n"
				conn.Write([]byte(errMsg))
			}
			//some other error
			if err != nil && ver != 1 {
				errMsg = "ERR_INTERNAL2"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			//passing the result to client
			conn.Write([]byte("OK" + " " + strconv.FormatInt(ver, 10) + "\r\n"))

		case "cas":
			if argLen < 4 || argLen > 5 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			//file name can't contains spaces and if file name is more than 250 bytes
			if strings.Contains(string(command[1]), " ") || len(command[1]) > 250 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			writeContent := bytes.Trim(dataBuffer[firstIndexOfnewLine:], "\r\n")
			numBytes, err := strconv.Atoi(command[3])
			if err != nil {
				errMsg = "ERR_INTERNAL"
				conn.Write([]byte(errMsg + "\r\n"))
			}
			if numBytes != len(writeContent) {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}

			//calling the cas(compare and swap) function function
			ver, rec_err := CasProcess(command, writeContent)

			if rec_err != "" && ver == 0 {
				conn.Write([]byte(rec_err + "\r\n"))
			}
			if rec_err != "" && ver != 0 {
				conn.Write([]byte(rec_err + " " + strconv.FormatInt(ver, 10) + "\r\n"))
			}
			//passing the result to client
			conn.Write([]byte("OK" + " " + strconv.FormatInt(ver, 10) + "\r\n"))

		case "delete":
			if argLen != 2 || len(command[1]) > 250 {
				errMsg = "ERR_CMD_ERR"
				conn.Write([]byte(errMsg + "\r\n"))
			}

			//calling the delete function function
			err := DeleteProcess(command)

			if err != "" {
				conn.Write([]byte(err + "\r\n"))
			}
			//passing the result to client
			conn.Write([]byte("OK" + "\r\n"))

		}

	}
}

func main() {
	serverMain()
}

//function defination for write
func WriteProcess(filename []string, data []byte) (int64, error) {
	key := filename[1]
	var errflag int64 = 0
	var exp int64 = -1
	noBytes, err := strconv.ParseInt(filename[2], 0, 64)
	if err != nil {
		return errflag, err
	}
	if len(filename) == 4 {
		exp, err = strconv.ParseInt(filename[3], 0, 64)
		if err != nil {
			return errflag, err
		}
	}
	//check whether the file with mentioned name is present already in memory or not
	myLock.Lock()
	value, ok := FileMap[key]
	myLock.Unlock()
	if ok {
		//check whether file has expired or not
		if value.ExpTime != -1 { //File has expired time associated with it
			if !(value.FutureExpDate).After(time.Now()) { //expired file
				myLock.Lock()
				delete(FileMap, key)
				myLock.Unlock()
				fmt.Println("write3")
				return 1, err
			}
		}

	}
	currentTime := time.Now()
	//add given expiry time to current time
	newTime := currentTime.Add((time.Duration(exp)) * time.Second)

	ver := rand.Int63()
	myLock.Lock()
	//insert data into map
	FileMap[key] = FileAttributes{ver, noBytes, exp, data, newTime}
	myLock.Unlock()
	return ver, err
}

//function defination for read
func ReadProcess(filename []string) (int64, int64, int64, []byte, string) {

	key := filename[1]
	//check whether the file with mentioned name is present in memory or not
	myLock.RLock()
	value, ok := FileMap[key]
	myLock.RUnlock()
	if ok {
		//check whether file has expired or not
		if value.ExpTime != -1 { //File has expired time associated with it
			if (value.FutureExpDate).After(time.Now()) { //still a valid file, not yet expired
				return value.version, value.NumBytes, value.ExpTime, value.data, ""
			} else { //file has expired
				delete(FileMap, key)
				return 0, 0, 0, nil, "No File"
			}
		}
		//expired time is not present for the file to be read
		return value.version, value.NumBytes, value.ExpTime, value.data, ""
	} else {
		return 0, 0, 0, nil, "No File"
	}

}

//function defination for compare and swap
func CasProcess(filename []string, data []byte) (int64, string) {
	key := filename[1]
	rec_ver := filename[2]
	var errflag int64 = 0
	var exp int64 = -1
	noBytes, err := strconv.ParseInt(filename[3], 0, 64)
	if err != nil {
		return errflag, "ERR_INTERNAL"
	}
	if len(filename) == 5 {
		exp, err = strconv.ParseInt(filename[4], 0, 64)
		if err != nil {
			return errflag, "ERR_INTERNAL"
		}
	}
	var oldVer int64
	var ver int64
	myLock.Lock()
	value, ok := FileMap[key]
	myLock.Unlock()
	if ok {
		oldVer = value.version
		ver, err = strconv.ParseInt(rec_ver, 0, 64)
		if err != nil {
			return errflag, "ERR_INTERNAL"
		}
		//check whether file has expired or not
		if value.ExpTime != -1 { //File has expired time associated with it
			if !(value.FutureExpDate).After(time.Now()) { //expired file
				myLock.Lock()
				delete(FileMap, key)
				myLock.Unlock()
				return errflag, "ERR_FILE_NOT_FOUND"
			}
		}
	} else {
		return errflag, "ERR_FILE_NOT_FOUND"
	}

	if oldVer != ver {
		return oldVer, "ERR_VERSION"
	}
	ver = rand.Int63()
	currentTime := time.Now()
	//add given expiry time to current time
	newTime := currentTime.Add((time.Duration(exp)) * time.Second)
	myLock.Lock()
	//insert data into map
	FileMap[key] = FileAttributes{ver, noBytes, exp, data, newTime}
	myLock.Unlock()
	return ver, ""

}

//function defination for delete
func DeleteProcess(filename []string) string {
	key := filename[1]
	myLock.Lock()
	value, ok := FileMap[key]
	myLock.Unlock()
	if ok {
		//check whether file has expired or not
		if value.ExpTime != -1 {
			if (value.FutureExpDate).After(time.Now()) { //still a valid file, not yet expired
				delete(FileMap, key)
				return ""
			} else { //file has already expired
				delete(FileMap, key)
				return "ERR_FILE_NOT_FOUND"
			}
		}
		myLock.Lock()
		delete(FileMap, key)
		myLock.Unlock()
		return ""
	} else {
		return "ERR_FILE_NOT_FOUND"
	}
}
