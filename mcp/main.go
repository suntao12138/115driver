// Copyright 2025 The 115driver Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/SheltonZhu/115driver/mcp/server"
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/spf13/viper"
)

var (
	cookie    = flag.String("cookie", "", "115 driver cookie for authentication")
	profile   = flag.String("profile", "", "Config profile name (default 'main')")
	configDir = flag.String("config", "", "Config file path (default ~/.115driver/config.toml)")
	help      = flag.Bool("help", false, "display help information")
)

func main() {
	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "115 Driver MCP Server - Provides access to 115 cloud storage via MCP protocol\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  With explicit cookie:   %s --cookie=\"UID=xxx;CID=xxx;SEID=xxx\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  From config file:       %s --profile main\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nThe --cookie flag can be omitted if the config file (~/.115driver/config.toml)\n")
		fmt.Fprintf(os.Stderr, "has a valid cookie under the specified profile. Run '115driver login' first.\n")
		os.Exit(1)
	}

	// Resolve cookie: --cookie flag > config file
	cookieStr := *cookie
	if cookieStr == "" {
		cookieStr = readConfigValue(*configDir, *profile, "cookie")
	}

	if cookieStr == "" {
		fmt.Fprintf(os.Stderr, "Error: Cookie is required. Provide via --cookie or configure in ~/.115driver/config.toml (run '115driver login' first).\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --cookie=\"UID=xxx;CID=xxx;SEID=xxx\" [--profile main] [--config ~/.115driver/config.toml]\n", os.Args[0])
		os.Exit(1)
	}

	cr := &driver.Credential{}
	cr.FromCookie(cookieStr)
	// Create 115 driver client and authenticate
	client := driver.New(driver.UA(driver.UA115Browser)).ImportCredential(cr)

	// Check login status
	if err := client.CookieCheck(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Read config values
	defaultSaveDir := readConfigValue(*configDir, *profile, "default_offline_save_dir")

	// Create and start the MCP server
	s := server.NewServer().WithClient(client).WithDefaultSaveDir(defaultSaveDir)
	if err := s.Start(context.Background()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// readConfigValue reads a config value from the config file for a given profile.
// Returns empty string if the key is not set, config file doesn't exist, or profile is not found.
func readConfigValue(configPath, profile, key string) string {
	path := configPath
	if path == "" {
		if envPath := os.Getenv("DRIVER115_CONFIG"); envPath != "" {
			path = envPath
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			path = filepath.Join(home, ".115driver", "config.toml")
		}
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return ""
	}

	prof := profile
	if prof == "" {
		if envProfile := os.Getenv("DRIVER115_PROFILE"); envProfile != "" {
			prof = envProfile
		}
	}
	if prof == "" {
		prof = v.GetString("default_profile")
	}
	if prof == "" {
		prof = "main"
	}

	return v.GetString("profiles." + prof + "." + key)
}
