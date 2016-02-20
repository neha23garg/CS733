package main

import (
	"fmt"
	"reflect"
	"testing"
)

//*****Follower Testing**************************************

//Test for success response of vote request received from a higher term candidate
func TestPositiveVoteRequestFollower(t *testing.T) {
	//Test for positive response of vote request
	s := &Server{currentTerm: 0, votedFor: 0, leader: 0, state: "FOLLOWER"}
	action := make([]Action, 0)
	// term:1,candidate id:1
	action = s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 1)
	expectInt(t, s.votedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 1)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 1)
	expectInt(t, vote, 1)
}

//Test for negative response of vote request if vote request is received from lower term candidate
func TestNegativeVoteRequestWithLowTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 0, state: "FOLLOWER"}
	// term:1,candidate id:1
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 0)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test duplicate vote Request from same server id and term for which follower has already voted
func TestDuplicateVoteRequestFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, leader: 0, state: "FOLLOWER"}
	// term:1,candidate id:1
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test vote Request if follower has already granted his vote for the term and receive vote request for the same term from another server
func TestVoteRequestForSameTermTwiceFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, leader: 0, state: "FOLLOWER"}
	// term:2,candidate id:3
	action := s.ProcessEvent(VoteReq{2, 3, 1, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 3)
}

//Test for negative response of vote request from a server with higher term with not upto date logs
func TestNegativeVoteRequestWithHigherTermFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: 0, leader: 0, state: "FOLLOWER"}
	// term:5,candidate id:1,lastlogIndex:0,lastLogTerm:1
	action := s.ProcessEvent(VoteReq{5, 1, 0, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.votedFor, 0)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)

}

//Test for negative response of vote request from a server with higher term and it's previous log index not matching with follower previous log index
func TestNegativeVoteRequestWithLowerIndexFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, leader: 0, state: "FOLLOWER"}
	// term:3,candidate id:1,lastlogIndex:0,lastLogTerm:1
	action := s.ProcessEvent(VoteReq{3, 1, 0, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for negative response of vote request from a server with higher term and it's previous log index matches with follower previous log index but previous log term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: 0, leader: 0, state: "FOLLOWER"}
	// term:3,candidate id:1,lastlogIndex:1,lastLogTerm:1
	action := s.ProcessEvent(VoteReq{3, 1, 1, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test vote response request with same term
func TestVoteRespFromSameTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	//term:2,voteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{2, false, 3})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test vote response request with lower term
func TestVoteRespFromLowerTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	//term:1,voteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{1, false, 3})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test vote response request with higher term
func TestVoteRespFromHigherTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	//term:3,voteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{3, false, 3})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for successfull append entry request
func TestSuccessAppendEntryRequestFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	s := &Server{currentTerm: 1, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, -1, -1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 1)
	var ind int
	var to int
	act, ev, ind, flag, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, _ = getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to = getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 1)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for successfull append entry request from higher term server
func TestSuccessAppendEntryRequestFromHigherTermFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 3, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.leader, 1)
	var ind int
	// var to int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 2)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 3)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 3)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever with higher term but with not matching previous log index and term
func TestAppendEntryRequestWithHigherTermHeartbeatFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "FOLLOWER", peers: peerArray}
	//logEntries=append(logEntries,LogInfo{term:1,data:[]byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 0)
	var ind int
	// var to int
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 5)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for successfull append entry request from same term leader but logs entries are having data
func TestSuccessAppendEntryRequestFromSameTermFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 2, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 1)
	var ind int
	// var to int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for successfull append entry request from same term leader but logs entries are having more data beyond last log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")}, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 2, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 1)
	var ind int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever with same term but with not matching previous log index
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevIndexFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 3, 2, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever with higher term but with  matching previous log index but not term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "FOLLOWER", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	expectInt(t, s.leader, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test append entry response request with same term from serverid 3
func TestAppendEntryRespFromSameTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 2, false, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with lower term from serverid 3
func TestAppendEntryRespFromLowerTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 1, false, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with higher term from serverid 3
func TestAppendEntryRespFromHigherTermFollower(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 3, false, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for timeout event for follower
func TestTimeOutFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 1, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "FOLLOWER", peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	expectInt(t, vote, 5)
	var ev string
	var ind int
	var to int
	//check whether vote request has sent to 4 other servers from the sever id 5
	for i := 1; i <= 4; i++ {
		act, ev, ind, _, vote, to = getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "VoteReq")
		expectInt(t, ind, 2)
		expectInt(t, vote, 5)
		//check candidate id as well
		expectInt(t, to, i)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ = getDetails(action, 5)
	expectString(t, act, "Alarm")
	//check time stamp**********

}

func TestAppendRequestFromClientToFollower(t *testing.T) {
	s := &Server{currentTerm: 1, votedFor: 1, leader: 1, state: "FOLLOWER"}
	action := s.ProcessEvent(Append{[]byte("hello")})
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	typeaction := action[0]
	obj := typeaction.(Commit)
	expectString(t, obj.error, "Leader is at Id 1")

}

//***********************Candidate Testing****************************************************************
func TestPositiveVoteRequestCandidate(t *testing.T) {
	//Test for positive response of vote request
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{currentTerm: 1, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := make([]Action, 0)
	action = s.ProcessEvent(VoteReq{2, 1, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	expectInt(t, vote, 1)
}

//Test for negative response of vote request
func TestNegativeVoteRequestWithLowerTermCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with same term and it's previous log index not matching with follower previous log index
func TestNegativeVoteRequestWithLowerIndexCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 0, 1})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for negative response of vote request from a server with higher term and it's previous log index matches with other candidate's previous log index but previous log term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with higher term and it's previous log index matches with other candidate's previous log index but previous log term differs
func TestNegativeVoteRequestFromHigherTermCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := s.ProcessEvent(VoteReq{5, 1, 1, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.votedFor, 0)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
}

//Test for candidate receiving vote response and become leader when you receive majority of votes
func TestVoteResponseCandidateToLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 4)
	//one yes vote already present in array
	voteArray[0] = 1
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray, voteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{2, true, 2})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	for i := 0; i <= 3; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		expectInt(t, ind, 2)
		expectInt(t, vote, 5)
		//check candidate id as well
		expectInt(t, to, i+1)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ := getDetails(action, 4)
	expectString(t, act, "Alarm")
}

//Test for candidate receiving vote response and remain in candidate state untill you receive majority of votes
func TestVoteResponseRemainCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 4)
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray, voteGranted: voteArray}
	_ = s.ProcessEvent(VoteResp{2, true, 2})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)

}

//Test for candidate receiving vote response and convert to follower if you receive majority of no votes
func TestVoteResponseCandidateToFollower(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 4)
	//two no vote already present in array
	voteArray[0] = -1
	voteArray[1] = -1
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray, voteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{2, false, 3})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Alarm")

}

//Test for candidate receiving vote response and convert to follower if you receive vote response from higher term server
func TestVoteResponseCandidateToFollowerAndUpdateTerm(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 4)
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray, voteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{3, false, 3})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, _, _, _, _, _ = getDetails(action, 1)
	expectString(t, act, "Alarm")

}

//Test for candidate receiving vote response and remain in candidate state if it receives vote response from lower term
func TestVoteResponseRemainInCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 4)
	//already has one yes vote
	voteArray[0] = 1
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "CANDIDATE", peers: peerArray, voteGranted: voteArray}
	_ = s.ProcessEvent(VoteResp{1, true, 2})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)

}

//Test for timeout event for candidate
func TestTimeOutCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 1, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	expectInt(t, vote, 5)
	var ev string
	var ind int
	var to int
	//check whether vote request has sent to 4 other servers from the sever id 5
	for i := 1; i <= 4; i++ {
		act, ev, ind, _, vote, to = getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "VoteReq")
		expectInt(t, ind, 2)
		expectInt(t, vote, 5)
		//check candidate id as well
		expectInt(t, to, i)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ = getDetails(action, 5)
	expectString(t, act, "Alarm")
	//check time stamp**********

}

//Test for successfull append entry request from higher term server with upto date logs
func TestSuccessAppendEntryRequestCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 3, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	var ind int
	// var to int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 2)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 3)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 3)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for successfull append entry request from same term server but more upto date logs
func TestSuccessAppendEntryRequestFromUptoDateLogsCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 2, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	act, ev, ind, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for append entry request from a sever with higher term but with not matching previous log index and term
func TestAppendEntryRequestWithHigherTermHeartbeatCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "CANDIDATE", peers: peerArray}
	//logEntries=append(logEntries,LogInfo{term:1,data:[]byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 0)
	var ind int
	// var to int
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 5)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for successfull append entry request from same term leader but logs entries are having more data beyond last log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")}, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "CANDIDATE", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 2, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 1)
	var ind int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever with higher term but with  matching previous log index but not term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "CANDIDATE", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	expectInt(t, s.leader, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for append entry request from a sever with lower term term
func TestFailAppendEntryRequestFromLowerTermCandidate(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "CANDIDATE", peers: peerArray}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test append entry response request with same term from serverid 3
func TestAppendEntryRespFromSameTermCandidate(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 2, false, 0, 0})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with lower term from serverid 3
func TestAppendEntryRespFromLowerTermCandidate(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 1, false, 0, 0})
	expectString(t, s.state, "CANDIDATE")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with higher term from serverid 3
func TestAppendEntryRespFromHigherTermCandidate(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 1, state: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 3, false, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

func TestAppendRequestFromClientToCandidate(t *testing.T) {
	s := &Server{currentTerm: 1, votedFor: 1, leader: 1, state: "CANDIDATE"}
	action := s.ProcessEvent(Append{[]byte("hello")})
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	typeaction := action[0]
	obj := typeaction.(Commit)
	expectString(t, obj.error, "Leader is at Id 1")

}

//**************Leader Testing*************************************************************
func TestPositiveVoteRequestLeader(t *testing.T) {
	//Test for positive response of vote request
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{currentTerm: 1, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "LEADER", peers: peerArray}
	action := make([]Action, 0)
	action = s.ProcessEvent(VoteReq{2, 1, 0, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	expectInt(t, vote, 1)
}

//Test for negative response of vote request
func TestNegativeVoteRequestWithLowerTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with same term and it's previous log index not matching with follower previous log index
func TestNegativeVoteRequestWithLowerIndexLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 0, 1})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for negative response of vote request from a server with higher term and it's previous log index matches with other candidate's previous log index but previous log term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with higher term and it's previous log index matches with other candidate's previous log index but previous log term differs
func TestNegativeVoteRequestFromHigherTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(VoteReq{5, 1, 1, 1})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.votedFor, 0)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
}

//Test for vote resp request receieved by leader
func TestVoteRespLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 0, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(VoteResp{2, false, 1})
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)

}

//Test for leader timeout
func TestTimeOutLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	//peer id array
	peerArray = append(peerArray, 1, 2, 3, 4)
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 5, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	for i := 0; i <= 3; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		//check term
		expectInt(t, ind, 2)
		//check voted for
		expectInt(t, vote, 5)
		//check candidate id as well
		expectInt(t, to, i+1)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ := getDetails(action, 4)
	expectString(t, act, "Alarm")

}

//Test for successfull append entry request from higher term server with uptodate logs
func TestSuccessAppendEntryRequestLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: -1, leader: 0, state: "LEADER"}
	logEntries = append(logEntries, LogInfo{term: 3, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	var ind int
	// var to int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 2)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 3)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 3)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for append entry request from a sever with higher term but with not matching previous log index and term
func TestAppendEntryRequestWithHigherTermHeartbeatLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "LEADER"}
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 0)
	var ind int
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 1)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 2)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 5)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for successfull append entry request from higher term leader but logs entries are having more data beyond last log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("hello")}, LogInfo{term: 1, data: []byte("hello")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 0, state: "LEADER"}
	logEntries = append(logEntries, LogInfo{term: 2, data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.leader, 1)
	expectInt(t, s.votedFor, 0)
	var ind int
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, ev, ind, flag, _, _ := getDetails(action, 2)
	expectString(t, act, "Commit")
	expectInt(t, ind, 0)
	act, ev, ind, flag, _, to := getDetails(action, 3)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 3)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever from a  higher term server but with  matching previous log index but not term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "LEADER"}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 3)
	expectInt(t, s.votedFor, 0)
	expectInt(t, s.leader, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
	act, _, trm, _, vote, _ := getDetails(action, 1)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for append entry request from a sever with lower term term
func TestFailAppendEntryRequestFromLowerTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	logEntries := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")})
	s := &Server{currentTerm: 2, votedFor: 1, log: logArray, commitIndex: -1, leader: 1, state: "LEADER"}
	logEntries = append(logEntries, LogInfo{term: 1, data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, 1, 1, logEntries, 0})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	expectInt(t, s.votedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append request made by client to Leader
func TestAppendRequestFromClientToLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	//peer id array
	peerArray = append(peerArray, 1, 2, 3, 4)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")})
	s := &Server{currentTerm: 2, votedFor: 5, log: logArray, commitIndex: 0, leader: 5, state: "LEADER", peers: peerArray}
	action := s.ProcessEvent(Append{[]byte("two")})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 2)
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	for i := 1; i <= 4; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		//check term
		expectInt(t, ind, 2)
		//check leaderId
		expectInt(t, vote, 5)
		//check receiver id as well
		expectInt(t, to, i)
	}
}

//Test append entry response sent by follower to leader indicating replication of logs from leader current term
func TestAppendEntryRespOfCurrentTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")}, LogInfo{term: 3, data: []byte("three")}, LogInfo{term: 4, data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{currentTerm: 4, votedFor: 5, log: logArray, commitIndex: 0, leader: 5, state: "LEADER", nextIndex: nextInd, matchIndex: matchInd}
	//follower with id 1 has sent a success reponse indicating it has appended all the entries starting from index 1 till index 3 (count=3, lastlogindex=0)
	action := s.ProcessEvent(AppendEntriesResp{1, 4, true, 3, 0})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 4)
	//commit index will change from 0 to 3
	expectInt(t, s.commitIndex, 3)
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	expectInt(t, ind, 3)
}

//Test append entry response sent by follower to leader indicating replication of logs but not from leader current term
func TestAppendEntryRespNotOfCurrentTermLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")}, LogInfo{term: 3, data: []byte("three")}, LogInfo{term: 4, data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{currentTerm: 5, votedFor: 5, log: logArray, commitIndex: 2, leader: 5, state: "LEADER", nextIndex: nextInd, matchIndex: matchInd}
	//follower with id 1 has sent a success reponse indicating it has appended all the entries starting from index 1 till index 3 (count=3, lastlogindex=0)  but not from leader current term
	_ = s.ProcessEvent(AppendEntriesResp{1, 4, true, 3, 0})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 5)
	//commit index should not change
	expectInt(t, s.commitIndex, 2)
}

//Test append entry response sent by follower with higher term to leader
func TestAppendEntryRespFromHigherTermServerToLeader(t *testing.T) {
	s := &Server{currentTerm: 2, votedFor: 5, commitIndex: 2, leader: 5, state: "LEADER"}
	//follower with id 1 having term 5 has sent reponse
	action := s.ProcessEvent(AppendEntriesResp{1, 5, false, 3, 0})
	expectString(t, s.state, "FOLLOWER")
	expectInt(t, s.currentTerm, 5)
	//commit index should not change
	expectInt(t, s.commitIndex, 2)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
}

//Test append entry false response sent by follower to leader indicating non replication of logs
func TestAppendEntryFalseRespLeader(t *testing.T) {
	logArray := make([]LogInfo, 0)
	logArray = append(logArray, LogInfo{term: 1, data: []byte("one")}, LogInfo{term: 2, data: []byte("two")}, LogInfo{term: 3, data: []byte("three")}, LogInfo{term: 4, data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{currentTerm: 5, votedFor: 5, log: logArray, commitIndex: 2, leader: 5, state: "LEADER", nextIndex: nextInd, matchIndex: matchInd}
	//follower with id 1 has sent a false reponse indicating it's log not match with leader's log when leader sent him append entry request with lastlogindex=2
	action := s.ProcessEvent(AppendEntriesResp{1, 4, false, 3, 2})
	expectString(t, s.state, "LEADER")
	expectInt(t, s.currentTerm, 5)
	//commit index should not change
	expectInt(t, s.commitIndex, 2)
	act, ev, ind, _, vote, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesReq")
	//check term
	expectInt(t, ind, 5)
	//check leaderId
	expectInt(t, vote, 5)
	//check candidate id as well
	expectInt(t, to, 1)
	typeaction := action[0]
	obj := typeaction.(Send)
	respObj := obj.event.(AppendEntriesReq)
	//leader needs to decrease last last index with 1
	expectInt(t, respObj.lastLogIndex, 1)
	expectInt(t, respObj.lastLogTerm, 2)
}

func getDetails(action []Action, index int) (string, string, int, bool, int, int) {
	var act, event string
	var term int
	var resp bool
	var vote int
	var to int
	typeaction := action[index]
	ty := reflect.TypeOf(typeaction)
	act = ty.Name()
	switch act {
	case "Send":
		obj := typeaction.(Send)
		to = obj.from
		ty = reflect.TypeOf(obj.event)
		event = ty.Name()
		switch event {
		case "VoteResp":
			respObj := obj.event.(VoteResp)
			term = respObj.term
			resp = respObj.voteGranted
		case "AppendEntriesResp":
			respObj := obj.event.(AppendEntriesResp)
			term = respObj.term
			resp = respObj.success
		case "VoteReq":
			respObj := obj.event.(VoteReq)
			term = respObj.term
			vote = respObj.candidateId
		case "AppendEntriesReq":
			respObj := obj.event.(AppendEntriesReq)
			term = respObj.term
			vote = respObj.leaderId
		}
	case "LogStore":
		obj := typeaction.(LogStore)
		term = obj.index
	case "Commit":
		obj := typeaction.(Commit)
		term = obj.index
	case "StateStore":
		obj := typeaction.(StateStore)
		term = obj.term
		vote = obj.votedFor

	}
	return act, event, term, resp, vote, to
}

func expectString(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}

func expectInt(t *testing.T, a int, b int) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
func expectBool(t *testing.T, a bool, b bool) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
