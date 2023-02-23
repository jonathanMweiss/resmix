package rrpc

import (
	"context"
	"sync"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
)

/*

Who holds the attestor?
it should be the relay.
the relay holds a struct of attestor which needs to receive a load.

The attestor needs to request a for each parcel.
if the parcel is part of, ignore.
on boot, attestor should request lists of times.

*/

type Workload struct {
	StartTime time.Time

	// for example can hold a struct which states the total work expected.
	Description any
}

// AccusationPlan describes for each hostname all workloads, and uuid expected.
type AccusationPlan struct {
	HostnameToAccuse string

	// uuid to Description.
	UuidToWorkload map[string]Workload
}

type WorkEstimator interface {
	// EstimateSystemTimes Used on boot/reboot to state the timeouts for each server.
	EstimateSystemTimes() []*AccusationPlan

	EstimateDuration(plan *AccusationPlan) time.Duration

	// UpdateDuration : on new note, should deduce whether it should update the acusation plan/ time given to this hostname.
	UpdateDuration(note *ExchangeNote, plan *AccusationPlan) time.Duration

	// IsDone states whether the exchangeNote states
	IsDone(plan AccusationPlan) bool // on true: dispose of the accusation plan
}

// who creates an attestor? the relay.
// what does it need to know?

type Attestor struct {
	E WorkEstimator

	Context    context.Context
	CancelFunc context.CancelFunc

	wg sync.WaitGroup

	input chan interface{}

	signingKey crypto.PrivateKey

	topItemTimer *time.Timer

	netdata NetworkData
}

// FailureMessage states that a failure in the system was detected by a majority of the nodes.
// receiving this message should trigger in the attestor a new estimation of the system.
type FailureMessage struct{}

func (a *Attestor) worker() {
	defer a.wg.Done()

	//todo := make(chan interface{}, 100)

	for {
		select {
		// clean channel:
		//case <-todo:
		//a.cleanmemory()
		//cleanChannelLoop:
		//	for {
		//		select {
		//		case <-a.input:
		//		default:
		//			break cleanChannelLoop
		//		}
		//	}
		case item := <-a.input:
			switch v := item.(type) {
			case *ExchangeNote:
				// TODO who is talking with whom, how do i get the accusation plan?
				_ = v
			case *FailureMessage:
				sysTimes := a.E.EstimateSystemTimes()
				_ = sysTimes
			}

			// I might receive a new note at the same time
		case <-a.topItemTimer.C:
			a.accuseTop()
		case <-a.Context.Done():
			a.topItemTimer.Stop()
			return
		}
	}
}

//
//import (
//	"bytes"
//	"container/heap"
//	"context"
//	"crypto/sha256"
//	"sync"
//	"time"
//
//	"github.com/jonathanMweiss/resmix/internal/crypto"
//)
//
///*
//1.
//	1.1. Need to receive parcels/ prrof-of-communications of direct calls.
//	1.2. Need to infer if uuid is done/ still need to keep times, and wait for more items.
//2. Need reload api -> upon failure detected, need to readjust the deadlines...
//4.
//*/
//// need to be able to accuse.
//// need to be informed of new data.
//// need to know if done receiving.
//// need to be able to reset self.
//

//type Common interface {
//	//AttestBroadcaster
//	//LoadEstimator
//	//OnlineState
//	ResetNotifier
//}
//
//// Attestor contain some proof of communication, on error, should drop that item ?
//// treat the error messages as no-show.
//// responsible for monitoring times of messages..
//// and to create attests once it takes too much time.
//type Attestor struct {
//	Context context.Context
//	context.CancelFunc
//
//	wg sync.WaitGroup
//
//	signingKey crypto.PrivateKey
//	//AttestBroadcaster
//	Estimator
//
//	topItemTimer *time.Timer
//	//doNotReport -> the heap has no way of removing from it nicely.
//	// just remember in a map not to report on specific uuid.
//	heap *attestPlanHeap
//
//	// If i get an attest and i saw the same uuid already, i will not change anything about it.
//	// in addition to that, if i receive a canceled call - that is, i've seen a response. then i can avoid
//	accusationCanceled map[AccuseId]bool
//
//	// need to insert new item. no need to pull from it.
//	// should be able to broadcast attestations.
//	// someone else collects attests, verifies them, and once ready will report on them?
//	input chan interface{} // all input types are the same.
//	name  string
//	//OnlineState        OnlineState
//	resetNotifications <-chan interface{}
//}
//
//type AccuseId [sha256.Size]byte
//
//func (a AccuseId) GetExoneratePlan() (AccuseId, error) {
//	return a, nil
//}
//
//// AccusationPlan contains workload to be used to estimate time, get sender and receiver ids, and to store uuid.
//type AccusationPlan struct {
//	// states the AccuseId to find this AccusationPlan
//	// assuming AccuseId is collision resistant
//	AccuseId
//
//	// name of the server that would be accused if it violates the Attestor's given time.
//	ToBeAccused []byte
//
//	// Attestor uses this Timestamp to estimate the deadline.
//	Timestamp time.Time
//
//	// WorkAmount states the amount of work unit where each unit should take the same amount of cpu time.
//	// for example 100 messages to decrypt or 3 ECC tasks
//	WorkAmount uint64
//}
//
//func (a AccusationPlan) GenerateAccusationPlan() AccusationPlan {
//	return a
//}
//
//type sortableProofOfComm struct {
//	AccusationPlan
//	endTime time.Time
//}
//
//// ServerFailureDetected after a server had failed, should readjust the deadlines(?) TODO
//func (a *Attestor) ServerFailureDetected() {} // maybe drop the timeouts...
//
//func (a *Attestor) WaitOn(ap Accuseable) {
//	a.in(ap.GenerateAccusationPlan())
//}
//
//// need to logic to asses whether one can attest or not.
//func (a *Attestor) attestMaker() {
//	defer a.wg.Done()
//	for {
//		select {
//		case <-a.resetNotifications:
//			a.cleanmemory()
//		cleanChannelLoop:
//			for {
//				select {
//				case <-a.input:
//				default:
//					break cleanChannelLoop
//				}
//			}
//		case item := <-a.input:
//			switch v := item.(type) {
//			case AccusationPlan:
//				a.receive(v)
//			case AccuseId:
//				a.accusationCanceled[v] = true
//			}
//		case <-a.topItemTimer.C:
//			a.accuseTop()
//		case <-a.Context.Done():
//			a.topItemTimer.Stop()
//			return
//		}
//	}
//}
//
//func (a *Attestor) accuseTop() {
//	if a.heap.Len() == 0 {
//		return
//	}
//	item := heap.Pop(a.heap).(sortableProofOfComm)
//
//	// attest only if the attest plan wasn't canceled yet.
//	if !a.accusationCanceled[item.AccuseId] {
//		a.accuse(item)
//	}
//
//	delete(a.accusationCanceled, item.AccuseId)
//
//	if a.heap.Len() == 0 {
//		return
//	}
//
//	a.SetTopAsTimer()
//}
//
//func (a *Attestor) accuse(item sortableProofOfComm) {
//	at := Attestation{Accused: item.ToBeAccused, AttesterName: a.name}
//
//	bf := bytes.NewBuffer(nil)
//	sig, err := a.signingKey.OWSign(&at, bf)
//	if err != nil {
//		return
//	}
//
//	at.Signature = sig
//	//a.BroadcastAttestation(at)
//}
//
//func (a *Attestor) receive(item AccusationPlan) {
//	// if seen: do nothing about this accusation.
//	if _, ok := a.accusationCanceled[item.AccuseId]; ok {
//		return
//	}
//
//	a.accusationCanceled[item.AccuseId] = false
//
//	heap.Push(a.heap, a.intoSortableProof(item))
//	// Assumes that endtime is always later on...
//
//	a.SetTopAsTimer()
//}
//
//func (a *Attestor) intoSortableProof(prf AccusationPlan) sortableProofOfComm {
//	return sortableProofOfComm{
//		AccusationPlan: prf,
//		endTime:        prf.Timestamp.Add(a.Estimator.EstimateResponseTime(prf)),
//	}
//}
//
//func (a *Attestor) SetTopAsTimer() {
//	endTime := a.heap.Peek().endTime
//
//	a.stopTimer()
//	a.topItemTimer.Reset(time.Until(endTime))
//}
//
//// TODO need to verify this!!
//func (a *Attestor) stopTimer() {
//	// stopping the timer, if its channel is not drained: drain it.
//	if !a.topItemTimer.Stop() && len(a.topItemTimer.C) > 0 {
//		<-a.topItemTimer.C
//	}
//}
//
//func (a *Attestor) cleanmemory() {
//	a.accusationCanceled = map[AccuseId]bool{}
//	tmp := &attestPlanHeap{}
//	heap.Init(tmp)
//	a.heap = tmp
//	//a.stopTimer()
//}
//
//func (a *Attestor) in(v interface{}) {
//	select {
//	case a.input <- v:
//	case <-a.Context.Done():
//	}
//}
//
//func (a *Attestor) Drop(exonerateable Exonerateable) {
//	id, err := exonerateable.GetExoneratePlan()
//	if err != nil {
//		return
//	}
//	a.in(id)
//}
//
//func (a *Attestor) Stop() {
//	a.CancelFunc()
//	a.wg.Wait()
//}
//
//func NewAttestor(relayName string, privkey crypto.PrivateKey, d Common) *Attestor {
//	ctx, cancel := context.WithCancel(context.Background())
//	a := &Attestor{
//		name:       relayName,
//		Context:    ctx,
//		CancelFunc: cancel,
//		signingKey: privkey,
//		//resetNotifications: d.RegisterForNotification("attestor"),
//		//AttestBroadcaster:  d,
//		//LoadEstimator:      d,
//		//OnlineState:        d,
//		topItemTimer:       time.NewTimer(time.Millisecond),
//		heap:               &attestPlanHeap{},
//		accusationCanceled: map[AccuseId]bool{},
//		input:              make(chan interface{}),
//	}
//
//	a.wg.Add(1)
//	go a.attestMaker()
//
//	return a
//}
//
//// MinimumTimeToDeadLineHeap
//type attestPlanHeap []sortableProofOfComm
//
//func (p attestPlanHeap) Len() int {
//	return len(p)
//}
//
//func (p attestPlanHeap) Less(i, j int) bool {
//	return p[i].endTime.Before(p[j].endTime)
//}
//
//func (p attestPlanHeap) Swap(i, j int) {
//	p[i], p[j] = p[j], p[i]
//}
//
//func (p *attestPlanHeap) Push(x interface{}) {
//	*p = append(*p, x.(sortableProofOfComm))
//}
//
//func (p *attestPlanHeap) Pop() interface{} {
//	tmp := *p
//	out := tmp[len(tmp)-1]
//	// shrink the underlying array
//	*p = tmp[:len(tmp)-1]
//	return out
//}
//
//func (p attestPlanHeap) Peek() *sortableProofOfComm {
//	if len(p) == 0 {
//		return nil
//	}
//	return &p[0]
//}
