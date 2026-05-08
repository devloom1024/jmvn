package cmd

import "context"

type executionStateKey struct{}

func withExecutionState(ctx context.Context, state *executionState) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, executionStateKey{}, state)
}

func executionStateFromContext(ctx context.Context) *executionState {
	state, _ := ctx.Value(executionStateKey{}).(*executionState)
	return state
}
