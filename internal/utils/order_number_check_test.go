package utils

import "testing"

func TestOrderNumberCheck(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty string",
			args: args{""},
			want: false,
		},
		{
			name: "wrong number",
			args: args{"4561261212345464"},
			want: false,
		},
		{
			name: "with space",
			args: args{"4561261 212345464"},
			want: false,
		},
		{
			name: "with literal",
			args: args{"4561261a212345464"},
			want: false,
		},
		{
			name: "ok",
			args: args{"4561261212345467"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OrderNumberCheck(tt.args.n); got != tt.want {
				t.Errorf("OrderNumberCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
