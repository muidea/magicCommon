package util

import "testing"

type Test struct {
	Name string `json:"name"`
}

func TestLoadConfig(t *testing.T) {
	type args struct {
		filePath string
		ptr      interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "testA",
			args: args{
				filePath: "",
				ptr:      nil,
			},
			wantErr: true,
		},
		{
			name: "testB",
			args: args{
				filePath: "./test.json",
				ptr:      &Test{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadConfig(tt.args.filePath, tt.args.ptr); (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	type args struct {
		filePath string
		ptr      interface{}
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
			if err := SaveConfig(tt.args.filePath, tt.args.ptr); (err != nil) != tt.wantErr {
				t.Errorf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
