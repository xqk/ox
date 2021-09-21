package otime

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	type args struct {
		str string
	}
	type test struct {
		name string
		args args
		want time.Duration
	}
	var tests []*test
	tests = append(tests, &test{
		name: "2021-09-22 00:00:00",
		args: args{
			str: "",
		},
		want: 30,
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Duration(tt.args.str); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}