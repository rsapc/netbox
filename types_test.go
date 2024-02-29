package netbox

import "testing"

func Test_getObjectType(t *testing.T) {
	type args struct {
		aModel string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test passing a string",
			args: args{aModel: "device"},
			want: "dcim.device",
		},
		{
			name: "Test an invalid",
			args: args{aModel: "dummy"},
			want: "Invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getObjectType(tt.args.aModel); got != tt.want {
				t.Errorf("getModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
