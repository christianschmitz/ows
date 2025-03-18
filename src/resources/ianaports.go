//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const outputPath = "./ianaports_gen.go"
const ianaURL = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv"
const portsPerLine = 10
const indent = "    "

func main() {
	resp, err := http.Get(ianaURL)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV (hint: retry)\n", err)
	}

	// reserved and assigned ports
	ports := newPortsCollection()

	for _, row := range records[1:] {
		rawPort := strings.TrimSpace(row[1])
		descr := strings.TrimSpace(row[3])

		if rawPort == "" || descr == "" || descr == "Unassigned" {
			continue
		}

		if strings.Contains(rawPort, "-") {
			fields := strings.Split(rawPort, "-")

			if len(fields) != 2 {
				log.Fatal(fmt.Sprintf("expected <start>-<end>, got %s", rawPort))
			}

			a := parsePort(fields[0])
			b := parsePort(fields[1])

			for p := a; p < b; p++ {
				ports.add(p)
			}
		} else {
			p := parsePort(row[1])

			ports.add(p)
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	prettyTime := time.Now().Format("2006/01/02")

	fmt.Fprintf(f, `package resources

// AUTO-GENERATED, DO NOT MODIFY
// taken from %s on %s
// %d out of 65536 ports are either Reserved or Assigned by IANA
var IANAPorts = []uint16{
%s
}`, ianaURL, prettyTime, ports.len(), ports.string(indent, portsPerLine))
}

func parsePort(s string) uint16 {
	p, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		log.Fatal(err)
	}

	return uint16(p)
}

type portsCollection struct {
	ports []uint16
}

func newPortsCollection() *portsCollection {
	return &portsCollection{[]uint16{}}
}

func (ps *portsCollection) add(p uint16) {
	if !slices.Contains(ps.ports, p) {
		ps.ports = append(ps.ports, p)
	}
}

func (ps *portsCollection) len() int {
	return len(ps.ports)
}

func (ps *portsCollection) string(indent string, fieldsPerRow int) string {
	lines := []string{}
	fields := []string{}

	flushFields := func() {
		lines = append(lines, fmt.Sprintf("%s%s,", indent, strings.Join(fields, ", ")))
		fields = []string{}
	}

	for _, p := range ps.ports {
		fields = append(fields, strconv.Itoa(int(p)))

		if len(fields) >= fieldsPerRow {
			flushFields()
		}
	}

	if len(fields) >= 0 {
		flushFields()
	}

	return strings.Join(lines, "\n")
}
