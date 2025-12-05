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

func TestInterface_GetMacAddress(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		mac  string
		want string
	}{
		{
			name: "Test getting new MAC address",
			mac:  "00:11:22:33:44:55",
			want: "00:11:22:33:44:55",
		},
		{
			name: "Test no MAC address",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macaddr := &MAC{MacAddress: &tt.mac}
			i := Interface{
				PrimaryMAC: macaddr,
			}
			got := i.GetMacAddress()
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetMacAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
