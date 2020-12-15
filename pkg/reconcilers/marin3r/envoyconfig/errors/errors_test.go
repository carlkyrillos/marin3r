package errors

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		t      ErrorReason
		method string
		msg    string
	}
	tests := []struct {
		name string
		args args
		want Error
	}{
		{
			"Returns a new Error",
			args{"SomeReason", "SomeMethod", "SomeMsg"},
			Error{"SomeReason", "SomeMethod", "SomeMsg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.t, tt.args.method, tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    Error
		want string
	}{
		{
			"Returns a string representation of Error",
			Error{"SomeReason", "SomeMethod", "SomeMsg"},
			"error in SomeMethod: SomeMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReasonForError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want ErrorReason
	}{
		{
			"Returns the Error Reason field",
			args{Error{"SomeReason", "SomeMethod", "SomeMsg"}},
			"SomeReason",
		},
		{
			"Returns the Unknown reason if not an Error",
			args{fmt.Errorf("unknown error")},
			UnknownError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReasonForError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReasonForError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAllRevisionsTainted(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Returns false",
			args{Error{"SomeReason", "SomeMethod", "SomeMsg"}},
			false,
		},
		{
			"Returns true",
			args{Error{AllRevisionsTaintedError, "SomeMethod", "SomeMsg"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllRevisionsTainted(tt.args.err); got != tt.want {
				t.Errorf("IsAllRevisionsTainted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRollbackOccurred(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Returns false",
			args{Error{"SomeReason", "SomeMethod", "SomeMsg"}},
			false,
		},
		{
			"Returns true",
			args{Error{RollbackOccurredError, "SomeMethod", "SomeMsg"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRollbackOccurred(tt.args.err); got != tt.want {
				t.Errorf("IsRollbackOccurred() = %v, want %v", got, tt.want)
			}
		})
	}
}
