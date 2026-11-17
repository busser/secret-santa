package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"

	"gopkg.in/yaml.v2"
)

// =============================================================================
// Main logic
// =============================================================================

const usage = `Usage:
  secret-santa draw [flags]   Draw the Santa chain and save each Santa's text.
  secret-santa send [flags]   Send each saved text that was not sent yet.

Run "secret-santa <command> -h" for the flags of a command.`

func main() {
	if len(os.Args) < 2 {
		log.Fatal(usage)
	}

	var err error
	switch os.Args[1] {
	case "draw":
		err = runDraw(os.Args[2:])
	case "send":
		err = runSend(os.Args[2:])
	default:
		log.Fatalf("Unknown command %q.\n\n%s", os.Args[1], usage)
	}
	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
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

func (ss secretSanta) shufflePeople(rng *rand.Rand) {
	rng.Shuffle(
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
