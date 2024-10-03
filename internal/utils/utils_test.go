package utils

import (
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	type args struct {
		dir      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid file path",
			args: args{
				dir:      "/var/log/",
				filename: "system.log",
			},
			want:    "/var/log/system.log",
			wantErr: false,
		},
		{
			name: "Invalid file path",
			args: args{
				dir:      "/var/log/",
				filename: "../system.log",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "File does not exist",
			args: args{
				dir:      "/var/log/",
				filename: "doesnotexist.log",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateFilePath(tt.args.dir, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
