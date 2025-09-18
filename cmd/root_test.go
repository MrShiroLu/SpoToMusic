package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command is properly configured
	if rootCmd.Use != "spotomusic" {
		t.Errorf("Expected root command use to be 'spotomusic', got %s", rootCmd.Use)
	}
	
	if rootCmd.Short == "" {
		t.Error("Expected root command to have a short description")
	}
	
	if rootCmd.Long == "" {
		t.Error("Expected root command to have a long description")
	}
}

func TestRootCommandFlags(t *testing.T) {
	// Test that required flags exist
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected 'verbose' flag to exist")
	}
	
	dryRunFlag := rootCmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected 'dry-run' flag to exist")
	}
}
