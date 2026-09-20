package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGroupsHeaderDefault(t *testing.T) {
	viper.Reset()
	SetDefaults()

	cfg, err := GetConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GroupsHeader != "Remote-Groups" {
		t.Errorf("GroupsHeader default = %q, want %q", cfg.GroupsHeader, "Remote-Groups")
	}
}
