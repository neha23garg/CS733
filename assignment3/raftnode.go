package main

import(
	"fmt"

	cluster "github.com/cs733-iitb/cluster"
	log "github.com/cs733-iitb/log"
)
// Returns a Node object
//func raft.New(config Config) Node{}

type Node interface {
	// Client's message to Raft node
	Append([]byte)
	// A channel for client to listen on. What goes into Append must come out of here at some point.
	CommitChannel() <- chan CommitInfo
	// Last known committed index in the log.This could be -1 until the system stabilizes.
	CommittedIndex() int
	// Returns the data at a log index, or an error.
	Get(index int) (error, []byte)
	// Node's id
	Id() int
	// Id of leader. -1 if unknown
	LeaderId() int
	// Signal to shut down all goroutines, stop sockets, flush log and close it, cancel timers.
	Shutdown()
	//raft state machine
	//StateMachine Server
}

// data goes in via Append, comes out as CommitInfo from the node's CommitChannel
// Index is valid only if err == nil
type  CommitInfo struct{
	Data []byte
	Index int64 // or int .. whatever you have in your code
	Err error // Err can be errred
}

// This is an example structure for Config .. change it to your convenience.
type  Config struct{
	cluster []NetConfig // Information about all servers, including this.
	Id int // this node's id. One of the cluster's entries should match.
	LogDir string // Log file directory for this node
	ElectionTimeout int
	HeartbeatTimeout int
}

type  NetConfig struct{
	Id int
	Host string
	Port int
}

type Time interface{}
// implements Node interface
type RaftNode struct { 
	eventCh chan Event
	timeoutCh chan Time
	StateMachine Server
	cluster_Server cluster.Server
}

func New(config Config) RaftNode {
	
	fmt.Println(config)

	var rNode RaftNode
	logArray := make([]LogInfo, 0)
	s := Server{myId:config.Id,currentTerm: 1, votedFor: -1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER"}
	rNode.StateMachine=s
	cluster_Config := cluster.Config{
        Peers: []cluster.PeerConfig{
            {Id: 100, Address: "localhost:7070"},
            {Id: 200, Address: "localhost:8080"},
            {Id: 300, Address: "localhost:9090"},
        }}

	server1, _ := cluster.New(config.Id, cluster_Config)
	rNode.cluster_Server=server1
  // server2, _ := cluster.New(200, cluster_Config) 
   //server3, _ := cluster.New(300, cluster_Config) 
   
	return rNode

}

func (rn *RaftNode) Append(data []byte) {
	rn.eventCh <- Append{data: data}
}

func (rn *RaftNode) ID() int {
	return rn.StateMachine.myId
}

func (rn *RaftNode) LeaderId() int {
	return rn.StateMachine.leader
}

func (rn *RaftNode) Get(index int) ([]byte, error){	
	logObject,_:=log.Open("./mylog")
	return logObject.Get(int64(index))
}

func (rn *RaftNode) CommittedIndex() int{	
	return rn.StateMachine.commitIndex
}

func makeRafts() []RaftNode{
	netConfiguration:=NetConfig{Id:100,Host:"localhost",Port:7070}
	arraynetConfiguration:=make([]NetConfig,0)
	arraynetConfiguration=append(arraynetConfiguration,netConfiguration)
	netConfiguration=NetConfig{Id:200,Host:"localhost",Port:8080}
	arraynetConfiguration=append(arraynetConfiguration,netConfiguration)
	netConfiguration=NetConfig{Id:300,Host:"localhost",Port:9090}
	arraynetConfiguration=append(arraynetConfiguration,netConfiguration)
	//fmt.Println(arraynetConfiguration)
	configuration:=Config{cluster:arraynetConfiguration,Id:100,LogDir:"log",ElectionTimeout:100,HeartbeatTimeout:100}
	node:=make([]RaftNode,0)
	result:=New(configuration)
	node=append(node,result)
	configuration=Config{cluster:arraynetConfiguration,Id:200,LogDir:"log",ElectionTimeout:100,HeartbeatTimeout:100}
	result=New(configuration)
	node=append(node,result)
	configuration=Config{cluster:arraynetConfiguration,Id:300,LogDir:"log",ElectionTimeout:100,HeartbeatTimeout:100}
	result=New(configuration)
	node=append(node,result)
	return node
	//fmt.Println(node)
}

func getLeader(rafts []RaftNode) RaftNode{
	leader:=rafts[0].LeaderId()
	raft:=rafts[0]
	for i:=0;i<len(rafts);i++ {
		if(rafts[i].ID()==leader){

			raft=rafts[i]
		}
	}
	return raft
}



