//+build debug

package bparams

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/algorand/go-deadlock"
)

func TestDeadlock(t *testing.T) {

	if os.Getenv("BE_CRASHER") == "1" {
		makedeadlock()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestDeadlock")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

	makedeadlock()

}

func makedeadlock() {
	//fmt.Println("Going down in flames!")
	//os.Exit(1)
	m1 := deadlock.Mutex{}
	m2 := deadlock.Mutex{}
	go func() {
		m1.Lock()
		time.Sleep(deadlock.Opts.DeadlockTimeout)
		m2.Lock()
	}()

	go func() {
		m2.Lock()
		time.Sleep(deadlock.Opts.DeadlockTimeout)
		m1.Lock()
	}()
	time.Sleep(deadlock.Opts.DeadlockTimeout * 10)
}
