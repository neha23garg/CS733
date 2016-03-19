package main

import (
	//"encoding/json"
	"fmt"
	"testing"
	"time"
	//"io/ioutil"
)

func TestBasic(t *testing.T) {

	//all statemechine are in follower state and leader election begins
	for i := 1; i <= 5; i++ {
		RemoveDir(i)
	}
	rafts := make([]RaftNode, 5)

	var ldrId int
	noLeader := true
	for noLeader {
		rafts = makeRafts() // array of []raft.Node		
		time.Sleep(5 * time.Second)
		//get leader of the raft nodes
		ldr := getLeader(rafts)
		if ldr != nil {
			//get leader id
			ldrId = ldr.ID()

			if ldrId == 1 || ldrId == 2 || ldrId == 3 || ldrId == 4 || ldrId == 5 {
				noLeader = false

			}
		}

	}
	//get leader of the raft nodes
	ldr := getLeader(rafts)
	//check the state of leader
	if ldr.StateMachine.State == "LEADER" {
		for i := 0; i < len(rafts); i++ {
			if i+1 != ldrId {
				expectNotString(t, rafts[i].StateMachine.State, "LEADER")
			}
		}
	}
	//get leader of the raft nodes
	ldr = getLeader(rafts)

	ldrId = ldr.ID()
	//append data on leader
	if ldrId == 1 || ldrId == 2 || ldrId == 3 || ldrId == 4 || ldrId == 5 {
		ldr.Append([]byte("foo"))
	}
	time.Sleep(1 * time.Second)

	time.Sleep(2 * time.Second)
	ldrId = ldr.ID()
	//check whether the data send by leader is present on commit channel of all other servers
	if ldrId == 1 || ldrId == 2 || ldrId == 3 || ldrId == 4 || ldrId == 5 {
		for _, node := range rafts {
			select {
			// to avoid blocking on channel.
			case ci := <-node.CommitChannel():
				if ci.Error != "" {
					t.Fatal(ci.Error)
				}
				if string(ci.Data) != "foo" {
					t.Fatal("Got different data")
				}

			default:
				t.Fatal("Expected message on all nodes")
			}
		}
	}
	//check whether the data send by leader is present in the log file of all other servers
	for i := 0; i < len(rafts); i++ {
		data, err := rafts[i].nodeLog.Get(0) // should return the Foo instance appended above
		if err == nil {
			foo, ok := data.(LogStore)
			if ok {
				expectString(t, "foo", string(foo.Data))
			}
		}
	}

	ldr = getLeader(rafts)
	ldr.StateMachine.NextIndex = make([]int, 10)
	//append second entry on leader
	ldrId = ldr.ID()
	if ldrId == 1 || ldrId == 2 || ldrId == 3 || ldrId == 4 || ldrId == 5 {
		ldr.Append([]byte("hello"))
	}
	time.Sleep(1 * time.Second)

	time.Sleep(1 * time.Second)
	ldrId = ldr.ID()
	//check whether the data send by leader is present on commit channel of all other servers
	if ldrId == 1 || ldrId == 2 || ldrId == 3 || ldrId == 4 || ldrId == 5 {
		for _, node := range rafts {
			select {
			// to avoid blocking on channel.
			case ci := <-node.CommitChannel():
				if ci.Error != "" {
					t.Fatal(ci.Error)
				}
				if string(ci.Data) != "hello" {
					t.Fatal("Got different data")
				}

			default:
				t.Fatal("Expected message on all nodes")
			}
		}
	}
	//check whether the data send by leader is present in the log file of all other servers
	for i := 0; i < len(rafts); i++ {
		data, err := rafts[i].nodeLog.Get(1) // should return the Foo instance appended above
		if err == nil {
			foo, ok := data.(LogStore)
			if ok {
				expectString(t, "hello", string(foo.Data))
			}

			//expectString(t,data,[]byte("foo"))
		}
	}

	/*file, e := ioutil.ReadFile("state.json")
	//myLock.Unlock()
	if e != nil {
		fmt.Printf("File error: %v\n", e)
	}

	var objArray []PersistantStore
	json.Unmarshal(file, &objArray)
	for i := 0; i < 5; i++ {
		objArray[i].Term = 0
		objArray[i].VotedFor = 0
	}

	fileJson, _ := json.Marshal(objArray)

	err := ioutil.WriteFile("state.json", fileJson, 0644)
	//  myLock.Unlock()
	if err != nil {
		fmt.Printf("WriteFileJson ERROR: %+v", err)
	}*/

	//test case for partitioning of cluster and leader is in the partition where  majority is there
	ldr.PartitionLeaderInMajority()
	time.Sleep(3 * time.Second)
	if ldr.ID()%2 == 0 {
		//Raft node 3 and Raft node 5 can be in candidate/follower state but can't be in leader state

		expectNotString(t, rafts[2].StateMachine.State, "LEADER")
		expectNotString(t, rafts[4].StateMachine.State, "LEADER")
		//existing leader will remain in leader state
		expectString(t, ldr.StateMachine.State, "LEADER")
	} else {
		//Raft node 2 and Raft node 4 can be in candidate/follower state but can't be in leader state

		expectNotString(t, rafts[1].StateMachine.State, "LEADER")
		expectNotString(t, rafts[3].StateMachine.State, "LEADER")
		//existing leader will remain in leader state
		expectString(t, ldr.StateMachine.State, "LEADER")
	}

	//heal the cluster that is dissolve the partition
	rafts[0].Heal()
	time.Sleep(3 * time.Second)
	//Now there should be exactly one leader in the cluster
	ldr = getLeader(rafts)

	for i := 0; i < len(rafts); i++ {
		if i+1 != ldr.ID() {
			expectNotString(t, rafts[i].StateMachine.State, "LEADER")
		}
	}

	//test case for partitioning of cluster and leader is in the partition where manority is there
	ldr.PartitionLeaderInManority()
	time.Sleep(3 * time.Second)
	if ldr.ID()%2 == 0 {
		//Leader can be elected from the partition having raft node 1, 3, 5
		if rafts[0].StateMachine.State == "LEADER" {

			expectNotString(t, rafts[2].StateMachine.State, "LEADER")
			expectNotString(t, rafts[4].StateMachine.State, "LEADER")
		} else if rafts[2].StateMachine.State == "LEADER" {

			expectNotString(t, rafts[0].StateMachine.State, "LEADER")
			expectNotString(t, rafts[4].StateMachine.State, "LEADER")
		} else if rafts[4].StateMachine.State == "LEADER" {

			expectNotString(t, rafts[0].StateMachine.State, "LEADER")
			expectNotString(t, rafts[2].StateMachine.State, "LEADER")
		}
	} else {
		if ldr.ID() == 1 {
			//Leader can be elected from the partition having raft node 3,4,5
			if rafts[2].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[3].StateMachine.State, "LEADER")
				expectNotString(t, rafts[4].StateMachine.State, "LEADER")
			} else if rafts[3].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[2].StateMachine.State, "LEADER")
				expectNotString(t, rafts[4].StateMachine.State, "LEADER")
			} else if rafts[4].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[3].StateMachine.State, "LEADER")
				expectNotString(t, rafts[2].StateMachine.State, "LEADER")
			}
		} else {
			//Leader can be elected from the partition having raft node 1,2,4
			if rafts[0].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[1].StateMachine.State, "LEADER")
				expectNotString(t, rafts[3].StateMachine.State, "LEADER")
			} else if rafts[1].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[0].StateMachine.State, "LEADER")
				expectNotString(t, rafts[3].StateMachine.State, "LEADER")
			} else if rafts[3].StateMachine.State == "LEADER" {

				expectNotString(t, rafts[0].StateMachine.State, "LEADER")
				expectNotString(t, rafts[1].StateMachine.State, "LEADER")
			}
		}

	}
	//heal the cluster that is dissolve the partition
	rafts[0].Heal()
	time.Sleep(3 * time.Second)
	//Now there should be exactly one leader in the cluster
	ldr = getLeader(rafts)
	oldldrId := ldr.ID()

	//Check if new leader is elected then shut it down
	if ldr.ID() == 1 || ldr.ID() == 2 || ldr.ID() == 3 || ldr.ID() == 4 || ldr.ID() == 5 {

		ldr.Shutdown()
		time.Sleep(3 * time.Second)

		newldr := getNewLeader(rafts, oldldrId)

		if newldr.status == true {
			//new leader can be selected from the remaining raft nodes
			expectNotInt(t, oldldrId, newldr.ID())
		}
	}

	for i := 1; i <= 5; i++ {
		RemoveDir(i)
	}

}

func expectInt(t *testing.T, a int, b int) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}

func expectNotInt(t *testing.T, a int, b int) {
	if a == b {
		t.Error(fmt.Sprintf("Not Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}

func expectString(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}

func expectNotString(t *testing.T, a string, b string) {
	if a == b {
		t.Error(fmt.Sprintf("Not Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
