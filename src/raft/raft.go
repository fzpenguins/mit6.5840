package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"

	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int
	votedFor    int
	log         []Entry

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int

	electionTime      time.Time
	heartBeatTime     *time.Timer
	heartBeatInterval time.Duration

	identity  string
	voteCount int
}

const (
	FOLLOWER  = "FOLLOWER"
	CANDIDATE = "CANDIDATE"
	LEADER    = "LEADER"
)

type Entry struct {
	Command interface{}
	Index   int
	Term    int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	rf.mu.Lock()
	term = rf.currentTerm
	isleader = rf.identity == LEADER
	rf.mu.Unlock()
	// Your code here (3A).
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).

	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).

	Term        int
	VoteGranted bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).
	// log.Println(2)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// if rf.identity == LEADER {
	// 	DPrintf("%d %s S%d identity is %s,can not vote", time.Now().Unix(), dLeader, rf.me, rf.identity)
	// 	reply.VoteGranted = false
	// 	reply.Term = rf.currentTerm
	// 	return
	// }
	log.Println(2)
	if args.Term > rf.currentTerm || (args.Term == rf.currentTerm && rf.votedFor == -1) {
		rf.identity = FOLLOWER
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		rf.electionTime = time.Now()
		// rf.heartBeatTime.Reset(rf.heartBeatInterval)
		//这里heartBeatTIme不重置的话会不会导致重复选举
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
	} else {

		reply.Term = rf.currentTerm
		reply.VoteGranted = false
	}
	log.Println(rf.me, "   identity=", rf.identity)
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (3B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// ticker use to detect election time
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here (3A)
		// Check if a leader election should be started.

		// pause for a random amount of time between 50 and 350
		// milliseconds.
		now := time.Now()
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)

		rf.mu.Lock()

		if rf.electionTime.Before(now) && rf.identity != LEADER { // leader不会被进行rpc调用

			rf.ChangeIdentity(CANDIDATE)
			rf.currentTerm++
			DPrintf("%d %s S%d election time out, restart election at T%d", time.Now().Unix()%10000, dTimer, rf.me, rf.currentTerm)
			rf.mu.Unlock()
			go rf.startElection()
			continue
		}
		rf.mu.Unlock()

	}
}

func (rf *Raft) ChangeIdentity(identity string) {
	if identity == FOLLOWER {
		rf.electionTime = time.Now()
		rf.identity = FOLLOWER
		rf.votedFor = -1
		rf.voteCount = 0
		//这里就不对term进行自增了，因为不一定是只加一
	} else if identity == LEADER {
		rf.identity = LEADER
		rf.voteCount = 0
		rf.votedFor = -1

	} else {
		rf.votedFor = rf.me
		rf.voteCount = 1
		rf.identity = CANDIDATE
		rf.electionTime = time.Now()
	}
}

func (rf *Raft) startElection() {

	for i := 0; i < len(rf.peers); i++ {

		if i != rf.me {
			rf.mu.Lock()
			if rf.identity != CANDIDATE { // 不是候选人就失去了投票的意义

				rf.mu.Unlock()
				return
			}

			DPrintf("%d %s  S%d  send vote request to S%d at T%d", time.Now().Unix()%10000, dVote, rf.me, i, rf.currentTerm)
			rf.mu.Unlock()
			go rf.voteRequest(i)
		}
	}
}

// voteRequest 负责投票逻辑
func (rf *Raft) voteRequest(server int) {
	rf.mu.Lock()

	if rf.identity == LEADER { //被别的线程修改了身份，直接退出
		DPrintf("%d %s S%d is not candidate,not need to vote req", time.Now().Unix()%10000, dLeader, rf.me)
		rf.mu.Unlock()
		return
	}

	args := &RequestVoteArgs{
		Term:        rf.currentTerm,
		CandidateId: rf.me,
	}
	rf.mu.Unlock()
	reply := &RequestVoteReply{}

	if rf.sendRequestVote(server, args, reply) {

		rf.mu.Lock()
		defer rf.mu.Unlock()
		if rf.identity != CANDIDATE {
			DPrintf("%d %s S%d is not candidate,not need to vote req", time.Now().Unix()%10000, dLeader, rf.me)

			return
		}

		if reply.Term > rf.currentTerm {
			DPrintf("%d %s S%d term is old, reply term is %d,Convert to FOLLOWER at T%d", time.Now().Unix()%10000, dTimer, rf.me, reply.Term, rf.currentTerm)
			rf.ChangeIdentity(FOLLOWER)
			rf.currentTerm = reply.Term
			return
		}
		if reply.VoteGranted {
			rf.voteCount++
			DPrintf("%d %s  S%d  receive vote from S%d at T%d", time.Now().Unix()%10000, dVote, rf.me, server, rf.currentTerm)

			if rf.voteCount > len(rf.peers)/2 {
				DPrintf("%d %s  S%d  has %d votes,Convert to LEADER at T%d", time.Now().Unix()%10000, dLeader, rf.me, rf.voteCount, rf.currentTerm)
				rf.ChangeIdentity(LEADER)

				go rf.HeartBeat()
			}
		}

	}

}

func (rf *Raft) HeartBeat() {

	for !rf.killed() {
		time.Sleep(rf.heartBeatInterval)
		// rf.mu.Lock()
		// if rf.identity != LEADER {
		// 	rf.mu.Unlock()
		// 	return
		// }
		// rf.mu.Unlock()
		for i := 0; i < len(rf.peers); i++ {
			if i != rf.me {
				go func(server int) {
					args := &AppendEntriesArgs{
						Term:     rf.currentTerm,
						LeaderId: rf.me,
					}

					reply := &AppendEntriesReply{}
					rf.mu.Lock()
					if rf.identity != LEADER {
						rf.mu.Unlock()
						return
					}
					rf.mu.Unlock()
					if server != rf.me {
						if rf.sendAppendEntries(server, args, reply) {
							if !reply.Success {
								rf.mu.Lock()
								if reply.Term > rf.currentTerm {
									rf.currentTerm = reply.Term
									rf.ChangeIdentity(FOLLOWER)
								}
								rf.mu.Unlock()
							}
						}
					}
				}(i)
			}

		}
	}

}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	return rf.peers[server].Call("Raft.AppendEntries", args, reply)
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm { // 过期的当作无效处理
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	} else if args.Term == rf.currentTerm {

		reply.Success = true
		reply.Term = rf.currentTerm
		DPrintf("%d %s S%d AppendEntries call %d successfully at T%d", time.Now().Unix()%10000, dTerm, args.LeaderId, rf.me, args.Term)
		rf.currentTerm = args.Term
		rf.electionTime = time.Now()
		rf.heartBeatTime.Reset(rf.heartBeatInterval)

	} else {
		rf.electionTime = time.Now()
		rf.heartBeatTime.Reset(rf.heartBeatInterval)
		rf.identity = FOLLOWER
		reply.Success = false
		reply.Term = rf.currentTerm
		rf.currentTerm = args.Term
	}
	// log.Println(rf.me, "appendentries  identity = ", rf.identity)

}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.heartBeatInterval = 10 * time.Millisecond
	rf.heartBeatTime = time.NewTimer(rf.heartBeatInterval)
	rf.heartBeatTime.Stop()
	rf.electionTime = time.Now()

	rf.votedFor = -1
	rf.currentTerm = 1
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = 1
		rf.matchIndex[i] = 0
	}
	// Your initialization code here (3A, 3B, 3C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	DPrintf("server %d is inited\n", rf.me)
	// start ticker goroutine to start elections
	go rf.ticker()
	go rf.heartBeatTicker()
	return rf
}

func (rf *Raft) heartBeatTicker() {
	for !rf.killed() {
		select {

		case <-rf.heartBeatTime.C:
			rf.mu.Lock()
			if rf.identity != LEADER {
				rf.heartBeatTime.Stop()
				rf.ChangeIdentity(CANDIDATE)
				rf.currentTerm++
				DPrintf("%d %s S%d heartbeat time out, restart election at T%d", time.Now().Unix()%10000, dTimer, rf.me, rf.currentTerm)
				rf.mu.Unlock()

				go rf.startElection()
				continue
			}
			rf.mu.Unlock()
		}

	}
}

type logTopic string

const (
	dClient  logTopic = "CLNT"
	dCommit  logTopic = "CMIT"
	dDrop    logTopic = "DROP"
	dError   logTopic = "ERRO"
	dInfo    logTopic = "INFO"
	dLeader  logTopic = "LEAD"
	dLog     logTopic = "LOG1"
	dLog2    logTopic = "LOG2"
	dPersist logTopic = "PERS"
	dSnap    logTopic = "SNAP"
	dTerm    logTopic = "TERM"
	dTest    logTopic = "TEST"
	dTimer   logTopic = "TIMR"
	dTrace   logTopic = "TRCE"
	dVote    logTopic = "VOTE"
	dWarn    logTopic = "WARN"
)

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}
