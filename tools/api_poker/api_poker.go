package main

// api_poker allows poking an exchange API using the
// exchange SDK.
// It was built as a way of exploring the way different
// exchange's APIs work.

import (
	"context"
	"log"
	"flag"
	"fmt"

	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/factory"
	"github.com/thecodedproject/crypto/io"
)

var (
	authName = flag.String("api_auth", "", "API auth name to use")
	authPath = flag.String("auth_path", "api_auth.json", "Auth file path")
	pairName = flag.String("pair", "btcusdt", "Exchange pair to use")
	runGetCommand = flag.Bool("get", false, "Run get command")
	runCustomCommand = flag.Bool("custom", false, "Run custom command")
)

type Command int

const (
	CommandUnknown Command = iota
	CommandGet
	CommandCustom
	CommandSentinal
)

func getCommand(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {

	return nil
}

func runCommand(
	ctx context.Context,
	command Command,
	exchangeClient exchangesdk.Client,
) error {

	switch command {
	case CommandGet:
		log.Println("Run get...")
		return nil
	case CommandCustom:
		//reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter `custom` to confirm command run: ")
		//text, err := reader.ReadString('\n')
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "custom" {
			return fmt.Errorf("wrong confirm code - refusing to run custom command")
		}
		log.Println("Run custom...")
		return nil
	default:
		return fmt.Errorf("Unknown command")
	}
}

func validateCmdArgs() {

	if *authName == "" {
		log.Fatal("auth_name is required")
	}
}

func resolveCommand() Command {

	if *runGetCommand {
		return CommandGet
	}
	if *runCustomCommand {
		return CommandCustom
	}

	log.Fatal("No command specified")
	return CommandUnknown
}


func main() {

	flag.Parse()
	validateCmdArgs()

	pair, err := crypto.PairString(*pairName)
	if err != nil {
		log.Fatal(err)
	}

	command := resolveCommand()

	apiCreds, err := io.GetAuthConfigByName(
		*authPath,
		*authName,
	)
	if err != nil {
		log.Fatal(err)
	}

	exchangeClient, err := factory.NewClient(
		crypto.Exchange{
			Provider: apiCreds.Provider,
			Pair: pair,
		},
		apiCreds.Key,
		apiCreds.Secret,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = runCommand(
		ctx,
		command,
		exchangeClient,
	)
	if err != nil {
		log.Fatal(err)
	}
}
