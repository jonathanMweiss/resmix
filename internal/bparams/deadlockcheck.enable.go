//+build deadlock

package bparams

import (
	"fmt"
	"time"

	"github.com/algorand/go-deadlock"
)

const Deadlock = true

func init() {
	fmt.Println("deadlock mutexes enabled.")
	deadlock.Opts.Disable = false
	deadlock.Opts.DeadlockTimeout = time.Second * 30
	//deadlock.Opts.OnPotentialDeadlock = func() { panic("deadlock") }
}
