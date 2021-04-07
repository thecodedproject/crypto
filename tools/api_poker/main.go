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
	"encoding/json"

	"github.com/thecodedproject/crypto"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/factory"
	"github.com/thecodedproject/crypto/io"
)

var (
	authName = flag.String("api_auth", "", "API auth name to use")
	authPath = flag.String("auth_path", "api_auth.json", "Auth file path")
	pairName = flag.String("pair", "btcusdt", "Exchange pair to use")
	runGetCommand = flag.Bool("get", false, "Run get order command")
	runCancelCommand = flag.Bool("cancel", false, "Run cancel order command")
	runCustomCommand = flag.Bool("custom", false, "Run custom command")
)

type Command int

const (
	CommandUnknown Command = iota
	CommandGet
	CommandCancel
	CommandCustom
	CommandSentinal
)

func getCommand(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {

	if flag.NArg() != 1 {
		return fmt.Errorf("Need order ID for get command;\nUSAGE api_poker --get <order_id>")
	}

	orderStatus, err := exchangeClient.GetOrderStatus(
		ctx,
		flag.Arg(0),
	)
	if err != nil {
		return err
	}

	str, err := json.Marshal(orderStatus)
	if err != nil {
		return err
	}

	fmt.Println(string(str))

	return nil
}

func cancelCommand(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {

	if flag.NArg() != 1 {
		return fmt.Errorf("Need order ID for cancel command;\nUSAGE api_poker --cancel <order_id>")
	}

	orderID := flag.Arg(0)

	err := exchangeClient.CancelOrder(
		ctx,
		orderID,
	)
	if err != nil {
		return err
	}

	fmt.Println("Cancelled order", orderID)

	return nil
}

func postLimitOrder(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {

	orderID, err := exchangeClient.PostLimitOrder(
		ctx,
		exchangesdk.Order{
			Type: exchangesdk.OrderTypeBid,
			Price: decimal.NewFromFloat(50010),
			Volume: decimal.NewFromFloat(0.001),
		},
	)
	if err != nil {
		return err
	}

	fmt.Println(orderID)

	return nil
}

func postStopLimitOrder(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {

	orderID, err := exchangeClient.PostStopLimitOrder(
		ctx,
		exchangesdk.StopLimitOrder{
			Side: exchangesdk.OrderBookSideAsk,
			StopPrice: decimal.NewFromFloat(56855),
			LimitPrice: decimal.NewFromFloat(56850),
			Volume: decimal.NewFromFloat(0.0002),
		},
	)
	if err != nil {
		return err
	}

	fmt.Println(orderID)

	return nil
}

func runCustom(
	ctx context.Context,
	exchangeClient exchangesdk.Client,
) error {
	fmt.Println("run custom...")
	return postStopLimitOrder(ctx, exchangeClient)
}

func runCommand(
	ctx context.Context,
	command Command,
	exchangeClient exchangesdk.Client,
) error {

	switch command {
	case CommandGet:
		return getCommand(ctx, exchangeClient)
	case CommandCancel:
		return cancelCommand(ctx, exchangeClient)
	case CommandCustom:
		//reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter `custom` to confirm command run: ")
		//text, err := reader.ReadString('\n')
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "custom" {
			return fmt.Errorf("wrong confirm code - refusing to run custom command")
		}
		return runCustom(ctx, exchangeClient)
	default:
		return fmt.Errorf("Unknown command")
	}
}

func validateCmdArgs() {

	if *authName == "" {
		log.Fatal("api_auth is required")
	}
}

func resolveCommand() Command {

	if *runGetCommand {
		return CommandGet
	}
	if *runCancelCommand {
		return CommandCancel
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
