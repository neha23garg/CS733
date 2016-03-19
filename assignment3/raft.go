package main

import (
	"sort"
	"strconv"
)

//attributs of servers
type Server struct {
	CurrentTerm int
	VotedFor    int
	Log         []LogInfo
	CommitIndex int
	NextIndex   []int
	MatchIndex  []int
	Leader      int
	State       string
	Peers       []int
	VoteGranted []int
	MyId        int
}

// attributes of Log
type LogInfo struct {
	Term int
	Data []byte
}

//assign severid and number of servers
//var s.MyId int = 100
var quorumSize int = 5

// VoteRequest RPC
type VoteReq struct {
	//From 		 int
	Term         int // the Term
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// VoteResponse
type VoteResp struct {
	Term        int
	VoteGranted bool
	VoterId     int
}

//append Entries RPC
type AppendEntriesReq struct {
	Term         int
	LeaderId     int
	LastLogIndex int
	LastLogTerm  int
	Entries      []LogInfo
	LeaderCommit int
}

//append Entries response
type AppendEntriesResp struct {
	From         int
	Term         int
	Success      bool
	Count        int
	LastLogIndex int
}

//append request From client struct
type Append struct {
	Data []byte
}

type Timeout struct {
}

type Event interface{}

type Action interface{}

type Send struct {
	From int
	//Event can be vote req/resp and append rpc request/response
	Event Event
}

type Alarm struct {
	t int
}

type LogStore struct {
	Index int
	Data  []byte
}

type Commit struct {
	Index int
	Data  []byte
	Error string
}

type StateStore struct {
	Term     int
	VotedFor int
}

//Function to handle all incoming request to RAFT State Machine
func (s *Server) ProcessEvent(inputEvents Event) []Action {
	result := make([]Action, 0)
	//check State of the server
	if s.State == "FOLLOWER" {
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
		//Handle timeout Event
		case Timeout:
			result = onTimeOutFollower(s)
		//Handle append entry request made by client
		case Append:
			appendObj := inputEvents.(Append)
			result = OnAppendRequestFromClientToFollower(s, appendObj)
		}

	} else if s.State == "CANDIDATE" {

		switch inputEvents.(type) {
		//Handle Vote request coming to candidate
		case VoteReq:
			voteReqObj := inputEvents.(VoteReq)
			result = OnVoteReqCandidate(s, voteReqObj)
		//Handle Vote response coming to candidate
		case VoteResp:
			voteRespObj := inputEvents.(VoteResp)
			result = OnVoteRespCandidate(s, voteRespObj)
		//Handle timeout Event
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
			appendObj := inputEvents.(Append)
			result = OnAppendRequestFromClientToCandidate(s, appendObj)

		}
	} else if s.State == "LEADER" {

		switch inputEvents.(type) {
		//Handle Vote request coming to Leader
		case VoteReq:
			voteReqObj := inputEvents.(VoteReq)
			result = OnVoteReqLeader(s, voteReqObj)
		//Handle Vote response coming to Leader
		case VoteResp:
			voteRespObj := inputEvents.(VoteResp)
			result = OnVoteRespLeader(s, voteRespObj)
		//Handle timeout Event
		case Timeout:
			result = onTimeOutLeader(s)
		//Handle append entry request
		case AppendEntriesReq:
			AppendEntriesReqObj := inputEvents.(AppendEntriesReq)
			result = OnAppendEntriesReqLeader(s, AppendEntriesReqObj)
		//Handle append entry request made by client
		case Append:
			appendObj := inputEvents.(Append)
			result = OnAppendRequestFromClientToLeader(s, appendObj)
		//Handle Vote response coming to Leader
		case AppendEntriesResp:
			AppendEntriesRespObj := inputEvents.(AppendEntriesResp)
			result = OnAppendEntriesRespLeader(s, AppendEntriesRespObj)
		}
	}

	return result
}

//*********************FOLLOWER State EventS*************************************************************
//*******************************************************************************************************

//Handle Vote request coming to follower
func OnVoteReqFollower(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	LogIndex := len(s.Log)
	LogTerm := 0
	if LogIndex > 0 {
		//get last Term From Logs
		LogTerm = s.Log[LogIndex-1].Term
	}

	var latestTerm int = 0
	if s.CurrentTerm < msg.Term {
		//store latest Term
		latestTerm = msg.Term
	}
	if s.CurrentTerm > msg.Term {
		// received request Term is less then  server's current Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if LogTerm > 0 && LogTerm > msg.LastLogTerm {
		// received request last Log Term is less then server's last Log Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if (LogTerm == msg.LastLogTerm) && LogIndex > 0 && LogIndex-1 > msg.LastLogIndex {
		// received request last Log Term is equal to server's last Log Term but last Log Index differs then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if s.VotedFor == 0 || s.VotedFor == msg.CandidateId {
		//if all criteria satisfies and follower has not voted for this Term or voted for the same candidate id then vote and update current Term,voted for
		s.CurrentTerm = msg.Term
		s.VotedFor = msg.CandidateId
		//send the vote	response
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: true, VoterId: s.MyId}})
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	} else {
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	}
	if latestTerm > s.CurrentTerm {
		//In case servers has not granted his vote to the sever with higher Term, then update the Term and clear VotedFor
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	//return action
	return actionArray
}

//Handle vote resp request coming to follower
func OnVoteRespFollower(s *Server, msg VoteResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's Term if Term in msg is greater than it's current Term
	if s.CurrentTerm < msg.Term {
		s.CurrentTerm = msg.Term
		s.VotedFor = 0
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	return actionArray
}

//Handle Append entry request coming to follower
func OnAppendEntriesReqFollower(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	LogLength := len(s.Log)
	var latestTerm int = 0
	if s.CurrentTerm < msg.Term {
		//store latest Term
		latestTerm = msg.Term
	}
	if s.CurrentTerm > msg.Term {
		// received request Term is less then  server's current Term then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if LogLength < msg.LastLogIndex {
		//last Log Index of follower Log not matches with received last Log Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if msg.LastLogIndex != -1 && LogLength != 0 && s.Log[msg.LastLogIndex].Term != msg.LastLogTerm {
		//last Log Index of follower Log  matches with received last Log Index but Terms differs on that Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else {
		//delete enteries From Log starting From LastLogIndex+1 to end
		if len(msg.Entries) > 0 {
			i := msg.LastLogIndex + 1
			//check Data
			if LogLength > i {
				s.Log = append(s.Log[:i], s.Log[:i+1]...)
			}
			if len(msg.Entries) > 0 {
				//send Log store action to raft node
				//fmt.Println("appendlastindex",msg.LastLogIndex,msg.Entries)
				actionArray = append(actionArray, LogStore{Index: i, Data: msg.Entries[0].Data})
				s.Log = append(s.Log, msg.Entries...)
			}
		}
		//set Leader id
		s.Leader = msg.LeaderId
		//update Term and voted for if Term in append entry rpc is greater than follower current Term
		if latestTerm > s.CurrentTerm {
			s.CurrentTerm = latestTerm
			s.VotedFor = 0
			//send State store action to raft node
			actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
		}
		if msg.LeaderCommit > s.CommitIndex {
			i := msg.LastLogIndex + 1
			if len(msg.Entries) > 0 {
				s.CommitIndex = min(msg.LeaderCommit, i-1)
			} else {
				s.CommitIndex = msg.LeaderCommit
			}

			//check Data Index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{Index: s.CommitIndex, Data: s.Log[s.CommitIndex].Data})
		}
		//send append entry action for the received request
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: true, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})

	}
	if latestTerm > s.CurrentTerm {
		//In case follower has rejected append entry request From the sever with higher Term, then update the Term and clear VotedFor
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	//return action
	return actionArray
}

//Handle append entry resp request coming to follower
func OnAppendEntriesRespFollower(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's Term if Term in msg is greater than it's current Term
	if s.CurrentTerm < msg.Term {
		s.CurrentTerm = msg.Term
		s.VotedFor = 0
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	return actionArray
}

//Handle follower Timeout
func onTimeOutFollower(s *Server) []Action {
	actionArray := make([]Action, 0)
	//convert to candidate
	s.State = "CANDIDATE"
	//increase current Term by 1
	s.CurrentTerm = s.CurrentTerm + 1
	//vote for self
	s.VotedFor = s.MyId
	//clear vote array
	s.VoteGranted = make([]int, 10)
	//send State store action
	actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	//send vote request message to Peers
	lastIndex := -1
	lastTerm := -1
	if len(s.Log) > 0 {
		//get last Log Index and Term
		lastIndex = len(s.Log) - 1
		lastTerm = s.Log[lastIndex].Term
	}

	for i := 0; i < len(s.Peers); i++ {
		//vote request action
		actionArray = append(actionArray, Send{From: s.Peers[i], Event: VoteReq{Term: s.CurrentTerm, CandidateId: s.MyId, LastLogIndex: lastIndex, LastLogTerm: lastTerm}})
	}
	//reset timer
	actionArray = append(actionArray, Alarm{10})
	return actionArray

}

//Handle coming append  request From client to follower
func OnAppendRequestFromClientToFollower(s *Server, msg Append) []Action {
	actionArray := make([]Action, 0)
	actionArray = append(actionArray, Commit{Data: msg.Data, Error: "Leader is at Id " + strconv.Itoa(s.Leader)})
	return actionArray
}

//********************************************Candidate Events***********************************************************
//*************************************************************************************************************************
//Handle Coming Vote request to Candidate
func OnVoteReqCandidate(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	LogIndex := len(s.Log)
	LogTerm := 0
	if LogIndex > 0 {
		//get last Term From Logs
		LogTerm = s.Log[LogIndex-1].Term
	}
	var latestTerm int = 0
	if s.CurrentTerm < msg.Term {
		//store latest Term
		latestTerm = msg.Term
	}

	if s.CurrentTerm >= msg.Term {
		// received request Term is less then or equal server's current Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if LogTerm > 0 && LogTerm > msg.LastLogTerm {
		// received request last Log Term is less then server's last Log Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if (LogTerm == msg.LastLogTerm) && LogIndex > 0 && LogIndex-1 > msg.LastLogIndex {
		// received request last Log Term is equal to server's last Log Term but last Log Index differs then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else {
		//if all criteria satisfies and follower has not voted for this higher Term  then vote and update current Term,voted for
		s.CurrentTerm = msg.Term
		s.VotedFor = msg.CandidateId
		//convert to follower State
		s.State = "FOLLOWER"
		//send the vote	response
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: true, VoterId: s.MyId}})
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	if latestTerm > s.CurrentTerm {
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//convert to follower State
		s.State = "FOLLOWER"
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	return actionArray
}

//Handle Vote resp Events in candidate State
func OnVoteRespCandidate(s *Server, msg VoteResp) []Action {
	actionArray := make([]Action, 0)
	//As candidate has voted for himself
	yesVoteCount := 0
	noVoteCount := 0
	s.VoteGranted[s.MyId] = 1

	if s.CurrentTerm < msg.Term {
		//If msg Term is higher than current Term convert to follower
		s.CurrentTerm = msg.Term
		s.VotedFor = 0
		s.State = "FOLLOWER"
		s.VoteGranted = make([]int, 5)
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
		//reset timer
		actionArray = append(actionArray, Alarm{10})
	} else if s.CurrentTerm == msg.Term && msg.VoteGranted == true {
		//collect yes votes
		s.VoteGranted[msg.VoterId-1] = 1
	} else if s.CurrentTerm == msg.Term && msg.VoteGranted == false {
		//collect no votes
		s.VoteGranted[msg.VoterId-1] = -1
	}

	neededVotes := (quorumSize + 1) / 2
	// 1 for Yes votes, 0 for No vote , -1 if not yet received vote response From peer
	for i := 0; i < 5; i++ {
		if s.VoteGranted[i] == 1 {
			yesVoteCount++
		}
		if s.VoteGranted[i] == -1 {
			noVoteCount++
		}
	}
	//sever has received majority of yes votes
	if yesVoteCount >= neededVotes {
		s.State = "LEADER"
		s.Leader = s.MyId
		//send heartbeat message
		lastIndex := -1
		lastTerm := -1
		if len(s.Log) > 0 {
			lastIndex = len(s.Log) - 1
			lastTerm = s.Log[lastIndex].Term
		}
		//set next Index and match Index
		s.NextIndex = make([]int, 10)
		s.MatchIndex = make([]int, 10)
		for i := 0; i < 4; i++ {
			s.NextIndex[i] = len(s.Log)
			s.MatchIndex[i] = 0
		}
		LogEntries := make([]LogInfo, 0)
		for i := 0; i < len(s.Peers); i++ {
			actionArray = append(actionArray, Send{From: s.Peers[i], Event: AppendEntriesReq{Term: s.CurrentTerm, LeaderId: s.Leader, LastLogIndex: lastIndex, LastLogTerm: lastTerm, Entries: LogEntries, LeaderCommit: s.CommitIndex}})
		}
		//send alarm action to reset timer
		actionArray = append(actionArray, Alarm{10})

	} else if noVoteCount >= neededVotes {
		//stop timer and convert to follower State
		s.State = "FOLLOWER"
		actionArray = append(actionArray, Alarm{10})
	}
	return actionArray
}

func onTimeOutCandidate(s *Server) []Action {
	actionArray := make([]Action, 0)
	s.CurrentTerm = s.CurrentTerm + 1
	s.VotedFor = s.MyId
	actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	//send vote request message to Peers
	lastIndex := -1
	lastTerm := -1
	if len(s.Log) > 0 {
		lastIndex = len(s.Log) - 1
		lastTerm = s.Log[lastIndex].Term
	}

	for i := 0; i < len(s.Peers); i++ {
		actionArray = append(actionArray, Send{From: s.Peers[i], Event: VoteReq{Term: s.CurrentTerm, CandidateId: s.MyId, LastLogIndex: lastIndex, LastLogTerm: lastTerm}})
	}

	actionArray = append(actionArray, Alarm{10})
	return actionArray

}

//Handle coming append entry request in candidate State
func OnAppendEntriesReqCandidate(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	LogLength := len(s.Log)
	var latestTerm int = 0
	if s.CurrentTerm < msg.Term {
		//store latest Term
		latestTerm = msg.Term
	}
	if s.CurrentTerm > msg.Term {
		// received request Term is less then  server's current Term then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if LogLength < msg.LastLogIndex {
		//convert to follower mode on receiving append entry req From Leader
		s.State = "FOLLOWER"
		s.Leader = msg.LeaderId
		//last Log Index of follower Log not matches with received last Log Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if msg.LastLogIndex != -1 && LogLength != 0 && s.Log[msg.LastLogIndex].Term != msg.LastLogTerm {
		//convert to follower mode on receiving append entry req From Leader
		s.Leader = msg.LeaderId
		s.State = "FOLLOWER"
		//last Log Index of follower Log  matches with received last Log Index but Terms differs on that Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else {
		//delete enteries From Log starting From LastLogIndex+1 to end
		i := msg.LastLogIndex + 1
		if LogLength > i {
			s.Log = append(s.Log[:i], s.Log[:i+1]...)
		}
		if len(msg.Entries) > 0 {
			//send Log store action to raft node
			actionArray = append(actionArray, LogStore{Index: i, Data: msg.Entries[0].Data})
			s.Log = append(s.Log, msg.Entries...)
		}
		//update Term and voted for if Term in append entry rpc is greater than follower current Term
		s.Leader = msg.LeaderId
		if latestTerm > s.CurrentTerm {
			s.CurrentTerm = msg.Term
			s.VotedFor = 0
			//send State store action to raft node
			actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
		}
		//change to follower
		s.State = "FOLLOWER"
		if msg.LeaderCommit > s.CommitIndex {
			s.CommitIndex = min(msg.LeaderCommit, i-1)
			//check Data Index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{Index: s.CommitIndex, Data: s.Log[s.CommitIndex].Data})
		}
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: true, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})

	}
	//In case candidate has rejected append entry request From the sever with higher Term, then update the Term and clear VotedFor
	if latestTerm > s.CurrentTerm {
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//change to follower
		s.State = "FOLLOWER"
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	//return action
	return actionArray
}

//Handle append entry resp request coming to follower
func OnAppendEntriesRespCandidate(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	//Follower will drop the response but update it's Term if Term in msg is greater than it's current Term
	if s.CurrentTerm < msg.Term {
		s.CurrentTerm = msg.Term
		s.VotedFor = 0
		s.State = "FOLLOWER"
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	return actionArray
}

//Handle coming append  request From client to candidate
func OnAppendRequestFromClientToCandidate(s *Server, msg Append) []Action {
	actionArray := make([]Action, 0)
	actionArray = append(actionArray, Commit{Data: msg.Data, Error: "Leader is at Id " + strconv.Itoa(s.Leader)})
	return actionArray
}

//**********************************Leader Events**************************************************************
//**************************************************************************************************************
//Handle Coming Vote request to Leader
func OnVoteReqLeader(s *Server, msg VoteReq) []Action {
	actionArray := make([]Action, 0)
	LogIndex := len(s.Log)
	LogTerm := 0
	if LogIndex > 0 {
		//get last Term From Logs
		LogTerm = s.Log[LogIndex-1].Term
	}
	var latestTerm int = 0
	//store latest Term
	if s.CurrentTerm < msg.Term {
		latestTerm = msg.Term
	}
	if s.CurrentTerm >= msg.Term {
		// received request Term is less then or equal to server's current Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if LogTerm > 0 && LogTerm > msg.LastLogTerm {
		// received request last Log Term is less then server's last Log Term then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else if (LogTerm == msg.LastLogTerm) && LogIndex > 0 && LogIndex-1 > msg.LastLogIndex {
		// received request last Log Term is equal to server's last Log Term but last Log Index differs then don't vote
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: false, VoterId: s.MyId}})
	} else {
		//if all criteria satisfies then vote and update current Term,voted for and convert to follower State
		s.CurrentTerm = msg.Term
		s.VotedFor = msg.CandidateId
		//convert to follower State
		s.State = "FOLLOWER"
		//send the vote	response
		actionArray = append(actionArray, Send{From: msg.CandidateId, Event: VoteResp{Term: s.CurrentTerm, VoteGranted: true, VoterId: s.MyId}})
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	if latestTerm > s.CurrentTerm {
		//In case servers has not granted his vote to the sever with higher Term, then update the Term and clear VotedFor
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//convert to follower State
		s.State = "FOLLOWER"
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	//return action
	return actionArray
}

//Handle Coming Vote request to Leader
func OnVoteRespLeader(s *Server, msg VoteResp) []Action {
	//Leader will drop vote response message as he has already majority of vote responses in past that's why he is a Leader now and it can't receive any vote resp msg From higher Term server
	actionArray := make([]Action, 0)
	return actionArray
}

//Handle time out Event in Leader State
func onTimeOutLeader(s *Server) []Action {
	//fmt.Println("onTimeOutLeader***")
	actionArray := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	lastIndex := -1
	lastTerm := -1
	if len(s.Log) > 0 {
		//store last Log Index and Term
		lastIndex = len(s.Log) - 1
		lastTerm = s.Log[lastIndex].Term
	}
	//set next Index and match Index
	s.NextIndex = make([]int, 10)
	s.MatchIndex = make([]int, 10)
	for i := 0; i < 4; i++ {
		s.NextIndex[i] = len(s.Log)
		s.MatchIndex[i] = 0
	}

	//send heatbeat message to all Peers
	for i := 0; i < len(s.Peers); i++ {
		//fmt.Println("leader timeout************")
		actionArray = append(actionArray, Send{From: s.Peers[i], Event: AppendEntriesReq{Term: s.CurrentTerm, LeaderId: s.Leader, LastLogIndex: lastIndex, LastLogTerm: lastTerm, Entries: LogEntries, LeaderCommit: s.CommitIndex}})
	}
	//fmt.Println("leader alarm************")
	actionArray = append(actionArray, Alarm{10})
	//return action
	return actionArray

}

//Handle coming append entry request in Leader State
func OnAppendEntriesReqLeader(s *Server, msg AppendEntriesReq) []Action {
	actionArray := make([]Action, 0)
	LogLength := len(s.Log)
	var latestTerm int = 0
	if s.CurrentTerm < msg.Term {
		//store latest Term
		latestTerm = msg.Term
	}
	if s.CurrentTerm >= msg.Term {
		// received request Term is less then or equal server's current Term then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if LogLength < msg.LastLogIndex {
		//convert to follower mode on receiving append entry req From Leader
		s.State = "FOLLOWER"
		s.Leader = msg.LeaderId
		//last Log Index of follower Log not matches with received last Log Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else if msg.LastLogIndex != -1 && LogLength != 0 && s.Log[msg.LastLogIndex].Term != msg.LastLogTerm {
		//convert to follower mode on receiving append entry req From Leader
		s.Leader = msg.LeaderId
		s.State = "FOLLOWER"
		//last Log Index of follower Log  matches with received last Log Index but Terms differs on that Index then don't append
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: false, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})
	} else {
		//delete enteries From Log starting From LastLogIndex+1 to end
		i := msg.LastLogIndex + 1
		if LogLength > i {
			s.Log = append(s.Log[:i], s.Log[:i+1]...)
		}
		if len(msg.Entries) > 0 {
			//send Log store action to raft node
			actionArray = append(actionArray, LogStore{Index: i, Data: msg.Entries[0].Data})
			s.Log = append(s.Log, msg.Entries...)
		}
		//update Term and voted for if Term in append entry rpc is greater than follower current Term
		s.Leader = msg.LeaderId
		if latestTerm > s.CurrentTerm {
			s.CurrentTerm = msg.Term
			s.VotedFor = 0
			//send State store action to raft node
			actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
		}
		//change to follower
		s.State = "FOLLOWER"
		if msg.LeaderCommit > s.CommitIndex {
			s.CommitIndex = min(msg.LeaderCommit, i-1)
			//check Data Index******
			//send commit action to raft node
			actionArray = append(actionArray, Commit{Index: s.CommitIndex, Data: s.Log[s.CommitIndex].Data})
		}
		actionArray = append(actionArray, Send{From: msg.LeaderId, Event: AppendEntriesResp{Term: s.CurrentTerm, Success: true, From: s.MyId, Count: len(msg.Entries), LastLogIndex: msg.LastLogIndex}})

	}
	//In case server has rejected append entry request From the sever with higher Term, then update the Term and clear VotedFor
	if latestTerm > s.CurrentTerm {
		s.CurrentTerm = latestTerm
		s.VotedFor = 0
		//change to follower
		s.State = "FOLLOWER"
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	}
	//return action
	return actionArray
}

//Handle coming append Data request From client in Leader State
func OnAppendRequestFromClientToLeader(s *Server, msg Append) []Action {
	actionArray := make([]Action, 0)
	//Leader will get his last Log Index
	lastIndex := len(s.Log) - 1
	lastTerm := -1
	if lastIndex >= 0 {
		lastTerm = s.Log[lastIndex].Term
	}
	//Leader will append client Data in his Log first
	s.Log = append(s.Log, LogInfo{Term: s.CurrentTerm, Data: msg.Data})
	//send Logstore action to raft node
	actionArray = append(actionArray, LogStore{Index: (lastIndex + 1), Data: msg.Data})
	//set next Index and match Index
	s.NextIndex = make([]int, 10)
	s.MatchIndex = make([]int, 10)
	for i := 0; i < 4; i++ {
		s.NextIndex[i] = lastIndex + 1
		s.MatchIndex[i] = 0
	}
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: s.CurrentTerm, Data: msg.Data})
	//send append entry request to all Peers
	for i := 0; i < len(s.Peers); i++ {
		actionArray = append(actionArray, Send{From: s.Peers[i], Event: AppendEntriesReq{Term: s.CurrentTerm, LeaderId: s.Leader, LastLogIndex: lastIndex, LastLogTerm: lastTerm, Entries: LogEntries, LeaderCommit: s.CommitIndex}})
	}
	return actionArray
}

//Handle append entry resp request coming to Leader
func OnAppendEntriesRespLeader(s *Server, msg AppendEntriesResp) []Action {
	actionArray := make([]Action, 0)
	id := msg.From - 1
	if s.CurrentTerm < msg.Term {
		// our State machine is lagging behind so convert to follower and update Term
		s.CurrentTerm = msg.Term
		s.VotedFor = 0
		s.State = "FOLLOWER"
		//send State store action
		actionArray = append(actionArray, StateStore{Term: s.CurrentTerm, VotedFor: s.VotedFor})
	} else {
		if msg.Success == false {
			//decrease next Index for the follower server id From which we have received this negative response
			if s.NextIndex[id] > 0 {
				s.NextIndex[id] = s.NextIndex[id] - 1
			} else {
				s.NextIndex[id] = 0
			}

			//update last Log Index and Term needed to sent to the follower server
			lastIndex := -1
			if msg.LastLogIndex >= 0 {
				lastIndex = msg.LastLogIndex - 1
			}
			lastTerm := 0
			if lastIndex >= 0 {
				lastTerm = s.Log[lastIndex].Term
			}
			//get slice of Data From Logs starting From last Index to length
			Entries := make([]LogInfo, 0)
			//fmt.Println("slice-----",s.Log,lastIndex,len(s.Log))
			if len(s.Log) > 0 {
				Entries = s.Log[lastIndex+1 : len(s.Log)]
			}

			//send append entry request again to that server
			actionArray = append(actionArray, Send{From: msg.From, Event: AppendEntriesReq{Term: s.CurrentTerm, LeaderId: s.Leader, LastLogIndex: lastIndex, LastLogTerm: lastTerm, Entries: Entries, LeaderCommit: s.CommitIndex}})

		} else {
			//update  match Index if positive append entry response is received From follower for it's id
			//fmt.Println("error",id,len(s.MatchIndex))
			if (msg.LastLogIndex + msg.Count) >= s.MatchIndex[id] {
				s.MatchIndex[id] = msg.LastLogIndex + msg.Count
				//fmt.Println("matchIndex",s.MatchIndex[id])
			}
		}
		// make a copy of MatchIndex in order to sort
		MatchIndexCopy := make([]int, 5)
		copy(MatchIndexCopy, s.MatchIndex)
		sort.IntSlice(MatchIndexCopy).Sort()
		//fmt.Println(MatchIndexCopy)
		//fmt.Println(s.MatchIndex)

		//commit Index is that match Index which is present on majority
		N := MatchIndexCopy[3]
		//Leader will commit Entries only if it's stored on a majority of servers and atleast one new entry From Leader's Term must also be stored on majority of servers.
		//fmt.Println("memory logs",s.Log)
		if len(s.Log) >= 1 {
			//fmt.Println("log +term",s.Log[N].Term,s.CurrentTerm,N,s.CommitIndex)
			if N > s.CommitIndex && s.Log[N].Term == s.CurrentTerm {
				//Update new commit Index
				s.CommitIndex = N
				//send commit action to client
				actionArray = append(actionArray, Commit{Index: N, Data: []byte(s.Log[N].Data)})
				/*//send heartbeat msg to all the servers to commit the entry
				blankEntries:=make([]LogInfo,0)
				lastIndex:=-1
				if(msg.LastLogIndex>=0){
						lastIndex = msg.LastLogIndex - 1
				}
				lastTerm := 0
				if lastIndex >= 0 {
					lastTerm = s.Log[lastIndex].Term
				}
				actionArray = append(actionArray, Send{From: msg.From, Event: AppendEntriesReq{Term: s.CurrentTerm, LeaderId: s.Leader, LastLogIndex: lastIndex, LastLogTerm: lastTerm, Entries: blankEntries, LeaderCommit: s.CommitIndex}})*/
			}
		}

	}
	return actionArray
}

//to find minimum of two numbers
func min(i int, j int) int {
	if i >= j {
		return i
	} else {
		return j
	}
}
