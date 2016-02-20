package main

import (
	"fmt"
	"sort"
	"strconv"
)

//attributs of servers
type Server struct {
	currentTerm int
	votedFor    int
	log         []LogInfo
	commitIndex int
	nextIndex   []int
	matchIndex  []int
	leader      int
	state       string
	peers       []int
	voteGranted []int
}

// attributes of log
type LogInfo struct {
	term int
	data []byte
}

//assign severid and number of servers
var serverId int = 5
var quorumSize int = 5

// VoteRequest RPC
type VoteReq struct {
	//from 		 int
	term         int // the term
	candidateId  int
	lastLogIndex int
	lastLogTerm  int
}

// VoteResponse
type VoteResp struct {
	term        int
	voteGranted bool
	voterId     int
}

//append entries RPC
type AppendEntriesReq struct {
	term         int
	leaderId     int
	lastLogIndex int
	lastLogTerm  int
	entries      []LogInfo
	leaderCommit int
}

//append entries response
type AppendEntriesResp struct {
	from    int
	term    int
	success bool
	count   int
	lastLogIndex int
}

//append request from client struct
type Append struct{
	data []byte
}

type Timeout struct {
}

type Event interface{}

type Action interface{}

type Send struct {
	from int
	//event can be vote req/resp and append rpc request/response
	event Event
}

type Alarm struct {
	t int
}

type LogStore struct {
	index int
	data  []byte
}

type Commit struct {
	index int
	data  []byte
	error string
}

type StateStore struct {
	term     int
	votedFor int
}

//Function to handle all incoming request to RAFT State Machine
func (s *Server) ProcessEvent(inputEvents Event) []Action {
	result := make([]Action, 0)
	//check state of the server
	if s.state == "FOLLOWER" {
		switch inputEvents.(type) {
		//Handle Vote request coming to follower
		case VoteReq:
			voteReqObj := inputEvents.(VoteReq)
			result = OnVoteReqFollower(s, voteReqObj)
		//Handle Vote response coming to candidate
		case VoteResp:
			voteRespObj := inputEvents.(VoteResp)
			result = OnVoteRespFollower(s, voteRespObj)
		//Handle append rpc request coming to follower
		case AppendEntriesReq:
			AppendEntriesReqObj := inputEvents.(AppendEntriesReq)
			result = OnAppendEntriesReqFollower(s, AppendEntriesReqObj)
		//Handle Vote response coming to candidate
		case AppendEntriesResp:
			AppendEntriesRespObj := inputEvents.(AppendEntriesResp)
			result = OnAppendEntriesRespFollower(s, AppendEntriesRespObj)
		//Handle timeout event
		case Timeout:
			result = onTimeOutFollower(s)
		//Handle append entry request made by client
		case Append:
			appendObj:=inputEvents.(Append)
			result=OnAppendRequestFromClientToFollower(s,appendObj)
		}

	} else if s.state == "CANDIDATE" {

		switch inputEvents.(type) {
		//Handle Vote request coming to candidate
		case VoteReq:
			voteReqObj := inputEvents.(VoteReq)
			result = OnVoteReqCandidate(s, voteReqObj)
		//Handle Vote response coming to candidate
		case VoteResp:
			voteRespObj := inputEvents.(VoteResp)
			result = OnVoteRespCandidate(s, voteRespObj)
		//Handle timeout event
		case Timeout:
			result = onTimeOutCandidate(s)
		//Handle append rpc request coming to candidate
		case AppendEntriesReq:
			AppendEntriesReqObj := inputEvents.(AppendEntriesReq)
			result = OnAppendEntriesReqCandidate(s, AppendEntriesReqObj)
		//Handle Vote response coming to candidate
		case AppendEntriesResp:
			AppendEntriesRespObj := inputEvents.(AppendEntriesResp)
			result = OnAppendEntriesRespCandidate(s, AppendEntriesRespObj)
		//Handle append entry request made by client
		case Append:
			appendObj:=inputEvents.(Append)
			result=OnAppendRequestFromClientToCandidate(s,appendObj)

		}
	} else if s.state == "LEADER" {

		switch inputEvents.(type) {
		//Handle Vote request coming to leader
		case VoteReq:
			voteReqObj := inputEvents.(VoteReq)
			result = OnVoteReqLeader(s, voteReqObj)
		//Handle Vote response coming to leader
		case VoteResp:
			voteRespObj := inputEvents.(VoteResp)
			result = OnVoteRespLeader(s, voteRespObj)
		//Handle timeout event
		case Timeout:
			result = onTimeOutLeader(s)
		//Handle append entry request
		case AppendEntriesReq:
			AppendEntriesReqObj := inputEvents.(AppendEntriesReq)
			result = OnAppendEntriesReqLeader(s, AppendEntriesReqObj)
		//Handle append entry request made by client
		case Append:
			appendObj:=inputEvents.(Append)
			result=OnAppendRequestFromClientToLeader(s,appendObj)
		//Handle Vote response coming to Leader
		case AppendEntriesResp:
			AppendEntriesRespObj := inputEvents.(AppendEntriesResp)
			result = OnAppendEntriesRespLeader(s, AppendEntriesRespObj)
		}
	} else {
		fmt.Println("NO STATE")
	}

	return result
}

//*********************FOLLOWER STATE EVENTS*************************************************************
//*******************************************************************************************************

//Handle Vote request coming to follower
func OnVoteReqFollower(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	logIndex := len(s.log)
	logTerm := 0
	if logIndex > 0 {
		//get last term from logs
		logTerm = s.log[logIndex-1].term
	}

	var latestTerm int = 0
	if s.currentTerm < msg.term {
		//store latest term
		latestTerm = msg.term
	}
	if s.currentTerm > msg.term {
		// received request term is less then  server's current term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if logTerm > 0 && logTerm > msg.lastLogTerm {
		// received request last log term is less then server's last log term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if (logTerm == msg.lastLogTerm) && logIndex > 0 && logIndex-1 > msg.lastLogIndex {
		// received request last log term is equal to server's last log term but last log index differs then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if s.votedFor == 0 || s.votedFor == msg.candidateId {
		//if all criteria satisfies and follower has not voted for this term or voted for the same candidate id then vote and update current term,voted for
		s.currentTerm = msg.term
		s.votedFor = msg.candidateId
		//send the vote	response
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: true, voterId: serverId}})
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	} else {
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	}
	if latestTerm > s.currentTerm {
		//In case servers has not granted his vote to the sever with higher term, then update the term and clear votedfor
		s.currentTerm = latestTerm
		s.votedFor = 0
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	//return action
	return actionArray
}

//Handle vote resp request coming to follower
func OnVoteRespFollower(s *Server, msg VoteResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's term if term in msg is greater than it's current term
	if(s.currentTerm<msg.term){
		s.currentTerm=msg.term
		s.votedFor=0
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	return actionArray
}

//Handle Append entry request coming to follower
func OnAppendEntriesReqFollower(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	logLength := len(s.log)
	var latestTerm int = 0
	if s.currentTerm < msg.term {
		//store latest term
		latestTerm = msg.term
	}
	if s.currentTerm > msg.term {
		// received request term is less then  server's current term then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if logLength < msg.lastLogIndex {
		//last log index of follower log not matches with received last log index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if msg.lastLogIndex != -1 && logLength != 0 && s.log[msg.lastLogIndex].term != msg.lastLogTerm {
		//last log index of follower log  matches with received last log index but terms differs on that index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else {
		//delete enteries from log starting from lastLogIndex+1 to end
		i := msg.lastLogIndex + 1
		//check data
		if logLength > i {
			s.log = append(s.log[:i], s.log[:i+1]...)
		}
		if len(msg.entries) > 0 {
			//send log store action to raft node
			actionArray = append(actionArray, LogStore{index: i, data: msg.entries[msg.lastLogIndex+1].data})
			s.log = append(s.log, msg.entries...)
		}
		//set leader id
		s.leader = msg.leaderId
		//update term and voted for if term in append entry rpc is greater than follower current term
		if latestTerm > s.currentTerm {
			s.currentTerm = latestTerm
			s.votedFor = 0
			//send state store action to raft node
			actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
		}
		if msg.leaderCommit > s.commitIndex {
			s.commitIndex = min(msg.leaderCommit, i-1)
			//check data index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{index: s.commitIndex, data: s.log[s.commitIndex].data})
		}
		//send append entry action for the received request
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: true,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})

	}
	if latestTerm > s.currentTerm {
		//In case follower has rejected append entry request from the sever with higher term, then update the term and clear votedfor
		s.currentTerm = latestTerm
		s.votedFor = 0
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	//return action
	return actionArray
}

//Handle append entry resp request coming to follower
func  OnAppendEntriesRespFollower(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's term if term in msg is greater than it's current term
	if(s.currentTerm<msg.term){
		s.currentTerm=msg.term
		s.votedFor=0
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	return actionArray
}

//Handle follower Timeout
func onTimeOutFollower(s *Server) []Action {
	actionArray := make([]Action, 0)
	//convert to candidate
	s.state = "CANDIDATE"
	//increase current term by 1
	s.currentTerm = s.currentTerm + 1
	//vote for self
	s.votedFor = serverId
	//send state store action
	actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	//send vote request message to peers
	lastIndex := -1
	lastTerm := -1
	if len(s.log) > 0 {
		//get last log index and term
		lastIndex = len(s.log) - 1
		lastTerm = s.log[lastIndex].term
	}

	for i := 0; i < len(s.peers); i++ {
		//vote request action
		actionArray = append(actionArray, Send{from: s.peers[i], event: VoteReq{term: s.currentTerm, candidateId: serverId, lastLogIndex: lastIndex, lastLogTerm: lastTerm}})
	}
	//reset timer
	actionArray = append(actionArray, Alarm{10})
	return actionArray

}
//Handle coming append  request from client to follower
func OnAppendRequestFromClientToFollower(s *Server, msg Append) []Action {
	actionArray := make([]Action, 0)
	actionArray=append(actionArray,Commit{data:msg.data,error:"Leader is at Id "+strconv.Itoa(s.leader)})
	return actionArray
}
//********************************************Candidate events***********************************************************
//*************************************************************************************************************************
//Handle Coming Vote request to Candidate
func OnVoteReqCandidate(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	logIndex := len(s.log)
	logTerm := 0
	if logIndex > 0 {
		//get last term from logs
		logTerm = s.log[logIndex-1].term
	}
	var latestTerm int = 0
	if s.currentTerm < msg.term {
		//store latest term
		latestTerm = msg.term
	}

	if s.currentTerm >= msg.term {
		// received request term is less then or equal server's current term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if logTerm > 0 && logTerm > msg.lastLogTerm {
		// received request last log term is less then server's last log term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if (logTerm == msg.lastLogTerm) && logIndex > 0 && logIndex-1 > msg.lastLogIndex {
		// received request last log term is equal to server's last log term but last log index differs then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else {
		//if all criteria satisfies and follower has not voted for this higher term  then vote and update current term,voted for
		s.currentTerm = msg.term
		s.votedFor = msg.candidateId
		//convert to follower state
		s.state = "FOLLOWER"
		//send the vote	response
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: true, voterId: serverId}})
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	if latestTerm > s.currentTerm {
		s.currentTerm = latestTerm
		s.votedFor = 0
		//convert to follower state
		s.state = "FOLLOWER"
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	return actionArray
}

//Handle Vote resp events in candidate state
func OnVoteRespCandidate(s *Server, msg VoteResp) []Action {
	actionArray := make([]Action, 0)
	//As candidate has voted for himself
	yesVoteCount := 1
	noVoteCount := 0

	if s.currentTerm < msg.term {
		//If msg term is higher than current term convert to follower
		s.currentTerm = msg.term
		s.votedFor = 0
		s.state = "FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
		//reset timer
		actionArray = append(actionArray, Alarm{10})
	} else if s.currentTerm == msg.term && msg.voteGranted == true {
		//collect yes votes
		s.voteGranted[msg.voterId-1] = 1
	} else if s.currentTerm == msg.term && msg.voteGranted == false {
		//collect no votes
		s.voteGranted[msg.voterId-1] = -1
	}

	neededVotes := (quorumSize + 1) / 2
	// 1 for Yes votes, 0 for No vote , -1 if not yet received vote response from peer
	for i := 0; i < 4; i++ {
		if s.voteGranted[i] == 1 {
			yesVoteCount++
		}
		if s.voteGranted[i] == -1 {
			noVoteCount++
		}
	}
	//sever has received majority of yes votes
	if yesVoteCount >= neededVotes {
		s.state = "LEADER"
		s.leader = serverId
		//send heartbeat message
		lastIndex := -1
		lastTerm := -1
		if len(s.log) > 0 {
			lastIndex = len(s.log) - 1
			lastTerm = s.log[lastIndex].term
		}
		//set next index and match index
		s.nextIndex = make([]int, 4)
		s.matchIndex=make([]int, 4)
		for i:=0;i<4;i++{
			s.nextIndex[i]=len(s.log)
			s.matchIndex[i]=0
		}
		logEntries := make([]LogInfo, 0)
		for i := 0; i < len(s.peers); i++ {
			actionArray = append(actionArray, Send{from: s.peers[i], event: AppendEntriesReq{term: s.currentTerm, leaderId: s.leader, lastLogIndex: lastIndex, lastLogTerm: lastTerm, entries: logEntries, leaderCommit: s.commitIndex}})
		}
		//send alarm action to reset timer
		actionArray = append(actionArray, Alarm{10})

	} else if noVoteCount >= neededVotes {
		//stop timer and convert to follower state
		s.state = "FOLLOWER"
		actionArray = append(actionArray, Alarm{10})
	}
	return actionArray
}

func onTimeOutCandidate(s *Server) []Action {
	actionArray := make([]Action, 0)
	s.currentTerm = s.currentTerm + 1
	s.votedFor = serverId
	actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	//send vote request message to peers
	lastIndex := -1
	lastTerm := -1
	if len(s.log) > 0 {
		lastIndex = len(s.log) - 1
		lastTerm = s.log[lastIndex].term
	}

	for i := 0; i < len(s.peers); i++ {
		actionArray = append(actionArray, Send{from: s.peers[i], event: VoteReq{term: s.currentTerm, candidateId: serverId, lastLogIndex: lastIndex, lastLogTerm: lastTerm}})
	}

	actionArray = append(actionArray, Alarm{10})
	return actionArray

}

//Handle coming append entry request in candidate state
func OnAppendEntriesReqCandidate(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	logLength := len(s.log)
	var latestTerm int = 0
	if s.currentTerm < msg.term {
		//store latest term
		latestTerm = msg.term
	}
	if s.currentTerm > msg.term {
		// received request term is less then  server's current term then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if logLength < msg.lastLogIndex {
		//convert to follower mode on receiving append entry req from leader
		s.state = "FOLLOWER"
		s.leader = msg.leaderId
		//last log index of follower log not matches with received last log index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if msg.lastLogIndex != -1 && logLength != 0 && s.log[msg.lastLogIndex].term != msg.lastLogTerm {
		//convert to follower mode on receiving append entry req from leader
		s.leader = msg.leaderId
		s.state = "FOLLOWER"
		//last log index of follower log  matches with received last log index but terms differs on that index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else {
		//delete enteries from log starting from lastLogIndex+1 to end
		i := msg.lastLogIndex + 1
		if logLength > i {
			s.log = append(s.log[:i], s.log[:i+1]...)
		}
		if len(msg.entries) > 0 {
			//send log store action to raft node
			actionArray = append(actionArray, LogStore{index: i, data: msg.entries[msg.lastLogIndex+1].data})
			s.log = append(s.log, msg.entries...)
		}
		//update term and voted for if term in append entry rpc is greater than follower current term
		s.leader = msg.leaderId
		if latestTerm > s.currentTerm {
			s.currentTerm = msg.term
			s.votedFor = 0
			//send state store action to raft node
			actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
		}
		//change to follower
		s.state = "FOLLOWER"
		if msg.leaderCommit > s.commitIndex {
			s.commitIndex = min(msg.leaderCommit, i-1)
			//check data index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{index: s.commitIndex, data: s.log[s.commitIndex].data})
		}
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: true,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})

	}
	//In case candidate has rejected append entry request from the sever with higher term, then update the term and clear votedfor
	if latestTerm > s.currentTerm {
		s.currentTerm = latestTerm
		s.votedFor = 0
		//change to follower
		s.state = "FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	//return action
	return actionArray
}

//Handle append entry resp request coming to follower
func  OnAppendEntriesRespCandidate(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's term if term in msg is greater than it's current term
	if(s.currentTerm<msg.term){
		s.currentTerm=msg.term
		s.votedFor=0
		s.state="FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	return actionArray
}

//Handle coming append  request from client to candidate
func OnAppendRequestFromClientToCandidate(s *Server, msg Append) []Action {
	actionArray := make([]Action, 0)
	actionArray=append(actionArray,Commit{data:msg.data,error:"Leader is at Id "+strconv.Itoa(s.leader)})
	return actionArray
}

//**********************************Leader events**************************************************************
//**************************************************************************************************************
//Handle Coming Vote request to Leader
func OnVoteReqLeader(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	logIndex := len(s.log)
	logTerm := 0
	if logIndex > 0 {
		//get last term from logs
		logTerm = s.log[logIndex-1].term
	}
	var latestTerm int = 0
	//store latest term
	if s.currentTerm < msg.term {
		latestTerm = msg.term
	}
	if s.currentTerm >= msg.term {
		// received request term is less then or equal to server's current term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if logTerm > 0 && logTerm > msg.lastLogTerm {
		// received request last log term is less then server's last log term then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else if (logTerm == msg.lastLogTerm) && logIndex > 0 && logIndex-1 > msg.lastLogIndex {
		// received request last log term is equal to server's last log term but last log index differs then don't vote
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: false, voterId: serverId}})
	} else {
		//if all criteria satisfies then vote and update current term,voted for and convert to follower state
		s.currentTerm = msg.term
		s.votedFor = msg.candidateId
		//convert to follower state
		s.state = "FOLLOWER"
		//send the vote	response
		actionArray = append(actionArray, Send{from: msg.candidateId, event: VoteResp{term: s.currentTerm, voteGranted: true, voterId: serverId}})
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	if latestTerm > s.currentTerm {
		//In case servers has not granted his vote to the sever with higher term, then update the term and clear votedfor
		s.currentTerm = latestTerm
		s.votedFor = 0
		//convert to follower state
		s.state = "FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	//return action
	return actionArray
}

//Handle Coming Vote request to Leader
func OnVoteRespLeader(s *Server, msg VoteResp) []Action {
	//Leader will drop vote response message as he has already majority of vote responses in past that's why he is a leader now and it can't receive any vote resp msg from higher term server
	actionArray := make([]Action, 0)
	return actionArray
}

//Handle time out event in leader state
func onTimeOutLeader(s *Server) []Action {
	actionArray := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	lastIndex := -1
	lastTerm := -1
	if len(s.log) > 0 {
		//store last log index and term
		lastIndex = len(s.log) - 1
		lastTerm = s.log[lastIndex].term
	}
	//set next index and match index
	s.nextIndex = make([]int, 4)
	s.matchIndex=make([]int, 4)
	for i:=0;i<4;i++{
		s.nextIndex[i]=len(s.log)
		s.matchIndex[i]=0
	}

	//send heatbeat message to all peers
	for i := 0; i < len(s.peers); i++ {
		actionArray = append(actionArray, Send{from: s.peers[i], event: AppendEntriesReq{term: s.currentTerm, leaderId: s.leader, lastLogIndex: lastIndex, lastLogTerm: lastTerm, entries: logEntries, leaderCommit: s.commitIndex}})
	}

	actionArray = append(actionArray, Alarm{10})
	//return action
	return actionArray

}

//Handle coming append entry request in leader state
func OnAppendEntriesReqLeader(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	logLength := len(s.log)
	var latestTerm int = 0
	if s.currentTerm < msg.term {
		//store latest term
		latestTerm = msg.term
	}
	if s.currentTerm >= msg.term {
		// received request term is less then or equal server's current term then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if logLength < msg.lastLogIndex {
		//convert to follower mode on receiving append entry req from leader
		s.state = "FOLLOWER"
		s.leader = msg.leaderId
		//last log index of follower log not matches with received last log index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else if msg.lastLogIndex != -1 && logLength != 0 && s.log[msg.lastLogIndex].term != msg.lastLogTerm {
		//convert to follower mode on receiving append entry req from leader
		s.leader = msg.leaderId
		s.state = "FOLLOWER"
		//last log index of follower log  matches with received last log index but terms differs on that index then don't append
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: false,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})
	} else {
		//delete enteries from log starting from lastLogIndex+1 to end
		i := msg.lastLogIndex + 1
		if logLength > i {
			s.log = append(s.log[:i], s.log[:i+1]...)
		}
		if len(msg.entries) > 0 {
			//send log store action to raft node
			actionArray = append(actionArray, LogStore{index: i, data: msg.entries[msg.lastLogIndex+1].data})
			s.log = append(s.log, msg.entries...)
		}
		//update term and voted for if term in append entry rpc is greater than follower current term
		s.leader = msg.leaderId
		if latestTerm > s.currentTerm {
			s.currentTerm = msg.term
			s.votedFor = 0
			//send state store action to raft node
			actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
		}
		//change to follower
		s.state = "FOLLOWER"
		if msg.leaderCommit > s.commitIndex {
			s.commitIndex = min(msg.leaderCommit, i-1)
			//check data index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{index: s.commitIndex, data: s.log[s.commitIndex].data})
		}
		actionArray = append(actionArray, Send{from: msg.leaderId, event: AppendEntriesResp{term: s.currentTerm, success: true,from:serverId,count:len(msg.entries),lastLogIndex:msg.lastLogIndex}})

	}
	//In case server has rejected append entry request from the sever with higher term, then update the term and clear votedfor
	if latestTerm > s.currentTerm {
		s.currentTerm = latestTerm
		s.votedFor = 0
		//change to follower
		s.state = "FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}
	//return action
	return actionArray
}

//Handle coming append data request from client in leader state
func OnAppendRequestFromClientToLeader(s *Server, msg Append) []Action{	
	actionArray := make([]Action, 0)
	//leader will get his last log index
	lastIndex:=len(s.log)-1
	//Leader will append client data in his log first
	s.log=append(s.log,LogInfo{term:s.currentTerm,data:msg.data})
	//send logstore action to raft node
	actionArray = append(actionArray, LogStore{index: (lastIndex+1) , data: msg.data})
	//set next index and match index
	s.nextIndex = make([]int, 4)
	s.matchIndex=make([]int, 4)
	for i:=0;i<4;i++{
		s.nextIndex[i]=lastIndex+1
		s.matchIndex[i]=0
	}
	logEntries := make([]LogInfo, 0)
	logEntries=append(logEntries, LogInfo{term: s.currentTerm, data: msg.data})
	//send append entry request to all peers
	for i := 0; i < len(s.peers); i++ {
		actionArray = append(actionArray, Send{from: s.peers[i], event: AppendEntriesReq{term: s.currentTerm, leaderId: s.leader, lastLogIndex: lastIndex, lastLogTerm: s.log[lastIndex].term, entries: logEntries, leaderCommit: s.commitIndex}})
	}
	return actionArray
}

//Handle append entry resp request coming to leader
func  OnAppendEntriesRespLeader(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	id:=msg.from-1
	if(s.currentTerm<msg.term){
		// our state machine is lagging behind so convert to follower and update term
		s.currentTerm=msg.term
		s.votedFor=0
		s.state="FOLLOWER"
		//send state store action
		actionArray = append(actionArray, StateStore{term: s.currentTerm, votedFor: s.votedFor})
	}else{		
		if(msg.success==false){
			//decrease next index for the follower server id from which we have received this negative response			
			if(s.nextIndex[id]>0){
				s.nextIndex[id]=s.nextIndex[id]-1
				}else{
					s.nextIndex[id]=0
				}
			
			//update last log index and term needed to sent to the follower server
			lastIndex:=msg.lastLogIndex-1
			lastTerm:=0
			if(lastIndex>=0){
				lastTerm=s.log[lastIndex].term 
			}			
			//get slice of data from logs starting from last index to length
			entries:=s.log[lastIndex+1:len(s.log)]			
			//send append entry request again to that server
			actionArray = append(actionArray, Send{from: msg.from, event: AppendEntriesReq{term: s.currentTerm, leaderId: s.leader, lastLogIndex: lastIndex, lastLogTerm: lastTerm, entries: entries, leaderCommit: s.commitIndex}})

		}else{
			//update  match index if positive append entry response is received from follower for it's id
			if(msg.lastLogIndex+msg.count)>=s.matchIndex[id]{
				s.matchIndex[id]=msg.lastLogIndex+msg.count				
			}
		}
		// make a copy of matchindex in order to sort
		matchIndexCopy := make([]int, 4)		
		copy(matchIndexCopy, s.matchIndex)
		sort.IntSlice(matchIndexCopy).Sort()
		
		
		//commit index is that match index which is present on majority
		N:=matchIndexCopy[2]		
		//Leader will commit entries only if it's stored on a majority of servers and atleast one new entry from leader's term must also be stored on majority of servers. 
		if(N>s.commitIndex && s.log[N].term==s.currentTerm){
			//Update new commit index
			s.commitIndex=N
			//send commit action to client
			actionArray = append(actionArray,Commit{index:N,data:[]byte("Response")})
		}
		
	}
	return actionArray
}

func min(i int, j int) int {
	if i >= j {
		return i
	} else {
		return j
	}
}
