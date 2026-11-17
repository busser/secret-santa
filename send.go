package main

import (
	"flag"
	"fmt"
	"time"
)

const ovhSMSService = "sms-dq52322-1"

func runSend(args []string) error {
	flags := flag.NewFlagSet("send", flag.ExitOnError)
	configFile := flags.String("config", "secret-santa.yml", "Path to the file containing the list of family members")
	drawFile := flags.String("state", "draw.json", "Path to the file that draw saved")
	flags.Parse(args)

	var ss secretSanta
	if err := ss.loadFromFile(*configFile); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	// Phone numbers come from the config, not the draw, so that a wrong number
	// can be fixed after the draw.
	phoneNumbers := make(map[string]string)
	for _, p := range ss.people {
		phoneNumbers[p.name] = p.phoneNumber
	}

	d, err := readDrawFile(*drawFile)
	if err != nil {
		return fmt.Errorf("loading draw: %w", err)
	}

	ovh, err := ovhClientFromEnv()
	if err != nil {
		return err
	}

	failures := 0
	for i := range d {
		text := &d[i]
		if text.SentAt != nil {
			fmt.Printf("Already sent to %s.\n", text.Santa)
			continue
		}

		phoneNumber := phoneNumbers[text.Santa]
		if phoneNumber == "" {
			fmt.Printf("Failed for %s: no phone number in %s.\n", text.Santa, *configFile)
			failures++
			continue
		}

		credits, err := ovh.sendSMS(ovhSMSService, phoneNumber, text.Message)
		if err != nil {
			fmt.Printf("Failed for %s: %s\n", text.Santa, err)
			failures++
			continue
		}

		now := time.Now().UTC()
		text.SentAt = &now
		if err := replaceDrawFile(*drawFile, d); err != nil {
			return fmt.Errorf("sent to %s, but could not record it, so a retry will send it again: %w", text.Santa, err)
		}
		fmt.Printf("Sent to %s (%g credits).\n", text.Santa, credits)
	}

	if failures > 0 {
		return fmt.Errorf("%d texts were not sent: run send again to retry them", failures)
	}
	return nil
}
