package main

import "testing"

func Test_getCurrentIp(t *testing.T) {
	tests := []struct {
		name   string
		wantIp string
	}{
		// TODO: Add test cases.
		{
			name: "132",
			wantIp: "10.23.6.52",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIp := getCurrentIp(); gotIp != tt.wantIp {
				t.Errorf("getCurrentIp() = %v, want %v", gotIp, tt.wantIp)
			}
		})
	}
}
