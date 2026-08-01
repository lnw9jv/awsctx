package cmd

import "testing"

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
