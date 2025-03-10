package main

import (
    "gioui.org/app"
    "testing"
)

func Test_runApp(t *testing.T) {
	type args struct {
		in_window *app.Window
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runApp(tt.args.in_window); (err != nil) != tt.wantErr {
				t.Errorf("runApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
