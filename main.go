package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishnusomank/policy-cli-2.0/pkg/discover_op"
	"github.com/vishnusomank/policy-cli-2.0/pkg/git_op"
	"github.com/vishnusomank/policy-cli-2.0/resources"
)

func banner() {
	fmt.Println()
	fmt.Println()
	fmt.Printf(strings.TrimSuffix(figure.NewFigure("AutoPolicy", "slant", true).String(), "\n") + "   v1.0.0")
	fmt.Println()
	fmt.Println()

}

func removeResidues(repo_path string) {

	err := os.RemoveAll(repo_path)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	// logging function generating following output
	// log.Info("") --> {"level":"info","msg":"","time":"2022-03-17T14:51:30+05:30"}
	// log.Warn("") --> {"level":"warning","msg":"","time":"2022-03-17T14:51:30+05:30"}
	// log.Error("") -- {"level":"error","msg":"","time":"2022-03-17T14:51:30+05:30"}

	log.SetFormatter(&log.JSONFormatter{})

	log_file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)

	// to get the current working directory
	resources.CURRENT_DIR, err = os.Getwd()
	if err != nil {
		log.Error(err)
	}

	// adding policy-template directory to current working directory
	resources.GIT_DIR = resources.CURRENT_DIR + "/policy-template"

	resources.AD_DIR = resources.CURRENT_DIR + "/ad-policy"

	log.Info("Current Working directory: " + resources.CURRENT_DIR)
	log.Info("Github clone directory: " + resources.GIT_DIR)

	myFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "git-username",
			Aliases:     []string{"git-user"},
			Usage:       "GitHub username",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "git-repo-url",
			Aliases:     []string{"git-url"},
			Usage:       "GitHub URL to push the updates",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "git-token",
			Aliases:     []string{"token"},
			Usage:       "GitHub token for authentication",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "git-branch-name",
			Aliases:     []string{"branch"},
			Usage:       "GitHub branch name for pushing updates",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "git-base-branch",
			Aliases:     []string{"basebranch"},
			Usage:       "GitHub base branch name for PR creation",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "action-value",
			Aliases:     []string{"action"},
			Usage:       "Action value for policy. Can be Audit, Block, Allow or no-change",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "Audit",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "exclude-namespace",
			Aliases:     []string{"exclude-ns"},
			Usage:       "Option to exclude generation of policies on certain namespaces",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.StringFlag{
			Name:        "only-on-namespace",
			Aliases:     []string{"only-on-ns"},
			Usage:       "Option to generation of policies only on certain namespaces",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "",
			DefaultText: "",
			Destination: new(string),
			HasBeenSet:  false,
		},
		&cli.BoolFlag{
			Name:        "auto-apply",
			Aliases:     []string{"auto"},
			Usage:       "If true, modifed YAML will be applied to the cluster",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Value:       false,
			DefaultText: "",
			Destination: new(bool),
			HasBeenSet:  false,
		},
		&cli.BoolFlag{
			Name:        "generate-locally",
			Aliases:     []string{"gen-loc"},
			Usage:       "If true, Policy YAML will only be generate locally under $(pwd)/accuknox-client-repo/",
			EnvVars:     []string{},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Value:       false,
			DefaultText: "",
			Destination: new(bool),
			HasBeenSet:  false,
		},
	}
	app := &cli.App{
		Name:      "Auto Policy",
		Usage:     "A simple CLI tool automates the creation of YAML-based runtime network & system security policies on top of Auto-Discovery feature by AccuKnox and Policy Templates",
		Version:   resources.CLI_VERSION,
		UsageText: "autopolicy [Flags]\n\n1. Generate policies locally\t-->  autopolicy --generate-locally --action=Audit --exclude-ns=kube-system\n2. Generate and push to GitHub\t-->  autopolicy --git_base_branch=deploy-branch --git-branch-name=temp-branch --git-token=gh-token123 --git-repo-url= https://github.com/testuser/demo.git --git-username=testuser",
		Flags:     myFlags,
		Action: func(c *cli.Context) error {
			resources.AUTOAPPLY = c.Bool("auto-apply")
			resources.GEN_LOC = c.Bool("generate-locally")
			if c.String("exclude-namespace") != "" {
				resources.EXCLUDE_NS = c.String("exclude-namespace")
			} else if c.String("exclude-ns") != "" {
				resources.EXCLUDE_NS = c.String("exclude-ns")
			}
			if c.String("only-on-ns") != "" {
				resources.INCLUDE_NS = c.String("only-on-ns")
			} else if c.String("only-on-namespace") != "" {
				resources.INCLUDE_NS = c.String("only-on-namespace")
			}
			if c.String("action") != "" {
				resources.ACTION_VAL = c.String("action")
			} else if c.String("action-value") != "" {
				resources.ACTION_VAL = c.String("action-value")
			} else {
				resources.ACTION_VAL = "Audit"
			}
			if resources.ACTION_VAL != "no-change" {
				resources.ACTION_VAL = strings.Title(resources.ACTION_VAL)
			}
			if resources.INCLUDE_NS != "" && resources.EXCLUDE_NS != "" {
				banner()
				fmt.Printf("[%s][%s] Please select only one option exclude-namespace or only-on-namespace.\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.CyanString("WARN"))

			} else {
				if resources.GEN_LOC == false {

					if c.String("git-username") == "" || c.String("git-token") == "" || c.String("git-repo-url") == "" || c.String("git-branch-name") == "" || c.String("git-base-branch") == "" {
						banner()
						fmt.Printf("[%s][%s] Parameters missing.\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.CyanString("WARN"))
						fmt.Printf("[%s][%s] Please use autopolicy --help for help\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.CyanString("WARN"))

					} else {
						startOperation(c.String("git-username"), c.String("git-token"), c.String("git-repo-url"), c.String("git-branch-name"), c.String("git-base-branch"))
					}

				} else {
					startOperation(c.String("git-username"), c.String("git-token"), c.String("git-repo-url"), c.String("git-branch-name"), c.String("git-base-branch"))

				}
			}
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}

}

func startOperation(git_uname, git_token, git_rep_url, git_branch_name, git_base_branch string) {
	banner()
	fmt.Printf("[%s][%s] Uses KubeConfig file to connect to cluster.\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.CyanString("WARN"))
	fmt.Printf("[%s][%s] Creates files and folders in current directory.\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.CyanString("WARN"))
	fileUrl := "https://raw.githubusercontent.com/accuknox/samples/main/discover/install.sh"
	discoverFileUrl := "https://raw.githubusercontent.com/accuknox/samples/main/discover/get_discovered_yamls.sh"
	git_op.Git_Operation(resources.GIT_DIR)
	discover_op.Auto_Discover(fileUrl, discoverFileUrl, resources.AD_DIR, resources.CURRENT_DIR)
	resources.REPO_PATH = resources.CURRENT_DIR + resources.REPO_PATH
	log.Info("repo_path=" + resources.REPO_PATH)
	git_op.Init_Git(git_uname, git_token, git_rep_url, git_branch_name, git_base_branch, resources.REPO_PATH, resources.AD_DIR, resources.CURRENT_DIR)

	removeResidues(resources.GIT_DIR)
	removeResidues(resources.AD_DIR)

	//removeResidues(resources.CURRENT_DIR + "/logs.log")
	if resources.GEN_LOC {
		fmt.Printf("[%s][%s] Policies generated and stored locally. Please navigate to %s\n", color.BlueString(time.Now().Format("01-02-2006 15:04:05")), color.GreenString("DONE"), color.CyanString(resources.REPO_PATH))
	}

}
