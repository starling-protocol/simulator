package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/starling-protocol/simulator/pcap"

	"github.com/urfave/cli/v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type LastReceived struct {
	EncryptedData []byte
	Timestamp     string
	Source        string
}

func main() {
	var input string
	var output string
	var source string

	app := &cli.App{
		Name:  "logconverter",
		Usage: "convert Starling debug logs to pcap files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Usage:       "path to log file",
				Required:    true,
				Destination: &input,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "output.pcap",
				Usage:       "path to pcap output file",
				Destination: &output,
			},
			&cli.StringFlag{
				Name:        "device",
				Aliases:     []string{"d"},
				Value:       "00:00:00:00:00:00",
				Usage:       "Device MAC address",
				Destination: &source,
			},
		},
		Action: func(cCtx *cli.Context) error {
			return decode(input, output, source)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func decode(input string, output string, inputDevice string) error {

	pcapFile := pcap.NewPCAPFile(time.Now())

	deviceAddr := getAddress(inputDevice)

	data, err := os.ReadFile(input)
	check(err)

	str := fmt.Sprintf("%s", data)

	lines := strings.Split(str, "\n")

	if lines[0][0] == '[' {
		lines = appendTime(lines)
	}

	lastSentDecryptedData := ""

	var lastReceived *LastReceived = nil

	for _, line := range lines {
		if strings.Contains(line, "link:send:packets") {
			info := strings.Split(line, ":")
			timestamp := fmt.Sprintf("%s:%s:%s", info[0], info[1], info[2])
			dest := getAddress(info[6])
			encodedPaket := info[8]

			packet := craftPacket(encodedPaket, lastSentDecryptedData)

			pcapFile.AddPacket(packet, getTime(timestamp), dest, deviceAddr)
			lastSentDecryptedData = ""

		} else if strings.Contains(line, "network:send:sess:session:") {
			info := strings.Split(line, ":")
			decryptedData := info[9]
			if decryptedData == "" {
				panic("Empty string encrypted")
			}
			lastSentDecryptedData = decryptedData

		} else if strings.Contains(line, "proto:receive_packet") {
			info := strings.Split(line, ":")
			timestamp := fmt.Sprintf("%s:%s:%s", info[0], info[1], info[2])
			source := getAddress(info[5])
			encodedPacket := info[6]

			rawPacket, err := base64.StdEncoding.DecodeString(encodedPacket)
			if err != nil {
				panic("Error decoding base64 packet from log file")
			}

			if rawPacket[2] == 0x3 {
				lastReceived = &LastReceived{
					EncryptedData: rawPacket,
					Timestamp:     timestamp,
					Source:        source,
				}
			} else {
				pcapFile.AddPacket(rawPacket, getTime(timestamp), "000000000000", source)
			}
		} else if strings.Contains(line, "network:packet:sess:decrypted:") {
			info := strings.Split(line, ":")
			decryptedEncoded := info[8]

			decryptedData, err := base64.StdEncoding.DecodeString(decryptedEncoded)
			if err != nil {
				panic("Error decoding base64 packet from log file")
			}

			if lastReceived == nil {
				panic("Error: Log file incorrect or incomplete")
			}

			rawPacket := lastReceived.EncryptedData
			copy(rawPacket[27:], decryptedData)

			pcapFile.AddPacket(rawPacket, getTime(lastReceived.Timestamp), deviceAddr, lastReceived.Source)
			lastReceived = nil
		}

	}

	pcapFile.WriteFile(output)
	return nil
}

func getTime(timestamp string) time.Time {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		fmt.Println(err)
		panic("Could not parse time")
	}
	return t
}

func getAddress(address string) string {
	if len(address) > 17 {
		// Handle Iphone address
		return address[len(address)-12:]
	} else if len(address) == 17 {
		// Handle Android address
		return strings.ReplaceAll(address, ":", "")
	} else {
		panic("Error decoding address")
	}
}

func craftPacket(encodedPacket string, lastEncryptedData string) []byte {

	rawPacket, err := base64.StdEncoding.DecodeString(encodedPacket)
	if err != nil {
		panic("Error decoding base64 packet from log file")
	}
	header := []byte{0, 0}
	header[0] = (byte(len(rawPacket)>>8) & 0b11) | 0b10000000
	header[1] = byte(len(rawPacket) & 0xff)

	if lastEncryptedData != "" && rawPacket[0] == 0x3 {
		// Contains encrypted data. Replace with decrypted data for debugging

		decryptedData, err := base64.StdEncoding.DecodeString(lastEncryptedData)
		if err != nil {
			panic("Error decoding base64 decryptedData from log file")
		}
		copy(rawPacket[25:], decryptedData)
	}

	packet := append(header, rawPacket...)
	return packet
}

func appendTime(lines []string) []string {
	fmt.Printf("Attempting to fix file. Prepending timestamps\n")
	currentTime := time.Now()
	newLines := []string{}

	for i, line := range lines {
		newLines = append(newLines, "")
		if len(line) > 0 {
			if line[0] == '[' {
				newLines[i] = fmt.Sprintf("%s: %s", fmt.Sprintf(currentTime.Format("2006-01-02T15:04:05.000Z")), line)
				// fmt.Printf("%s: %s\n", fmt.Sprintf(currentTime.Format("2006-01-02T15:04:05.000Z")), line)
				currentTime = currentTime.Add(1 * time.Second)
			} else {
				panic("Error attempting to fix file")
			}
		}
	}
	return newLines
}
