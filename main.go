package main

import (
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

	var ss secretSanta
	ss.loadFromFile(*configFile)

	// Prints raw data, useful during development.
	fmt.Println(ss)

	return nil
}

// =============================================================================
// Data structures
// =============================================================================

type secretSanta struct {
	people []person
}

type person struct {
	name   string
	family string
}

// =============================================================================
// Configuration parsing
// =============================================================================

func (ss *secretSanta) loadFromFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	defer f.Close()

	var config struct {
		Families []struct {
			Name    string
			Members []string
		}
	}

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return fmt.Errorf("decoding yaml: %w", err)
	}

	for _, f := range config.Families {
		for _, p := range f.Members {
			ss.people = append(ss.people, person{p, f.Name})
		}
	}

	return nil
}
