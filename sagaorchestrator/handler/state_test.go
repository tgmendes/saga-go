package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tgmendes/saga-go/sagaorchestrator/handler"
)

func TestExecuteNext(t *testing.T) {
	var gotNextPayload []byte
	var gotCompensatePayload []byte
	tt := handler.StateTable{
		"start": {
			handler.Next: {
				State: "middle",
				TransitionFunc: func(b []byte) {
					gotNextPayload = b
				},
			},
		},
		"middle": {
			handler.Next: {
				State: "end",
				TransitionFunc: func(b []byte) {
					gotNextPayload = b
				},
			},
			handler.Compensate: {
				State: "start",
				TransitionFunc: func(b []byte) {
					gotCompensatePayload = b
				},
			},
		},
		"end": {
			handler.Next: {},
			handler.Compensate: {
				State: "middle",
				TransitionFunc: func(b []byte) {
					gotCompensatePayload = b
				},
			},
		},
	}

	tests := []struct {
		desc                 string
		state                string
		transition           handler.Transition
		expUpdatePayload     []byte
		expCompensatePayload []byte
	}{
		{
			desc:             "when updating start",
			state:            "start",
			transition:       handler.Next,
			expUpdatePayload: []byte("hello"),
		}, {
			desc:       "when compensating start",
			state:      "start",
			transition: handler.Compensate,
		}, {
			desc:             "when updating middle",
			state:            "middle",
			transition:       handler.Next,
			expUpdatePayload: []byte("hello"),
		}, {
			desc:                 "when compensating middle",
			state:                "middle",
			transition:           handler.Compensate,
			expCompensatePayload: []byte("hello"),
		}, {
			desc:       "when updating end",
			state:      "end",
			transition: handler.Next,
		}, {
			desc:                 "when compensating end",
			state:                "end",
			transition:           handler.Compensate,
			expCompensatePayload: []byte("hello"),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotNextPayload = nil
			gotCompensatePayload = nil
			err := tt.ExecuteNext(test.state, test.transition, []byte("hello"))
			assert.NoError(t, err)
			assert.Equal(t, test.expUpdatePayload, gotNextPayload)
			assert.Equal(t, test.expCompensatePayload, gotCompensatePayload)
		})
	}
}

func TestNextState(t *testing.T) {
	tt := handler.StateTable{
		"start": {
			handler.Next: {
				State: "middle",
			},
		},
		"middle": {
			handler.Next: {
				State: "end",
			},
			handler.Compensate: {
				State: "start",
			},
		},
		"end": {
			handler.Next: {},
			handler.Compensate: {
				State: "middle",
			},
		},
	}

	tests := []struct {
		desc         string
		state        string
		transition   handler.Transition
		expNextState string
	}{
		{
			desc:         "when updating start",
			state:        "start",
			transition:   handler.Next,
			expNextState: "middle",
		}, {
			desc:         "when compensating start",
			state:        "start",
			transition:   handler.Compensate,
			expNextState: "",
		}, {
			desc:         "when updating middle",
			state:        "middle",
			transition:   handler.Next,
			expNextState: "end",
		}, {
			desc:         "when compensating middle",
			state:        "middle",
			transition:   handler.Compensate,
			expNextState: "start",
		}, {
			desc:         "when updating end",
			state:        "end",
			transition:   handler.Next,
			expNextState: "",
		}, {
			desc:         "when compensating end",
			state:        "end",
			transition:   handler.Compensate,
			expNextState: "middle",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			next, err := tt.NextState(test.state, test.transition)
			assert.NoError(t, err)
			assert.Equal(t, test.expNextState, next)
		})
	}
}
