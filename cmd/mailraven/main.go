package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "quickstart":
		if err := RunQuickstart(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "check-config":
		if err := RunCheckConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Config Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid.")
	case "serve":
		if err := RunServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("MailRaven v0.1.0-alpha")
		fmt.Println("Mobile-first email server")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("MailRaven - Mobile-first email server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mailraven quickstart    Run initial setup (DKIM, config, admin user)")
	fmt.Println("  mailraven serve         Start SMTP and API servers")
	fmt.Println("  mailraven version       Show version information")
	fmt.Println()
	fmt.Println("For more information, see: specs/001-mobile-email-server/quickstart.md")
}
