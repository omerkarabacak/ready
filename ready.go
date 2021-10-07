package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type (
	task struct {
		Command   string `yaml:"command"`
		Directory string `yaml:"directory"`
		Name      string `yaml:"name"`
	}
	config struct {
		Tasks []task `yaml:"tasks"`
	}
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		err := installHook()
		if err != nil {
			log.Fatalf("Error installing hook: %v\n", err)
		}

		fmt.Println("Ready ready ✅")

		os.Exit(0)
	}

	cfg, err := newConfig().withYAML()
	if err != nil {
		log.Fatalf("Failed to get config: %v\n", err)
	}

	failures := 0

	for _, t := range cfg.Tasks {
		fmt.Printf("⏳ Running %q... ", t.Name)

		output, err := runTask(t)
		if err != nil {
			fmt.Printf("Failure ❌\n\n%v\n", err)

			failures++

			continue
		}

		if output == "" {
			fmt.Printf("Success ✅\n\n")
		} else {
			fmt.Printf("Success ✅\n\n%v\n", output)
		}
	}

	if failures > 0 {
		if failures == 1 {
			fmt.Println("Got a failure ⚠️  Please fix it and commit again")
		} else {
			fmt.Println("Got some failures ⚠️  Please fix them and commit again")
		}

		os.Exit(1)
	}
}

func installHook() error {
	hook := "./.git/hooks/pre-commit"
	_, err := os.Open(hook)
	if err == nil {
		fmt.Println("A pre-commit hook already exists ℹ️  Do you want to overwrite it? [yes/no]")

		res := ""
		fmt.Fscan(os.Stdin, &res)

		if res != "yes" {
			fmt.Println("Ready stopped 🛑")

			os.Exit(0)
		}
	}

	content := []byte(`
#!/bin/sh
# Hook created by Ready https://github.com/lewislbr/ready

initial_state=$(git diff --name-only)

ready

exit_status=$?
if [ $exit_status -ne 0 ]; then
	exit $exit_status
fi

latest_state=$(git diff --name-only)
if [[ $latest_state != $initial_state ]]; then
	echo "Some files have been modified by the hook. Please handle them and commit again"

	exit 1
fi

exit 0
`)
	err = os.WriteFile(hook, content, 0o755)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	return nil
}

func newConfig() *config {
	return &config{}
}

func (c *config) withYAML() (*config, error) {
	path, err := exec.Command("pwd").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("determining current path: %w", err)
	}

	file := strings.ReplaceAll(string(path), "\n", "") + "/ready.yaml"
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling data: %w", err)
	}

	return c, nil
}

func runTask(t task) (string, error) {
	parts := strings.Split(t.Command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)

	if t.Directory != "" {
		cmd.Dir = t.Directory
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if string(output) == "" {
			return "", err
		}

		return "", errors.New(string(output))
	}

	return string(output), nil
}