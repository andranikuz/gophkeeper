package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/andranikuz/gophkeeper/internal/auth"
	"github.com/andranikuz/gophkeeper/internal/client"
)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client -server=<server_url> -grpc-server=<grpc-server_url> -db=<local_db_path> <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  register             -username=<username> -password=<password>")
	fmt.Println("  login                -username=<username> -password=<password>")
	fmt.Println("  get")
	fmt.Println("  save-credential      -login=<login> -password=<password>")
	fmt.Println("  save-text            -text='<text>'")
	fmt.Println("  save-card            -number=<card_number> -exp=<expiration_date> -cvv=<cvv> -holder=<card_holder_name>")
	fmt.Println("  save-file            -file=<file_path>")
	fmt.Println("  sync")
}

func main() {
	serverURL := flag.String("server", "http://127.0.0.1:8080", "Server URL")
	grpcServerURL := flag.String("grpc-server", "127.0.0.1:50051", "grpc server URL")
	dbPath := flag.String("db", "data/client.db", "Path to local BoltDB file")
	flag.Parse()
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}
	command := flag.Arg(0)
	// Создаем базовый контекст с таймаутом.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Открываем локальное хранилище BoltDB.
	localDB, err := client.OpenLocalStorage(*dbPath)
	if err != nil {
		fmt.Println("Error opening local database:", err)
		os.Exit(1)
	}
	defer localDB.Close()

	cli := client.NewClient(*serverURL, *grpcServerURL, auth.NewSession(), localDB)

	switch command {
	case "register":
		register(ctx, cli, flag.Args()[1:])
	case "login":
		login(ctx, cli, flag.Args()[1:])
	default:
		// Для остальных проверяем наличие сессии.
		if cli.Session.GetUserID() == "" {
			fmt.Println("No session found, please login first.")
			os.Exit(1)
		}
	}

	switch command {
	case "get":
		getItems(ctx, cli, flag.Args()[1:])
	case "save-credential":
		saveCredentials(ctx, cli, flag.Args()[1:])
	case "save-text":
		saveText(ctx, cli, flag.Args()[1:])
	case "save-card":
		saveCard(ctx, cli, flag.Args()[1:])
	case "save-file":
		saveFile(ctx, cli, flag.Args()[1:])
	case "sync":
		sync(ctx, cli, flag.Args()[1:])
	case "delete":
		delete(ctx, cli, flag.Args()[1:])
	default:
		if command != "register" && command != "login" {
			fmt.Println("Unknown command:", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func register(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("register", flag.ExitOnError)
	username := cmd.String("username", "", "Username")
	password := cmd.String("password", "", "Password")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *username == "" || *password == "" {
		fmt.Println("username and password must be provided")
		os.Exit(1)
	}
	dto := client.RegisterDTO{
		Username: *username,
		Password: *password,
	}
	if err := cli.Register(ctx, dto); err != nil {
		fmt.Println("Registration error:", err)
		os.Exit(1)
	}
	fmt.Println("Registration successful")
}

func login(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("login", flag.ExitOnError)
	username := cmd.String("username", "", "Username")
	password := cmd.String("password", "", "Password")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *username == "" || *password == "" {
		fmt.Println("username and password must be provided")
		os.Exit(1)
	}
	dto := client.LoginDTO{
		Username: *username,
		Password: *password,
	}
	err := cli.Login(ctx, dto)
	if err != nil {
		fmt.Println("Login error:", err)
		os.Exit(1)
	}
	fmt.Println("Login successful. Session saved.")
}

func saveText(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("save-text", flag.ExitOnError)
	text := cmd.String("text", "", "Text content")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *text == "" {
		fmt.Println("Text must be provided")
		os.Exit(1)
	}
	dto := client.TextDTO{
		Text: *text,
	}
	if err := cli.SaveText(ctx, dto); err != nil {
		fmt.Println("Save text error:", err)
		os.Exit(1)
	}
	fmt.Println("Text data saved successfully")
}

func saveCredentials(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("save-credential", flag.ExitOnError)
	login := cmd.String("login", "", "Credential login")
	password := cmd.String("password", "", "Credential password")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *login == "" || *password == "" {
		fmt.Println("Both login and password must be provided")
		os.Exit(1)
	}
	dto := client.CredentialDTO{
		Login:    *login,
		Password: *password,
	}
	if err := cli.SaveCredential(ctx, dto); err != nil {
		fmt.Println("Save credential error:", err)
		os.Exit(1)
	}
	fmt.Println("Credential data saved successfully")
}

func saveCard(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("save-card", flag.ExitOnError)
	cardNumber := cmd.String("number", "", "Card number (13-19 digits)")
	expiration := cmd.String("exp", "", "Expiration date (MM/YY or MM/YYYY)")
	cvv := cmd.String("cvv", "", "CVV (3-4 digits)")
	holder := cmd.String("holder", "", "Card holder name")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *cardNumber == "" || *expiration == "" || *cvv == "" || *holder == "" {
		fmt.Println("All card fields (-number, -exp, -cvv, -holder) must be provided")
		os.Exit(1)
	}
	dto := client.CardDTO{
		CardNumber:     *cardNumber,
		ExpirationDate: *expiration,
		CVV:            *cvv,
		CardHolderName: *holder,
	}
	if err := cli.SaveCard(ctx, dto); err != nil {
		fmt.Println("Save card error:", err)
		os.Exit(1)
	}
	fmt.Println("Card data saved successfully")
}

func saveFile(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("save-file", flag.ExitOnError)
	file := cmd.String("file", "", "Path to file")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *file == "" {
		fmt.Println("File path must be provided")
		os.Exit(1)
	}
	dto := client.FileDTO{
		FilePath: *file,
	}
	if err := cli.SaveFile(ctx, dto); err != nil {
		fmt.Println("Save file error:", err)
		os.Exit(1)
	}
	fmt.Println("File data saved successfully")
}

func getItems(ctx context.Context, cli *client.Client, args []string) {
	data, err := cli.GetItems(ctx)
	if err != nil {
		fmt.Println("Get data error:", err)
		os.Exit(1)
	}

	if len(data) == 0 {
		fmt.Println("No data found.")
		return
	}

	// Создаем tabwriter для красивого вывода таблицы.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// Заголовок таблицы.
	fmt.Fprintln(w, "ID\tType\tUpdated At\tContent")
	for _, item := range data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.ID,
			item.Type.String(),
			item.UpdatedAt.Format(time.RFC3339),
			string(item.Content),
		)
	}
	if err := w.Flush(); err != nil {
		fmt.Println("Failed to print data")
		os.Exit(1)
	}
}

func delete(ctx context.Context, cli *client.Client, args []string) {
	cmd := flag.NewFlagSet("delete", flag.ExitOnError)
	id := cmd.String("id", "", "item id")
	if err := cmd.Parse(args); err != nil {
		fmt.Println("Failed to parse arguments")
		os.Exit(1)
	}
	if *id == "" {
		fmt.Println("id must be provided")
		os.Exit(1)
	}
	if err := cli.DeleteItem(ctx, *id); err != nil {
		fmt.Println("Delete error:", err)
		os.Exit(1)
	}
	fmt.Println("Item deleted")
}

func sync(ctx context.Context, cli *client.Client, args []string) {
	if err := cli.SyncGRPC(ctx); err != nil {
		fmt.Println("Sync error:", err)
		os.Exit(1)
	}
	fmt.Println("Synchronization completed")
}
