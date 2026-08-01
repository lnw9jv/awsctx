package cmd

import "testing"

func TestStatusMessage(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		region  string
		cleared bool
		want    string
	}{
		{name: "profile", profile: "dev", want: "Switched to profile dev\n"},
		{name: "region", region: "us-east-1", want: "Switched to region us-east-1\n"},
		{name: "profile and region", profile: "dev", region: "ap-southeast-1", want: "Switched to profile dev and region ap-southeast-1\n"},
		{name: "cleared", cleared: true, want: "Cleared AWS profile and region\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusMessage(tt.profile, tt.region, tt.cleared); got != tt.want {
				t.Errorf("statusMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateExportValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "plain", value: "dev; still data"},
		{name: "newline", value: "dev\nprod", wantErr: true},
		{name: "carriage return", value: "dev\rprod", wantErr: true},
		{name: "tab", value: "dev\tprod", wantErr: true},
		{name: "nul", value: "dev\x00prod", wantErr: true},
		{name: "delete", value: "dev\x7fprod", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExportValue("profile", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExportValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
