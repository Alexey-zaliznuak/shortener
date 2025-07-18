package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type FlagsInitialConfig struct {
	StartupAddress      *string
	ShortLinksURLPrefix *string
}

type AppConfig struct {
	Port                int
	StartupAddress      string
	ShortLinksLength    int
	ShortLinksURLPrefix string
}

func CreateFLagsInitialConfig() *FlagsInitialConfig {
	return &FlagsInitialConfig{
		StartupAddress:      flag.String("a", "", "startup address"),
		ShortLinksURLPrefix: flag.String("b", "", "short links url prefix"),
	}
}

var GetConfig = func(flagsConfig *FlagsInitialConfig) *AppConfig {
	startupAddress := ""
	shortLinksURLPrefix := ""

	if flagsConfig.StartupAddress != nil {
		startupAddress = *flagsConfig.StartupAddress
	}

	if flagsConfig.ShortLinksURLPrefix != nil {
		shortLinksURLPrefix = *flagsConfig.ShortLinksURLPrefix
	}

	if startupAddress == "" {
		startupAddress = os.Getenv("SERVER_STARTUP_ADDRESS")
	}
	if startupAddress == "" {
		startupAddress = "localhost:8080"
		fmt.Printf("configuration warning: 'SERVER_STARTUP_ADDRESS' not specified, using default: %s\n", startupAddress)
	}

	if shortLinksURLPrefix == "" {
		shortLinksURLPrefix = os.Getenv("SHORT_LINKS_URL_PREFIX")
	}
	if shortLinksURLPrefix == "" {
		fmt.Println("configuration warning: 'SHORT_LINKS_URL_PREFIX' not specified")
	}

	shortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))

	if err != nil {
		shortLinksLength = 8
		fmt.Printf("configuration warning: 'SHORT_LINKS_LENGTH' not specified, using default: %d\n", shortLinksLength)
	}

	return &AppConfig{
		StartupAddress:      startupAddress,
		ShortLinksLength:    shortLinksLength,
		ShortLinksURLPrefix: shortLinksURLPrefix,
	}
}
