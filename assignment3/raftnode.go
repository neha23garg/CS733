package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	cluster "github.com/cs733-iitb/cluster"
	mock "github.com/cs733-iitb/cluster/mock"
	log "github.com/cs733-iitb/log"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type Node interface {
	// Client's message to Raft node
	Append([]byte)
	// A channel for client to listen on. What goes into Append must come out of here at some point.
	CommitChannel() <-chan Commit
	// Last known committed index in the Log.This could be -1 until the system stabilizes.
	CommittedIndex() int
	// Returns the data at a Log index, or an error.
	Get(index int) (error, []byte)
	// Node's id
	Id() int
	// Id of Leader. -1 if unknown
	LeaderId() int
	// Signal to shut down all goroutines, stop sockets, flush Log and close it, cancel timers.
	Shutdown()
}
//to store term and votedfor in non-volatile memory
type PersistantStore struct {
	Id       int
	Term     int
	VotedFor int
}

// This is an example structure for Config .. change it to your convenience.
type Config struct {
	cluster            []NetConfig // Information about all servers, including this.
	Id                 int         // this node's id. One of the cluster's entries should match.
	LogDir             string      // Log file directory for this node
	ElectionTimeout    int
	HeartbeatTimeout   int
	config_MockCluster *mock.MockCluster
}

type NetConfig struct {
	Id   int
	Host string
	Port int
}

type Time interface{}

// implements Node interface
type RaftNode struct {
	eventCh        chan Event
	timeoutCh      chan Time
	StateMachine   *Server
	cluster_Server cluster.Server
	CommitChan     chan Commit
	Timer          *time.Timer
	HeartbeatTimer *time.Ticker
	shutdownch     chan int
	status         bool
	nodeLog        *log.Log
	cluster        *mock.MockCluster
	//timeoutchLeader chan Time
	//RandNo int
}

//declaration of wait group
var wg sync.WaitGroup

//declaration of lock variable
var myLock = &sync.RWMutex{}

//Initialize the instance of every raft node 
func New(config Config) RaftNode {
	var rNode RaftNode
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4, 5)
	voteArray := make([]int, 6)
	matchInd := make([]int, 10)
	nextInd := make([]int, 10)
	//get votedfor, term from persistant storage
	var votefor int
	var term int
	file, e := ioutil.ReadFile("state.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
	}

	var objArray []PersistantStore
	json.Unmarshal(file, &objArray)
	for i := 0; i < len(objArray); i++ {
		if objArray[i].Id == config.Id {
			votefor = objArray[i].VotedFor
			term = objArray[i].Term
		}
	}
	//intialize logs for the node
	lg, err := log.Open(config.LogDir)
	if err != nil {
		fmt.Println("Error : Log Open() failed", err)

	}
	lg.SetCacheSize(50)
	rNode.nodeLog = lg

	// get logs stored in the files
	logLastIndex := rNode.nodeLog.GetLastIndex()
	var startindex int64
	startindex = 0
	//restore logs
	for logLastIndex >= 0 {
		data, err := rNode.nodeLog.Get(startindex) // should return the Foo instance appended above
		if err == nil {
			foo, ok := data.(LogStore)

			if ok {

				LogArray = append(LogArray, LogInfo{Term: term, Data: foo.Data})
			}
		}
		startindex++
		logLastIndex--

	}
	//rNode.nodeLog.Close()
	//initialize state machine instance
	s := &Server{MyId: config.Id, CurrentTerm: term, VotedFor: votefor, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray, VoteGranted: voteArray, NextIndex: nextInd, MatchIndex: matchInd}
	rNode.StateMachine = s
	// Give each raftNode its own "Server" from the cluster.
	rNode.cluster_Server = config.config_MockCluster.Servers[config.Id]
	rNode.cluster = config.config_MockCluster

	//rNode.cluster_Server=server1
	rNode.eventCh = make(chan Event, 500)
	rNode.CommitChan = make(chan Commit, 500)
	rNode.timeoutCh = make(chan Time)
	rNode.shutdownch = make(chan int)

	//initially all raft nodes are running
	rNode.status = true
	rand.Seed(time.Now().UnixNano())
	RandNo := rand.Intn(500)
	//set election timeout
	rNode.Timer = time.AfterFunc(time.Duration(config.ElectionTimeout+RandNo)*time.Millisecond, func() { rNode.timeoutCh <- Timeout{} })
	//set heartbeat timer
	rNode.HeartbeatTimer = time.NewTicker(time.Duration(config.HeartbeatTimeout) * time.Millisecond)
	//return node object
	return rNode

}

//handle incoming append request
func (rn *RaftNode) Append(Data []byte) {
	rn.eventCh <- Append{Data: Data}
}
//handle timout event
func (rn *RaftNode) TimeOut() {
	rn.timeoutCh <- Timeout{}

}

func (rn *RaftNode) ID() int {
	return rn.StateMachine.MyId
}

func (rn *RaftNode) LeaderId() int {
	return rn.StateMachine.Leader
}

//commit channel method
func (rn *RaftNode) CommitChannel() chan Commit {

	return rn.CommitChan
}

//remove log directory method
func RemoveDir(i int) {

	os.RemoveAll("./NodeLog" + strconv.Itoa(i))

}

func (rn *RaftNode) CommittedIndex() int {
	return rn.StateMachine.CommitIndex
}

func (rn *RaftNode) Shutdown() {

	wg.Add(1)
	close(rn.shutdownch)
}

func makeRafts() []RaftNode {
	Register()
	netConfiguration := NetConfig{Id: 1, Host: "localhost", Port: 8900}
	arraynetConfiguration := make([]NetConfig, 0)
	arraynetConfiguration = append(arraynetConfiguration, netConfiguration)
	netConfiguration = NetConfig{Id: 2, Host: "localhost", Port: 8000}
	arraynetConfiguration = append(arraynetConfiguration, netConfiguration)
	netConfiguration = NetConfig{Id: 3, Host: "localhost", Port: 9000}
	arraynetConfiguration = append(arraynetConfiguration, netConfiguration)
	netConfiguration = NetConfig{Id: 4, Host: "localhost", Port: 9022}
	arraynetConfiguration = append(arraynetConfiguration, netConfiguration)
	netConfiguration = NetConfig{Id: 5, Host: "localhost", Port: 9011}
	arraynetConfiguration = append(arraynetConfiguration, netConfiguration)
	//initialize peer config file
	cluster_Config := cluster.Config{
		Peers: []cluster.PeerConfig{
			{Id: 1},
			{Id: 2},
			{Id: 3},
			{Id: 4},
			{Id: 5},
		}}

	//create cluster
	cl, err := mock.NewCluster(cluster_Config)
	if err != nil {
		fmt.Println("error", err)
	}

	configuration := Config{cluster: arraynetConfiguration, Id: 1, LogDir: "./NodeLog1", ElectionTimeout: 500, HeartbeatTimeout: 250, config_MockCluster: cl}
	node := make([]RaftNode, 0)
	result1 := New(configuration)
	// Give each raftNode its own "Server" from the cluster.
	//result1.cluster_Server = cl.Servers[1]
	go result1.ProcessEvents()
	node = append(node, result1)
	configuration = Config{cluster: arraynetConfiguration, Id: 2, LogDir: "./NodeLog2", ElectionTimeout: 500, HeartbeatTimeout: 250, config_MockCluster: cl}
	result2 := New(configuration)
	go result2.ProcessEvents()
	node = append(node, result2)
	configuration = Config{cluster: arraynetConfiguration, Id: 3, LogDir: "./NodeLog3", ElectionTimeout: 500, HeartbeatTimeout: 250, config_MockCluster: cl}
	result3 := New(configuration)
	go result3.ProcessEvents()
	node = append(node, result3)
	configuration = Config{cluster: arraynetConfiguration, Id: 4, LogDir: "./NodeLog4", ElectionTimeout: 500, HeartbeatTimeout: 250, config_MockCluster: cl}
	result4 := New(configuration)
	go result4.ProcessEvents()
	node = append(node, result4)
	configuration = Config{cluster: arraynetConfiguration, Id: 5, LogDir: "./NodeLog5", ElectionTimeout: 500, HeartbeatTimeout: 250, config_MockCluster: cl}
	result5 := New(configuration)
	go result5.ProcessEvents()
	node = append(node, result5)
	return node

}

func getLeader(rafts []RaftNode) *RaftNode {
	//Leader:=rafts[0].LeaderId()
	var leader *RaftNode
	leader = nil
	for i := 0; i < len(rafts); i++ {
		if rafts[i].status == true && rafts[i].StateMachine.State == "LEADER" {
			leader = &(rafts[i])

		}

	}

	return leader
}
//process channels events
func (rn *RaftNode) ProcessEvents() {
	actions := make([]Action, 0)
	for {
		var ev Event
		select {
		case ev = <-rn.eventCh:

			actions = rn.StateMachine.ProcessEvent(ev)
			rn.DoActions(actions)

		case ev = <-rn.timeoutCh:

			actions = rn.StateMachine.ProcessEvent(ev)
			rn.DoActions(actions)

		case env := <-rn.cluster_Server.Inbox():
			actions = rn.StateMachine.ProcessEvent(env.Msg)

			rn.DoActions(actions)
		case <-rn.HeartbeatTimer.C:
			if rn.StateMachine.State == "LEADER" {
				rn.eventCh <- Timeout{}

			}

		case _, ok := <-rn.shutdownch:
			if !ok {

				rn.status = false
				rn.Timer.Stop()
				rn.HeartbeatTimer.Stop()

				rn.cluster_Server.Close()
				wg.Done()
			}

			return

		default:
		}

	}
}
//handle actions generated by statemachine
func (rn *RaftNode) DoActions(actions []Action) {
	for _, action := range actions {
		switch action.(type) {
		case Send:
			//reset timer
			if rn.StateMachine.State != "LEADER" {
				rn.Timer.Stop()
				rn.Timer = time.AfterFunc(time.Duration(1000+rand.Intn(150))*time.Millisecond, func() { rn.timeoutCh <- Timeout{} })
			}
			obj := action.(Send)
			//fmt.Println("EvENT",reflect.TypeOf(obj.Event).Name(),obj.Event,rn.StateMachine.State,rn.StateMachine.MyId)
			rn.cluster_Server.Outbox() <- &cluster.Envelope{Pid: obj.From, Msg: obj.Event}

		case Alarm:
			//reset timer
			rn.Timer.Stop()
			rn.Timer = time.AfterFunc(time.Duration(1000+rand.Intn(150))*time.Millisecond, func() { rn.timeoutCh <- Timeout{} })

		case Commit:
			obj := action.(Commit)
			rn.CommitChan <- obj

		case StateStore:
			obj := action.(StateStore)
			term := obj.Term
			votefor := obj.VotedFor
			myLock.Lock()
			file, e := ioutil.ReadFile("state.json")
			//myLock.Unlock()
			if e != nil {
				fmt.Printf("File error: %v\n", e)
			}

			var objArray []PersistantStore
			json.Unmarshal(file, &objArray)

			objArray[rn.ID()-1].Term = term
			objArray[rn.ID()-1].VotedFor = votefor

			fileJson, _ := json.Marshal(objArray)

			err := ioutil.WriteFile("state.json", fileJson, 0644)
			myLock.Unlock()
			if err != nil {
				fmt.Printf("WriteFileJson ERROR: %+v", err)
			}

		case LogStore:

			i := rn.nodeLog.GetLastIndex()
			obj := action.(LogStore)
			index := obj.Index
			if i > int64(index) {
				rn.nodeLog.TruncateToEnd(int64(index))
			}
			rn.nodeLog.Append(action)

		}
	}
}

func getNewLeader(rafts []RaftNode, old int) *RaftNode {
	//Leader:=rafts[0].LeaderId()
	var leader *RaftNode
	leader = nil
	for i := 0; i < len(rafts); i++ {
		if i+1 != old && rafts[i].StateMachine.State == "LEADER" {
			leader = &(rafts[i])

		}

	}

	return leader
}

func (rn *RaftNode) PartitionLeaderInMajority() {
	if rn.ID()%2 == 0 {
		rn.cluster.Partition([]int{3, 5}, []int{1, 2, 4})
	} else {
		rn.cluster.Partition([]int{2, 4}, []int{1, 3, 5})
	}
}

func (rn *RaftNode) PartitionLeaderInManority() {
	if rn.ID()%2 == 0 {
		rn.cluster.Partition([]int{1, 3, 5}, []int{2, 4})
	} else {
		if rn.ID() == 1 {
			rn.cluster.Partition([]int{1, 2}, []int{3, 4, 5})
		} else {
			rn.cluster.Partition([]int{1, 2, 4}, []int{3, 5})
		}
	}
}

func (rn *RaftNode) Heal() {

	rn.cluster.Heal()
}

func Register() {
	gob.Register(VoteReq{})
	gob.Register(VoteResp{})
	gob.Register(Alarm{})
	gob.Register(AppendEntriesReq{})
	gob.Register(AppendEntriesResp{})
	gob.Register(LogStore{})
	gob.Register(StateStore{})

}
