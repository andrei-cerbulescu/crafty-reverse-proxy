package proxy

import (
	"sync/atomic"
)

// Constants representing the possible states of the state machine.
const (
	StateOff State = iota
	StateStartingUp
	StateRunning
	StateEmpty
	StateShuttingDown
)

func String(state State) string {
	switch state {
	case StateOff:
		return "Off"
	case StateStartingUp:
		return "StartingUp"
	case StateRunning:
		return "Running"
	case StateEmpty:
		return "Empty"
	case StateShuttingDown:
		return "ShuttingDown"
	default:
		return "unknown"
	}
}

// State represents the state of the state machine
type State = int32

type StateMachine struct {
	state  State
	logger Logger
}

func NewStateMachine(initial State, logger Logger) *StateMachine {
	return &StateMachine{
		state:  initial,
		logger: logger,
	}
}

func (sm *StateMachine) updateState(old, new State) bool {
	ok := atomic.CompareAndSwapInt32((*int32)(&sm.state), old, new)
	if ok {
		sm.logger.Debug("Updating state from %s to %s", String(old), String(new))
	} else {
		sm.logger.Debug("Failed to update state from %s to %s", String(old), String(new))
	}
	return ok
}
func (sm *StateMachine) SetState(newState State) (ok bool) {
	switch newState {
	case StateStartingUp:
		ok = sm.updateState(StateOff, StateStartingUp)
	case StateRunning:
		ok = sm.updateState(StateStartingUp, StateRunning) || sm.updateState(StateEmpty, StateRunning)
	case StateEmpty:
		ok = sm.updateState(StateRunning, StateEmpty)
	case StateShuttingDown:
		ok = sm.updateState(StateRunning, StateShuttingDown)
	case StateOff:
		ok = sm.updateState(StateShuttingDown, StateOff)
	}

	return
}

func (sm *StateMachine) GetState() State {
	return sm.state
}

func (sm *StateMachine) Reset(state State) {
	sm.logger.Debug("Resetting state from %s to %s", String(sm.state), String(state))
	atomic.StoreInt32((*int32)(&sm.state), state)
}
