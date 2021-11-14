package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// =============================================================================
// CLI flags
// =============================================================================

var (
	configFile = flag.String("config", "secret-santa.yml", "Path to the file containing the list of family members")
)

// =============================================================================
// Main logic
// =============================================================================

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}

func run() error {
	flag.Parse()

	var ss SecretSanta
	ss.loadFromFile(*configFile)

	// Prints raw data, useful during development.
	json.NewEncoder(os.Stdout).Encode(ss)

	return nil
}

// =============================================================================
// Data structures
// =============================================================================

type SecretSanta struct {
	Families []Family
}

type Family struct {
	Name    string
	Members []Person
}

type Person string

// =============================================================================
// Configuration parsing
// =============================================================================

func (ss *SecretSanta) loadFromFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(ss); err != nil {
		return fmt.Errorf("decoding yaml: %w", err)
	}

	return nil
}
