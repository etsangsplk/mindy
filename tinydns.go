package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type TinyDNSData map[string]*ResourceRecord

type ResourceRecord struct {
	Type    string `json:"type"`
	FQDN    string `json:"fqdn"`
	Address string `json:"address"`
}

func TinyDNSDataFromFile(file string) (TinyDNSData, error) {
	// read tinydns file
	b, err := ioutil.ReadFile(tinydnsDataFileFlag)
	if err != nil {
		return nil, err
	}

	tinydnsData := make(map[string]*ResourceRecord)
	dataLines := strings.Split(string(b), "\n")
	for _, line := range dataLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		first := string(line[0])
		fields := strings.Split(line[1:], ":")
		name, ip := fields[0], fields[1]
		var typ string
		switch first {
		case ".":
			typ = "NS"
		case "=", "+":
			typ = "A"
		case "#":
			continue
		default:
			return nil, fmt.Errorf("Unknown first character in tinydns data: %s", first)
		}
		tinydnsData[name] = &ResourceRecord{typ, name, ip}
	}
	return tinydnsData, nil
}
