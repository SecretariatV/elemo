package asynq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestTaskType_String(t *testing.T) {
	tests := []struct {
		name string
		t    TaskType
		want string
	}{
		{
			name: "health check task",
			t:    TaskTypeSystemHealthCheck,
			want: "system:health_check",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.t.String())
		})
	}
}

func TestWithTaskLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := new(baseTaskHandler)
			err := WithTaskLogger(tt.args.logger)(handler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, handler.logger)
		})
	}
}

func TestWithTaskTracer(t *testing.T) {
	type args struct {
		tracer trace.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    trace.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := new(baseTaskHandler)
			err := WithTaskTracer(tt.args.tracer)(handler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, handler.tracer)
		})
	}
}
