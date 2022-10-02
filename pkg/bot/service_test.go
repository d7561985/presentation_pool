package bot

import (
	"testing"
)

func Test_emailValidation(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"ok",
			args{email: "d.eee@qqqq-errr.team"},
			true,
		},
		{
			"false",
			args{email: "@"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emailValidation(tt.args.email); got != tt.want {
				t.Errorf("emailValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}
