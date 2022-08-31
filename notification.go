package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func main() {
	// arg1 is the file name
	arg1, arg2, arg3, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	// build the path
	path, err := os.Getwd()
	if err != nil {
		exitGracefully(err, "Error finding the folder path")
	}
	p := filepath.Join(path, arg1)

	// parse csv file
	tokens := readCsvFile(p)

	// send the notification
	for _, v := range tokens {
		sendNotif(v, arg2, arg3)
	}

	color.Green("Finished successfuly!")
}

func readCsvFile(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		exitGracefully(err, "Error opening the file")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		exitGracefully(err, "Error parsing the file")
	}

	var tokens []string
	for _, v := range records {
		if len(v[18]) > 0 {
			tokens = append(tokens, v[18])
		}
	}

	return tokens
}

func validateInput() (string, string, string, error) {
	var arg1, arg2, arg3 string

	if len(os.Args) > 1 {

		arg1 = os.Args[1]

		if len(os.Args) >= 3 {
			arg2 = os.Args[2]
		}

		if len(os.Args) >= 4 {
			arg3 = os.Args[3]
		}

	} else {
		color.Red("Error: file name required")
		return "", "", "", errors.New("Enter a file name in the command")
	}

	return arg1, arg2, arg3, nil
}

func sendNotif(token, title, body string) {
	pushToken, err := expo.NewExponentPushToken(token)
	if err != nil {
		panic(err)
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{pushToken},
			Body:     body,
			Data:     map[string]string{"withSome": "data"},
			Sound:    "default",
			Title:    title,
			Priority: expo.DefaultPriority,
		},
	)

	// Check errors
	if err != nil {
		panic(err)
	}

	// Validate responses
	if response.ValidateResponse() != nil {
		fmt.Println(response.PushMessage.To, "failed")
	}
}

func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Yellow("Finished without completing!")
	}

	os.Exit(0)
}
