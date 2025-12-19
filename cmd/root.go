/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/3stadt/vivialdi-url-profile-router/gui"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Configuration struct {
	Browser struct {
		Path           string `toml:"path" mapstructure:"path" validate:"file"`
		DefaultProfile string `toml:"default" mapstructure:"default" validate:"required"`
	} `toml:"browser"`
	ShowSelection []string `toml:"selectprofile" mapstructure:"selectprofile"`
	URLMapping    []struct {
		Name       string   `toml:"name" mapstructure:"name" validate:"required"`
		Folder     string   `toml:"folder" mapstructure:"folder" validate:"required"`
		TargetUrls []string `toml:"urls" mapstructure:"urls" validate:"min=1,dive,url"`
	} `toml:"mapping" mapstructure:"mapping"`
}

var config Configuration

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vivialdi-url-profile-router",
	Short: "rule-based opening of URLs in specific Vivialdi profiles",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: route,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file
func initConfig() {

	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read general config: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshall the config %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		log.Fatalf("Missing required attributes %v\n", err)
	}

	if err := checkDuplicates(&config); err != nil {
		log.Fatal(err)
	}
}

func route(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		log.Fatalln("No URL provided. Exiting.")
	}

	var err error = nil
	launchURL := ""
	routeHost := ""

	for _, arg := range args {
		launchURL, routeHost, err = sanitizeURL(arg)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatal("can't parse URL from arguments")
	}

	browserArgs := []string{}

	profile, showSelection := getProfile(routeHost)

	if showSelection {
		// used for human readable select options in the GUI
		profileSelection := map[string]string{
			"Default": config.Browser.DefaultProfile,
		}

		for _, p := range config.URLMapping {
			profileSelection[p.Name] = p.Folder
		}

		choosenProfile, cancel := gui.ChooseProfile(profile, profileSelection)
		if cancel {
			os.Exit(0)
		}

		if choosenProfile != "" {
			profile = &choosenProfile
		}
	}
	if profile != nil {
		browserArgs = append(browserArgs, fmt.Sprintf(`--profile-directory=%s`, *profile))
	} else {
		browserArgs = append(browserArgs, fmt.Sprintf(`--profile-directory=%s`, config.Browser.DefaultProfile))
	}

	browserArgs = append(browserArgs, launchURL)

	if err := startDetached(config.Browser.Path, browserArgs...); err != nil {
		log.Fatalf("error opening browser: %v\n", err)
	}

	os.Exit(0)
}

// checkDuplicates returns an error describing duplicates found.
func checkDuplicates(cfg *Configuration) error {
	// global seen map to detect duplicates across all mappings
	globalSeen := make(map[string]bool)
	// to collect occurrences (for nicer error message)
	type occ struct{ profile, url string }
	dups := make(map[string][]occ)

	for _, profile := range cfg.URLMapping {
		for _, u := range profile.TargetUrls {

			if globalSeen[u] {
				dups[u] = append(dups[u], occ{profile: profile.Name, url: u})
			} else {
				globalSeen[u] = true
			}

		}
	}

	if len(dups) == 0 {
		return nil
	}

	// Build error message
	msg := "duplicate URLs found:\n"
	for u, occs := range dups {
		msg += fmt.Sprintf("- %s appears in:\n", u)
		for _, o := range occs {
			msg += fmt.Sprintf("    - profile: %q\n", o.profile)
		}
	}
	return errors.New(msg)
}

// SanitizeURL parses and minimally validates a URL for routing purposes.
// It returns:
//   - launchURL: the cleaned URL string to pass to the browser
//   - routeHost: normalized host to use for profile lookup
func sanitizeURL(raw string) (launchURL string, routeHost string, err error) {
	if raw == "" {
		return "", "", errors.New("empty URL")
	}

	// Trim whitespace and common surrounding quotes
	clean := strings.TrimSpace(raw)
	clean = strings.Trim(clean, `"'`)

	// Basic length guard (avoid pathological input)
	if len(clean) > 64*1024 {
		return "", "", errors.New("URL too long")
	}

	u, err := url.Parse(clean)
	if err != nil {
		return "", "", err
	}

	// Require explicit scheme
	switch u.Scheme {
	case "http", "https":
		// allowed
	default:
		return "", "", errors.New("unsupported URL scheme")
	}

	// Host must be present
	if u.Host == "" {
		return "", "", errors.New("missing host")
	}

	// Extract host without credentials
	host := u.Host
	if strings.Contains(host, "@") {
		// strip userinfo if present
		if u.User != nil {
			host = strings.TrimPrefix(host, u.User.String()+"@")
		}
	}

	// Separate port if present
	if h, _, splitErr := net.SplitHostPort(host); splitErr == nil {
		host = h
	}

	// Normalize for routing only
	host = strings.ToLower(host)
	host = strings.TrimSuffix(host, ".")

	if host == "" {
		return "", "", errors.New("invalid host")
	}

	return clean, host, nil
}

// getProfile is used to check if a URL is defined in the config
// It returns:
//   - profileFolder: The profile to start the browser with, or nil if no profile should be used
//   - showSelection: If true, the user should be prompted which profile to use
func getProfile(checkURL string) (profileFolder *string, showSelection bool) {
	showSelection = false
	profile := ""

	for _, selectionURL := range config.ShowSelection {
		// get host from config URL - only host/domains are checked, no subfolders
		u, _ := url.Parse(selectionURL)
		host := u.Host
		host = strings.ToLower(host)
		host = strings.TrimSuffix(host, ".")
		if checkURL == host {
			showSelection = true
			return &profile, showSelection
		}
	}

	for _, um := range config.URLMapping {
		profile = um.Folder
		for _, targetURL := range um.TargetUrls {

			// get host from config URL - only host/domains are checked, no subfolders
			u, _ := url.Parse(targetURL)
			host := u.Host
			host = strings.ToLower(host)
			host = strings.TrimSuffix(host, ".")
			if checkURL == host {
				return &profile, showSelection
			}
		}
	}
	return nil, showSelection
}
