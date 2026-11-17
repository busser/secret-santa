package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

func runDraw(args []string) error {
	flags := flag.NewFlagSet("draw", flag.ExitOnError)
	configFile := flags.String("config", "secret-santa.yml", "Path to the file containing the list of family members")
	messagesFile := flags.String("messages", "messages.yml", "Path to the file containing the message templates")
	drawFile := flags.String("state", "draw.json", "Path to the file to save the draw to")
	flags.Parse(args)

	var ss secretSanta
	if err := ss.loadFromFile(*configFile); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := ss.checkReadyToSend(); err != nil {
		return err
	}

	templates, err := loadTemplates(*messagesFile)
	if err != nil {
		return fmt.Errorf("loading messages: %w", err)
	}
	if len(templates) < len(ss.people) {
		return fmt.Errorf("%s has %d messages for %d people: each Santa needs a different one", *messagesFile, len(templates), len(ss.people))
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ss.shufflePeople(rng)
	chain, err := ss.findChain()
	if err != nil {
		return fmt.Errorf("finding Santa chain: %w", err)
	}

	// Each Santa gets a different template, so that two people who compare
	// their texts never see the same wording.
	templateOrder := rng.Perm(len(templates))
	d := make(draw, len(chain))
	for i, santa := range chain {
		giftee := chain[(i+1)%len(chain)]
		var message strings.Builder
		data := struct{ Santa, Giftee string }{santa.name, giftee.name}
		if err := templates[templateOrder[i]].Execute(&message, data); err != nil {
			return fmt.Errorf("rendering message: %w", err)
		}
		d[i] = drawnText{Santa: santa.name, Message: message.String()}
	}

	if err := createDrawFile(*drawFile, d); err != nil {
		return fmt.Errorf("saving draw: %w", err)
	}
	fmt.Printf("Saved the draw for %d people to %s.\n", len(d), *drawFile)
	return nil
}

// checkReadyToSend makes sure that send can later find each Santa's phone
// number by name.
func (ss secretSanta) checkReadyToSend() error {
	seen := make(map[string]bool)
	var missingPhone []string
	for _, p := range ss.people {
		if seen[p.name] {
			return fmt.Errorf("two people are named %s: names must be unique", p.name)
		}
		seen[p.name] = true
		if p.phoneNumber == "" {
			missingPhone = append(missingPhone, p.name)
		}
	}
	if len(missingPhone) > 0 {
		return fmt.Errorf("no phone number for %s", strings.Join(missingPhone, ", "))
	}
	return nil
}

func loadTemplates(fileName string) ([]*template.Template, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var file struct {
		Messages []string
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decoding yaml: %w", err)
	}

	templates := make([]*template.Template, len(file.Messages))
	for i, text := range file.Messages {
		t, err := template.New(fmt.Sprintf("message %d", i+1)).Parse(text)
		if err != nil {
			return nil, err
		}
		templates[i] = t
	}
	return templates, nil
}
