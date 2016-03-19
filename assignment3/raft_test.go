package main

import (
	"fmt"
	"reflect"
	"testing"
)

//*****Follower Testing**************************************

//Test for Success response of vote request received from a higher Term candidate
func TestPositiveVoteRequestFollower(t *testing.T) {
	//Test for positive response of vote request
	s := &Server{CurrentTerm: 0, VotedFor: 0, Leader: 0, State: "FOLLOWER"}
	action := make([]Action, 0)
	// Term:1,candidate id:1
	action = s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 1)
	expectInt(t, s.VotedFor, 1)
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

//Test for negative response of vote request if vote request is received from lower Term candidate
func TestNegativeVoteRequestWithLowTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 0, State: "FOLLOWER"}
	// Term:1,candidate id:1
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 0)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test duplicate vote Request from same server id and Term for which follower has already voted
func TestDuplicateVoteRequestFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, Leader: 0, State: "FOLLOWER"}
	// Term:1,candidate id:1
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, true)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test vote Request if follower has already granted his vote for the Term and receive vote request for the same Term from another server
func TestVoteRequestForSameTermTwiceFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, Leader: 0, State: "FOLLOWER"}
	// Term:2,candidate id:3
	action := s.ProcessEvent(VoteReq{2, 3, 1, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 3)
}

//Test for negative response of vote request from a server with higher Term with not upto date Logs
func TestNegativeVoteRequestWithHigherTermFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: 0, Leader: 0, State: "FOLLOWER"}
	// Term:5,candidate id:1,LastLogIndex:0,LastLogTerm:1
	action := s.ProcessEvent(VoteReq{5, 1, 0, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.VotedFor, 0)
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

//Test for negative response of vote request from a server with higher Term and it's previous Log index not matching with follower previous Log index
func TestNegativeVoteRequestWithLowerIndexFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, Leader: 0, State: "FOLLOWER"}
	// Term:3,candidate id:1,LastLogIndex:0,LastLogTerm:1
	action := s.ProcessEvent(VoteReq{3, 1, 0, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
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

//Test for negative response of vote request from a server with higher Term and it's previous Log index matches with follower previous Log index but previous Log Term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: 0, Leader: 0, State: "FOLLOWER"}
	// Term:3,candidate id:1,LastLogIndex:1,LastLogTerm:1
	action := s.ProcessEvent(VoteReq{3, 1, 1, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
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

//Test vote response request with same Term
func TestVoteRespFromSameTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	//Term:2,VoteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{2, false, 3})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test vote response request with lower Term
func TestVoteRespFromLowerTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	//Term:1,VoteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{1, false, 3})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test vote response request with higher Term
func TestVoteRespFromHigherTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	//Term:3,VoteGranted=false,server id:3
	action := s.ProcessEvent(VoteResp{3, false, 3})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for Successfull append entry request
func TestSuccessAppendEntryRequestFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	s := &Server{CurrentTerm: 1, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, -1, -1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 1)
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

//Test for Successfull append entry request from higher Term server
func TestSuccessAppendEntryRequestFromHigherTermFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 3, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.Leader, 1)
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

//Test for append entry request from a sever with higher Term but with not matching previous Log index and Term
func TestAppendEntryRequestWithHigherTermHeartbeatFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "FOLLOWER", Peers: peerArray}
	//LogEntries=append(LogEntries,LogInfo{Term:1,Data:[]byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 0)
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

//Test for Successfull append entry request from same Term Leader but Logs entries are having Data
func TestSuccessAppendEntryRequestFromSameTermFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 2, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 1)
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

//Test for Successfull append entry request from same Term Leader but Logs entries are having more Data beyond last Log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")}, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 2, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 1)
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

//Test for append entry request from a sever with same Term but with not matching previous Log index
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevIndexFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 3, 2, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append entry request from a sever with higher Term but with  matching previous Log index but not Term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "FOLLOWER", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	expectInt(t, s.Leader, 1)
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

//Test append entry response request with same Term from serverid 3
func TestAppendEntryRespFromSameTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 2, false, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with lower Term from serverid 3
func TestAppendEntryRespFromLowerTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 1, false, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with higher Term from serverid 3
func TestAppendEntryRespFromHigherTermFollower(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "FOLLOWER"}
	action := s.ProcessEvent(AppendEntriesResp{3, 3, false, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

//Test for timeout Event for follower
func TestTimeOutFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4, 5)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 1, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "FOLLOWER", Peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	expectInt(t, vote, 0)
	var ev string
	var ind int
	var to int
	//check whether vote request has sent to 4 other servers from the sever id 5
	for i := 1; i <= 5; i++ {
		act, ev, ind, _, vote, to = getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "VoteReq")
		expectInt(t, ind, 2)
		expectInt(t, vote, 0)
		//check candidate id as well
		expectInt(t, to, i)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ = getDetails(action, 6)
	expectString(t, act, "Alarm")
	//check time stamp**********

}

func TestAppendRequestFromClientToFollower(t *testing.T) {
	s := &Server{CurrentTerm: 1, VotedFor: 1, Leader: 1, State: "FOLLOWER"}
	action := s.ProcessEvent(Append{[]byte("hello")})
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	typeaction := action[0]
	obj := typeaction.(Commit)
	expectString(t, obj.Error, "Leader is at Id 1")

}

//***********************Candidate Testing****************************************************************
func TestPositiveVoteRequestCandidate(t *testing.T) {
	//Test for positive response of vote request
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{CurrentTerm: 1, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := make([]Action, 0)
	action = s.ProcessEvent(VoteReq{2, 1, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
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
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with same Term and it's previous Log index not matching with follower previous Log index
func TestNegativeVoteRequestWithLowerIndexCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 0, 1})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for negative response of vote request from a server with higher Term and it's previous Log index matches with other candidate's previous Log index but previous Log Term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with higher Term and it's previous Log index matches with other candidate's previous Log index but previous Log Term differs
func TestNegativeVoteRequestFromHigherTermCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{5, 1, 1, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.VotedFor, 0)
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

//Test for candidate receiving vote response and become Leader when you receive majority of votes
func TestVoteResponseCandidateToLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 5)
	//two yes vote already present in array
	voteArray[0] = 1
	voteArray[4] = 1
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray, VoteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{2, true, 2})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	for i := 0; i <= 3; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		expectInt(t, ind, 2)
		expectInt(t, vote, 0)
		//check candidate id as well
		expectInt(t, to, i+1)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ := getDetails(action, 4)
	expectString(t, act, "Alarm")
}

//Test for candidate receiving vote response and remain in candidate State untill you receive majority of votes
func TestVoteResponseRemainCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 5)
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray, VoteGranted: voteArray}
	_ = s.ProcessEvent(VoteResp{2, true, 2})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)

}

//Test for candidate receiving vote response and convert to follower if you receive majority of no votes
func TestVoteResponseCandidateToFollower(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 5)
	//three no vote already present in array
	voteArray[0] = -1
	voteArray[1] = -1
	voteArray[4] = -1
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray, VoteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{2, false, 3})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Alarm")

}

//Test for candidate receiving vote response and convert to follower if you receive vote response from higher Term server
func TestVoteResponseCandidateToFollowerAndUpdateTerm(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 5)
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray, VoteGranted: voteArray}
	action := s.ProcessEvent(VoteResp{3, false, 3})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
	act, _, _, _, _, _ = getDetails(action, 1)
	expectString(t, act, "Alarm")

}

//Test for candidate receiving vote response and remain in candidate State if it receives vote response from lower Term
func TestVoteResponseRemainInCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	voteArray := make([]int, 5)
	//already has one yes vote
	voteArray[0] = 1
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "CANDIDATE", Peers: peerArray, VoteGranted: voteArray}
	_ = s.ProcessEvent(VoteResp{1, true, 2})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)

}

//Test for timeout Event for candidate
func TestTimeOutCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	peerArray = append(peerArray, 1, 2, 3, 4)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 1, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	act, _, trm, _, _, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 2)
	var ev string
	var ind int
	var to int
	//check whether vote request has sent to 4 other servers from the sever id 5
	for i := 1; i <= 4; i++ {
		act, ev, ind, _, _, to = getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "VoteReq")
		expectInt(t, ind, 2)
		//check candidate id as well
		expectInt(t, to, i)
	}
	//check the presence of alarm action
	act, _, _, _, _, _ = getDetails(action, 5)
	expectString(t, act, "Alarm")
	//check time stamp**********

}

//Test for Successfull append entry request from higher Term server with upto date Logs
func TestSuccessAppendEntryRequestCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 3, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
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

//Test for Successfull append entry request from same Term server but more upto date Logs
func TestSuccessAppendEntryRequestFromUptoDateLogsCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 2, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	act, ev, ind, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectInt(t, ind, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for append entry request from a sever with higher Term but with not matching previous Log index and Term
func TestAppendEntryRequestWithHigherTermHeartbeatCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "CANDIDATE", Peers: peerArray}
	//LogEntries=append(LogEntries,LogInfo{Term:1,Data:[]byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 0)
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

//Test for Successfull append entry request from same Term Leader but Logs entries are having more Data beyond last Log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")}, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "CANDIDATE", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 2, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{2, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 1)
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

//Test for append entry request from a sever with higher Term but with  matching previous Log index but not Term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "CANDIDATE", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	expectInt(t, s.Leader, 1)
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

//Test for append entry request from a sever with lower Term Term
func TestFailAppendEntryRequestFromLowerTermCandidate(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "CANDIDATE", Peers: peerArray}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test append entry response request with same Term from serverid 3
func TestAppendEntryRespFromSameTermCandidate(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 2, false, 0, 0})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with lower Term from serverid 3
func TestAppendEntryRespFromLowerTermCandidate(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 1, false, 0, 0})
	expectString(t, s.State, "CANDIDATE")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
}

//Test append entry response request with higher Term from serverid 3
func TestAppendEntryRespFromHigherTermCandidate(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 1, State: "CANDIDATE"}
	action := s.ProcessEvent(AppendEntriesResp{3, 3, false, 0, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 3)
	expectInt(t, vote, 0)
}

func TestAppendRequestFromClientToCandidate(t *testing.T) {
	s := &Server{CurrentTerm: 1, VotedFor: 1, Leader: 1, State: "CANDIDATE"}
	action := s.ProcessEvent(Append{[]byte("hello")})
	act, _, _, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	typeaction := action[0]
	obj := typeaction.(Commit)
	expectString(t, obj.Error, "Leader is at Id 1")

}

//**************Leader Testing*************************************************************
func TestPositiveVoteRequestLeader(t *testing.T) {
	//Test for positive response of vote request
	LogArray := make([]LogInfo, 0)
	//peerArray := make([]int, 0)
	s := &Server{CurrentTerm: 1, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "LEADER"}
	action := make([]Action, 0)
	action = s.ProcessEvent(VoteReq{2, 1, 0, 0})
	//fmt.Println(action)
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
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
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{1, 1, 0, 0})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, trm, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectInt(t, trm, 2)
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with same Term and it's previous Log index not matching with follower previous Log index
func TestNegativeVoteRequestWithLowerIndexLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 0, 1})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)

}

//Test for negative response of vote request from a server with higher Term and it's previous Log index matches with other candidate's previous Log index but previous Log Term differs
func TestNegativeVoteRequestWithNonMatchingPrevLogTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{2, 1, 1, 1})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 5)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "VoteResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for negative response of vote request from a server with higher Term and it's previous Log index matches with other candidate's previous Log index but previous Log Term differs
func TestNegativeVoteRequestFromHigherTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(VoteReq{5, 1, 1, 1})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.VotedFor, 0)
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

//Test for vote resp request receieved by Leader
func TestVoteRespLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 0, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(VoteResp{2, false, 1})
	length := len(action)
	//action array length should be zero
	expectInt(t, length, 0)
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)

}

//Test for Leader timeout
func TestTimeOutLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	//peer id array
	peerArray = append(peerArray, 1, 2, 3, 4)
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 5, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(Timeout{})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	for i := 0; i <= 3; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		//check Term
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

//Test for Successfull append entry request from higher Term server with uptodate Logs
func TestSuccessAppendEntryRequestLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: -1, Leader: 0, State: "LEADER"}
	LogEntries = append(LogEntries, LogInfo{Term: 3, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
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

//Test for append entry request from a sever with higher Term but with not matching previous Log index and Term
func TestAppendEntryRequestWithHigherTermHeartbeatLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "LEADER"}
	action = s.ProcessEvent(AppendEntriesReq{5, 1, -1, -1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 0)
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

//Test for Successfull append entry request from higher Term Leader but Logs entries are having more Data beyond last Log index
func TestSuccessAppendEntryRequestHavingExtraLogEntriesLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("hello")}, LogInfo{Term: 1, Data: []byte("hello")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 0, State: "LEADER"}
	LogEntries = append(LogEntries, LogInfo{Term: 2, Data: []byte("bye")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 0, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.Leader, 1)
	expectInt(t, s.VotedFor, 0)
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

//Test for append entry request from a sever from a  higher Term server but with  matching previous Log index but not Term
func TestFailAppendEntryRequestWithSameTermNonMatchingPrevTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "LEADER"}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{3, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 3)
	expectInt(t, s.VotedFor, 0)
	expectInt(t, s.Leader, 1)
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

//Test for append entry request from a sever with lower Term Term
func TestFailAppendEntryRequestFromLowerTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	action := make([]Action, 0)
	LogEntries := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")})
	s := &Server{CurrentTerm: 2, VotedFor: 1, Log: LogArray, CommitIndex: -1, Leader: 1, State: "LEADER"}
	LogEntries = append(LogEntries, LogInfo{Term: 1, Data: []byte("one")})
	action = s.ProcessEvent(AppendEntriesReq{1, 1, 1, 1, LogEntries, 0})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	expectInt(t, s.VotedFor, 1)
	act, ev, _, flag, _, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesResp")
	expectBool(t, flag, false)
	//check candidate id as well
	expectInt(t, to, 1)
}

//Test for append request made by client to Leader
func TestAppendRequestFromClientToLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	peerArray := make([]int, 0)
	//peer id array
	peerArray = append(peerArray, 1, 2, 3, 4)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")})
	s := &Server{CurrentTerm: 2, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 5, State: "LEADER", Peers: peerArray}
	action := s.ProcessEvent(Append{[]byte("two")})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 2)
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "LogStore")
	expectInt(t, ind, 1)
	for i := 1; i <= 4; i++ {
		act, ev, ind, _, vote, to := getDetails(action, i)
		expectString(t, act, "Send")
		expectString(t, ev, "AppendEntriesReq")
		//check Term
		expectInt(t, ind, 2)
		//check LeaderId
		expectInt(t, vote, 5)
		//check receiver id as well
		expectInt(t, to, i)
	}
}

//Test append entry response sent by follower to Leader indicating replication of Logs from Leader current Term
func TestAppendEntryRespOfCurrentTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")}, LogInfo{Term: 3, Data: []byte("three")}, LogInfo{Term: 4, Data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{CurrentTerm: 4, VotedFor: 5, Log: LogArray, CommitIndex: 0, Leader: 5, State: "LEADER", NextIndex: nextInd, MatchIndex: matchInd}
	//follower with id 1 has sent a Success reponse indicating it has appended all the entries starting from index 1 till index 3 (count=3, LastLogIndex=0)
	action := s.ProcessEvent(AppendEntriesResp{1, 4, true, 3, 0})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 4)
	//commit index will change from 0 to 3
	expectInt(t, s.CommitIndex, 3)
	act, _, ind, _, _, _ := getDetails(action, 0)
	expectString(t, act, "Commit")
	expectInt(t, ind, 3)
}

//Test append entry response sent by follower to Leader indicating replication of Logs but not from Leader current Term
func TestAppendEntryRespNotOfCurrentTermLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")}, LogInfo{Term: 3, Data: []byte("three")}, LogInfo{Term: 4, Data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{CurrentTerm: 5, VotedFor: 5, Log: LogArray, CommitIndex: 2, Leader: 5, State: "LEADER", NextIndex: nextInd, MatchIndex: matchInd}
	//follower with id 1 has sent a Success reponse indicating it has appended all the entries starting from index 1 till index 3 (count=3, LastLogIndex=0)  but not from Leader current Term
	_ = s.ProcessEvent(AppendEntriesResp{1, 4, true, 3, 0})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 5)
	//commit index should not change
	expectInt(t, s.CommitIndex, 2)
}

//Test append entry response sent by follower with higher Term to Leader
func TestAppendEntryRespFromHigherTermServerToLeader(t *testing.T) {
	s := &Server{CurrentTerm: 2, VotedFor: 5, CommitIndex: 2, Leader: 5, State: "LEADER"}
	//follower with id 1 having Term 5 has sent reponse
	action := s.ProcessEvent(AppendEntriesResp{1, 5, false, 3, 0})
	expectString(t, s.State, "FOLLOWER")
	expectInt(t, s.CurrentTerm, 5)
	//commit index should not change
	expectInt(t, s.CommitIndex, 2)
	act, _, trm, _, vote, _ := getDetails(action, 0)
	expectString(t, act, "StateStore")
	expectInt(t, trm, 5)
	expectInt(t, vote, 0)
}

//Test append entry false response sent by follower to Leader indicating non replication of Logs
func TestAppendEntryFalseRespLeader(t *testing.T) {
	LogArray := make([]LogInfo, 0)
	LogArray = append(LogArray, LogInfo{Term: 1, Data: []byte("one")}, LogInfo{Term: 2, Data: []byte("two")}, LogInfo{Term: 3, Data: []byte("three")}, LogInfo{Term: 4, Data: []byte("four")})
	nextInd := make([]int, 0)
	nextInd = append(nextInd, 4, 4, 4, 4)
	matchInd := make([]int, 0)
	matchInd = append(matchInd, 0, 3, 1, 1)
	s := &Server{CurrentTerm: 5, VotedFor: 5, Log: LogArray, CommitIndex: 2, Leader: 5, State: "LEADER", NextIndex: nextInd, MatchIndex: matchInd}
	//follower with id 1 has sent a false reponse indicating it's Log not match with Leader's Log when Leader sent him append entry request with LastLogIndex=2
	action := s.ProcessEvent(AppendEntriesResp{1, 4, false, 3, 2})
	expectString(t, s.State, "LEADER")
	expectInt(t, s.CurrentTerm, 5)
	//commit index should not change
	expectInt(t, s.CommitIndex, 2)
	act, ev, ind, _, vote, to := getDetails(action, 0)
	expectString(t, act, "Send")
	expectString(t, ev, "AppendEntriesReq")
	//check Term
	expectInt(t, ind, 5)
	//check LeaderId
	expectInt(t, vote, 5)
	//check candidate id as well
	expectInt(t, to, 1)
	typeaction := action[0]
	obj := typeaction.(Send)
	respObj := obj.Event.(AppendEntriesReq)
	//Leader needs to decrease last last index with 1
	expectInt(t, respObj.LastLogIndex, 1)
	expectInt(t, respObj.LastLogTerm, 2)
}

func getDetails(action []Action, index int) (string, string, int, bool, int, int) {
	var act, Event string
	var Term int
	var resp bool
	var vote int
	var to int
	typeaction := action[index]
	ty := reflect.TypeOf(typeaction)
	act = ty.Name()
	switch act {
	case "Send":
		obj := typeaction.(Send)
		to = obj.From
		ty = reflect.TypeOf(obj.Event)
		Event = ty.Name()
		switch Event {
		case "VoteResp":
			respObj := obj.Event.(VoteResp)
			Term = respObj.Term
			resp = respObj.VoteGranted
		case "AppendEntriesResp":
			respObj := obj.Event.(AppendEntriesResp)
			Term = respObj.Term
			resp = respObj.Success
		case "VoteReq":
			respObj := obj.Event.(VoteReq)
			Term = respObj.Term
			vote = respObj.CandidateId
		case "AppendEntriesReq":
			respObj := obj.Event.(AppendEntriesReq)
			Term = respObj.Term
			vote = respObj.LeaderId
		}
	case "LogStore":
		obj := typeaction.(LogStore)
		Term = obj.Index
	case "Commit":
		obj := typeaction.(Commit)
		Term = obj.Index
	case "StateStore":
		obj := typeaction.(StateStore)
		Term = obj.Term
		vote = obj.VotedFor

	}
	return act, Event, Term, resp, vote, to
}

func expectBool(t *testing.T, a bool, b bool) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
