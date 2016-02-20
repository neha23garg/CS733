package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	//"io/ioutil"
)

// Simple serial check of getting and setting
func TCPSimple(t *testing.T, wg *sync.WaitGroup) {
	time.Sleep(1 * time.Second)
	name := "hi.txt"
	contents := "bye"
	exptime := 300000
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error()) // report error through testing framework
	}

	scanner := bufio.NewScanner(conn)

	// Write a file
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()                  // read first line
	resp := scanner.Text()          // extract the text from the buffer
	arr := strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	ver, err := strconv.Atoi(arr[1]) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version := int64(ver)
	//fmt.Println("----writeone----------")
	fmt.Println(arr[0], " ", version)

	// try a read now
	fmt.Fprintf(conn, "read %v\r\n", name)
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	//expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
	if len(arr) == 4 {
		fmt.Println(arr[0], arr[1], arr[2], arr[3])
	} else {
		fmt.Println(arr[0], arr[1], arr[2])
	}
	scanner.Scan()
	//fmt.Println("-------readone-----------")

	expect(t, contents, scanner.Text())
	fmt.Println(scanner.Text())
	//time.Sleep(10* time.Second)

	//write a file
	//name = "hi.txt"
	contents = "two"
	exptime = 3
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()                 // read first line
	resp = scanner.Text()          // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	ver, err = strconv.Atoi(arr[1]) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version = int64(ver)
	//fmt.Println("--------writetwo-------------")
	fmt.Println(arr[0], " ", version)
	//time.Sleep(10* time.Second)
	// try a read now
	fmt.Fprintf(conn, "read %v\r\n", name)
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	//expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
	if len(arr) == 4 {
		fmt.Println(arr[0], arr[1], arr[2], arr[3])
	} else {
		fmt.Println(arr[0], arr[1], arr[2])
	}
	scanner.Scan()
	//fmt.Println("-------readtwo-----------")
	expect(t, contents, scanner.Text())
	fmt.Println(scanner.Text())

	// compare and swap
	fmt.Fprintf(conn, "cas %v %v %v %v\r\n%v\r\n", name, version, len(contents), exptime, contents)
	scanner.Scan()                 // read first line
	resp = scanner.Text()          // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	//expect(t, arr[0], "OK")
	ver, err = strconv.Atoi(arr[1]) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version = int64(ver)
	//fmt.Println("--------------casone----------")
	fmt.Println(arr[0], " ", version)

	//delete
	fmt.Fprintf(conn, "delete %v\r\n", name)
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	//expect(t, arr[0], "OK")
	fmt.Println(arr[0])

	// try a read now
	fmt.Fprintf(conn, "read %v\r\n", name)
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	//expect(t, arr[0], "ERR_FILE_NOT_FOUND")

	fmt.Println(scanner.Text())

	//exptime=1
	// Write a file
	/*fmt.Println("actualcontent",contents)
		fmt.Fprintf(conn, "write %v %v\r\n%v\r\n", name, len(contents), contents)
		scanner.Scan() // read first line
		resp = scanner.Text() // extract the text from the buffer
		arr = strings.Split(resp, " ") // split into OK and <version>
		//expect(t, arr[0], "OK")
		fmt.Println("resp**************",string(resp))
		ver, err = strconv.Atoi(arr[1]) // parse version as number
		if err != nil {
			t.Error("Non-numeric version found")
		}
		version = int64(ver)
		fmt.Println(arr[0]," ",version)

		 //time.Sleep(2 * time.Second)
		// try a read now
	/*	fmt.Fprintf(conn, "read %v\r\n", name)
		scanner.Scan()

		arr = strings.Split(scanner.Text(), " ")
		expect(t, arr[0], "CONTENTS")
		expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
		expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
		scanner.Scan()
		fmt.Println("-------------readthree--------------")
		fmt.Println(scanner.Text())
		expect(t, contents, scanner.Text())

		 exptime=2
		// compare and swap
		fmt.Fprintf(conn, "cas %v %v %v %v\r\n%v\r\n", name,version ,len(contents), exptime, contents)
		scanner.Scan() // read first line
		resp = scanner.Text() // extract the text from the buffer
		arr = strings.Split(resp, " ") // split into OK and <version>
		expect(t, arr[0], "OK")
		ver, err = strconv.Atoi(arr[1]) // parse version as number
		if err != nil {
			t.Error("Non-numeric version found")
		}
		version = int64(ver)
		fmt.Println("--------------casone-------------")
		fmt.Println(version)

		//time.Sleep(1 * time.Second)
		fmt.Println("------hi-------------")
		// try a read now
	/*	fmt.Println("read %v\r\n", name)
		fmt.Fprintf(conn, "read %v\r\n", name)
		scanner.Scan()
		fmt.Println(scanner.Text())
		arr = strings.Split(scanner.Text(), " ")
		expect(t, arr[0], "CONTENTS")
		expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
		expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
		scanner.Scan()
		fmt.Println("----------------readfour--------------")
		fmt.Println(scanner.Text())
		expect(t, contents, scanner.Text())*/
	wg.Done()
}

// Useful testing function
func expect(t *testing.T, a string, b string) {
	//fmt.Println("a",a,b)
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}

func TestMultipleClient(t *testing.T) {
	go serverMain()
	var wg sync.WaitGroup

	wg.Add(5) //We need to wait for 3 calls to 'done' on this wait group
	for i := 0; i < 5; i++ {
		go TCPSimple(t, &wg)
	}

	wg.Wait()

}
