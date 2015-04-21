package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/routing-api"
	"github.com/cloudfoundry-incubator/routing-api-cli/commands"
	"github.com/cloudfoundry-incubator/routing-api/db"
	token_fetcher "github.com/cloudfoundry-incubator/uaa-token-fetcher"
	"github.com/codegangsta/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "api",
		Usage: "Endpoint for the routing-api. (required)",
	},
	cli.StringFlag{
		Name:  "oauth-name",
		Usage: "Name of the OAuth client. (required)",
	},
	cli.StringFlag{
		Name:  "oauth-password",
		Usage: "Password for OAuth client. (required)",
	},
	cli.StringFlag{
		Name:  "oauth-url",
		Usage: "URL for OAuth client. (required)",
	},
}

var cliCommands = []cli.Command{
	{
		Name:  "register",
		Usage: "Registers routes with the routing-api",
		Description: `Routes must be specified in JSON format, like so:
'[{"route":"foo.com", "port":12345, "ip":"1.2.3.4", "ttl":5, "log_guid":"log-guid"}]'`,
		Action: registerRoutes,
		Flags:  flags,
	},
	{
		Name:  "unregister",
		Usage: "Unregisters routes with the routing-api",
		Description: `Routes must be specified in JSON format, like so:
'[{"route":"foo.com", "port":12345, "ip":"1.2.3.4", "ttl":5, "log_guid":"log-guid"}]'`,
		Action: unregisterRoutes,
		Flags:  flags,
	},
}

func main() {
	fmt.Println()
	app := cli.NewApp()
	app.Name = "rtr"
	app.Usage = "A CLI for the Router API server."
	authors := []cli.Author{cli.Author{Name: "Cloud Foundry Runtime Team", Email: "vcap-dev@cloudfoundry.org"}}
	app.Authors = authors
	app.Commands = cliCommands
	app.CommandNotFound = commandNotFound

	app.Run(os.Args)
	os.Exit(0)
}

func registerRoutes(c *cli.Context) {
	checkFlagsAndArguments(c, "register")

	client := routing_api.NewClient(c.String("api"))

	config := buildOauthConfig(c)
	fetcher := token_fetcher.NewTokenFetcher(&config)

	desiredRoutes := c.Args().First()
	var routes []db.Route

	err := json.Unmarshal([]byte(desiredRoutes), &routes)
	if err != nil {
		fmt.Println("Invalid json format.")
		os.Exit(3)
	}

	err = commands.Register(client, fetcher, routes)
	if err != nil {
		fmt.Println("route registration failed:", err)
		os.Exit(3)
	}

	fmt.Printf("Successfuly registered routes: %s", desiredRoutes)
}

func unregisterRoutes(c *cli.Context) {
	checkFlagsAndArguments(c, "unregister")

	client := routing_api.NewClient(c.String("api"))

	config := buildOauthConfig(c)
	fetcher := token_fetcher.NewTokenFetcher(&config)

	desiredRoutes := c.Args().First()
	var routes []db.Route
	err := json.Unmarshal([]byte(desiredRoutes), &routes)
	if err != nil {
		fmt.Println("Invalid json format.")
		os.Exit(3)
	}

	err = commands.UnRegister(client, fetcher, routes)
	if err != nil {
		fmt.Println("route unregistration failed:", err)
		os.Exit(3)
	}

	fmt.Printf("Successfuly unregistered routes: %s", desiredRoutes)
}

func buildOauthConfig(c *cli.Context) token_fetcher.OAuthConfig {
	var port int
	oauthUrl, _ := url.Parse(c.String("oauth-url"))
	addr := strings.Split(oauthUrl.Host, ":")
	host := addr[0]

	if len(addr) > 1 {
		port, _ = strconv.Atoi(addr[1])
	} else {
		if strings.ToLower(oauthUrl.Scheme) == "https" {
			port = 443
		} else if strings.ToLower(oauthUrl.Scheme) == "http" {
			port = 80
		}
	}

	return token_fetcher.OAuthConfig{
		TokenEndpoint: oauthUrl.Scheme + "://" + host,
		ClientName:    c.String("oauth-name"),
		ClientSecret:  c.String("oauth-password"),
		Port:          port,
	}
}

func checkFlagsAndArguments(c *cli.Context, cmd string) {
	var issues []string

	if c.String("api") == "" {
		issues = append(issues, "Must provide an API endpoint for the routing-api component.")
	}

	if c.String("oauth-name") == "" {
		issues = append(issues, "Must provide the name of an OAuth client.")
	}

	if c.String("oauth-password") == "" {
		issues = append(issues, "Must provide an OAuth password/secret.")
	}

	if c.String("oauth-url") == "" {
		issues = append(issues, "Must provide an URL to the OAuth client.")
	}

	if !c.Args().Present() {
		issues = append(issues, "Must provide routes JSON.")
	}

	_, err := url.Parse(c.String("oauth-url"))
	if err != nil {
		issues = append(issues, "Invalid OAuth client URL")
	}

	if len(issues) > 0 {
		for _, issue := range issues {
			fmt.Println(issue)
		}
		fmt.Println()
		cli.ShowCommandHelp(c, cmd)
		os.Exit(1)
	}
}

func commandNotFound(c *cli.Context, cmd string) {
	fmt.Println("Not a valid command:", cmd)
	os.Exit(1)
}
