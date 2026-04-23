package cli

import "testing"

func TestNewWebReplicaCmdDefaultsToGeneratedSite(t *testing.T) {
	cmd := NewWebReplicaCmd()
	if cmd.Use != "webreplica [url]" {
		t.Fatalf("unexpected use line: %q", cmd.Use)
	}
	outFlag := cmd.PersistentFlags().Lookup("out")
	if outFlag == nil {
		t.Fatal("expected --out flag")
	}
	if outFlag.DefValue != "./generated-site" {
		t.Fatalf("expected webreplica default out dir to be ./generated-site, got %q", outFlag.DefValue)
	}
	if cmd.RunE == nil {
		t.Fatal("expected root command to support direct URL execution")
	}
}

func TestNewRootCmdKeepsSiteforgeDefaults(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "siteforge [url]" {
		t.Fatalf("unexpected use line: %q", cmd.Use)
	}
	outFlag := cmd.PersistentFlags().Lookup("out")
	if outFlag == nil {
		t.Fatal("expected --out flag")
	}
	if outFlag.DefValue != "./siteforge-output" {
		t.Fatalf("expected siteforge default out dir to remain ./siteforge-output, got %q", outFlag.DefValue)
	}
}
