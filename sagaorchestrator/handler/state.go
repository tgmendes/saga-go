package handler

import (
	"fmt"
)

type Transition string

const (
	Next       Transition = "next"
	Compensate Transition = "compensate"
)

type State struct {
	State          string
	TransitionFunc func(payload []byte)
}

type TransitionStates map[Transition]State
type StateTable map[string]TransitionStates

func (st StateTable) NextState(currStateType string, transition Transition) (string, error) {
	next, err := st.GetNextState(currStateType, transition)
	if err != nil {
		return "", err
	}

	return next.State, nil
}

func (st StateTable) ExecuteNext(currStateType string, transition Transition, payload []byte) error {
	next, err := st.GetNextState(currStateType, transition)
	if err != nil {
		return err
	}

	if next.TransitionFunc != nil {
		next.TransitionFunc(payload)
	}

	return nil
}

func (st StateTable) GetNextState(currentState string, transition Transition) (State, error) {
	curr, ok := st[currentState]
	if !ok {
		return State{}, fmt.Errorf("unknown state: %s", currentState)
	}

	return curr[transition], nil
}
