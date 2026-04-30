package cmd_test

import (
	"strings"
	"testing"

	"github.com/lnw9jv/awsctx/cmd"
)

func TestAWSRegionsNotEmpty(t *testing.T) {
	if len(cmd.AWSRegions) == 0 {
		t.Fatal("AWSRegions is empty")
	}
}

func TestAWSRegionsFormat(t *testing.T) {
	for _, entry := range cmd.AWSRegions {
		parts := strings.SplitN(entry, "\t", 2)
		if len(parts) != 2 {
			t.Errorf("region entry missing tab-separated description: %q", entry)
			continue
		}
		region := parts[0]
		if !strings.Contains(region, "-") {
			t.Errorf("region name looks invalid: %q", region)
		}
	}
}

func TestAWSRegionsContainsCommon(t *testing.T) {
	want := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	index := make(map[string]bool, len(cmd.AWSRegions))
	for _, entry := range cmd.AWSRegions {
		index[strings.SplitN(entry, "\t", 2)[0]] = true
	}
	for _, r := range want {
		if !index[r] {
			t.Errorf("expected region %q not found in AWSRegions", r)
		}
	}
}
