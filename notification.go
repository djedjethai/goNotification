package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func main() {
	// arg1 is the file name
	arg1, arg2, arg3, err := validateInput()
	// arg1, _, _, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	// build the path
	path, err := os.Getwd()
	if err != nil {
		exitGracefully(err, "Error finding the folder path")
	}
	p := filepath.Join(path, arg1)

	tokens := make(chan string)

	file, err := recorder("invalidTokens.txt")
	if err != nil {
		fmt.Println("err from file: ", err)
	}

	// parse csv file
	go readCsvFile(p, tokens)

	// read on the token channel and spawn the sendingNotification on 4 threats
	// dest return an array with the 4(from each thread) which returns the invalidTokens
	dest := split(tokens, arg2, arg3)

	var wg sync.WaitGroup
	wg.Add(len(dest))
	var mu sync.RWMutex

	// read the invalidTokens from the 4 returned channel
	for _, ch := range dest {
		go func(d <-chan string, f *os.File) {
			defer wg.Done()

			for val := range d {
				mu.Lock()
				fmt.Println(val, " failed")
				_, _ = fmt.Fprintf(f, "%s\n", val)
				mu.Unlock()
			}
		}(ch, file)
	}

	wg.Wait()

	color.Green("Finished successfuly!")
}

func split(source <-chan string, arg2, arg3 string) []<-chan string {
	dest := make([]<-chan string, 0)

	for i := 0; i < 5; i++ {
		ch := make(chan string)
		dest = append(dest, ch)
		go func() {
			defer close(ch)
			for val := range source {
				sendNotif(ch, val, arg2, arg3)
			}
		}()
	}
	return dest
}

func recorder(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("Can not open invalidToken file: %w", err)
	}
	return file, nil
}

func readCsvFile(filePath string, tokens chan string) {

	defer close(tokens)

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

	for _, v := range records {
		if len(v[18]) > 0 && len(v[18]) == 41 {
			// if len(v[18]) > 0 && (v[0] == "cleartoo_official" || v[0] == "guillaume27089") {
			tokens <- v[18]
		}
	}
}

func sendNotif(chRecord chan string, token, title, body string) {
	// chRecord <- token
	pushToken, err := expo.NewExponentPushToken(token)
	if err != nil {
		panic(err)
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, _ := client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{pushToken},
			Body:     body,
			Data:     map[string]string{"withSome": "data"},
			Sound:    "default",
			Title:    title,
			Priority: expo.DefaultPriority,
		},
	)

	if err != nil {
		color.Red("Error pushing message: ", err)
	}

	// Validate responses
	if response.ValidateResponse() != nil {
		chRecord <- fmt.Sprintf("%s", response.PushMessage.To)
	}
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
