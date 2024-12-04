package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
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
	if err := ss.loadFromFile(*configFile); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ss.shufflePeople()

	chain, err := ss.findChain()
	if err != nil {
		return fmt.Errorf("finding Santa chain: %w", err)
	}

	for i := range chain {
		santa := chain[i]
		giftee := chain[(i+1)%len(chain)]
		fmt.Printf("%s (%s) -> %s\n", santa.name, santa.phoneNumber, giftee.name)
	}

	return nil
}

// =============================================================================
// Data structures
// =============================================================================

type secretSanta struct {
	people []person
}

type person struct {
	name        string
	family      string
	phoneNumber string
}

func (p person) canBeSantaFor(other person) bool {
	return p.family != other.family
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
			Members []struct {
				Name  string
				Phone string
			}
		}
	}

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return fmt.Errorf("decoding yaml: %w", err)
	}

	for _, f := range config.Families {
		for _, p := range f.Members {
			ss.people = append(ss.people, person{p.Name, f.Name, p.Phone})
		}
	}

	return nil
}

// =============================================================================
// Finding possible Santa chains
// =============================================================================

func (ss secretSanta) shufflePeople() {
	rand.Shuffle(
		len(ss.people),
		func(i, j int) {
			ss.people[i], ss.people[j] = ss.people[j], ss.people[i]
		},
	)
}

func (ss secretSanta) findChain() ([]person, error) {
	chain := make([]person, len(ss.people))
	alreadyPlaced := make([]bool, len(ss.people))

	var pickPerson func(position int) bool
	pickPerson = func(position int) bool {
		for i, p := range ss.people {
			if alreadyPlaced[i] {
				continue
			}

			if position > 0 && !chain[position-1].canBeSantaFor(p) {
				continue
			}
			if position == len(ss.people)-1 && !p.canBeSantaFor(chain[0]) {
				// The last person in the chain must be Santa for the first
				// person in the chain.
				continue
			}

			chain[position] = p
			alreadyPlaced[i] = true

			if position == len(ss.people)-1 {
				// We found a chain!
				return true
			}

			if pickPerson(position + 1) {
				return true
			}

			chain[position] = person{}
			alreadyPlaced[i] = false
		}

		return false
	}

	if ok := pickPerson(0); !ok {
		return nil, errors.New("no valid Santa chain exists")
	}

	return chain, nil
}
