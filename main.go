package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

type Config struct {
	Allowed []string
}

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("JailSH is not made for windows. Sorry")
		os.Exit(1)
	}
	args := os.Args
	if len(args) > 1 {
		switch args[1] {
		case "jail":
			current, _ := user.Current()
			if current.Username != "root" {
				fmt.Println("You must be root to use jail")
				os.Exit(1)
			}
			if len(args) < 3 {
				fmt.Println("Please specify a user")
				os.Exit(1)
			}
			user := args[2]
			randId := uuid.New()
			file := "/tmp/jailsh-" + user + "-" + randId.String() + ".jailmsg"
			os.WriteFile(file, []byte("Please write a jail message (delete this line first)"), 0644)
			cmd := exec.Command("nano", file)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			MoveFile(file, "/home/"+user+"/jailmsg")
			cmd = exec.Command("chsh", "-s", "/bin/jailsh", user)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			fmt.Println("Done")
			os.Exit(0)
		case "unjail":
			current, _ := user.Current()
			if current.Username != "root" {
				fmt.Println("You must be root to use unjail")
				os.Exit(1)
			}
			if len(args) < 3 {
				fmt.Println("Please specify a user")
				os.Exit(1)
			}
			os.Remove("/home/" + args[2] + "/jailmsg")
			cmd := exec.Command("chsh", "-s", "/bin/bash", args[2])
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			fmt.Println("Done")
			os.Exit(0)
		}
	}
	config, err := os.ReadFile("/etc/jailsh/config.json")
	msg, err := os.ReadFile(os.Getenv("HOME") + "/jailmsg")
	var conf Config
	if err != nil {
		fmt.Println(error(err))
	}
	json.Unmarshal([]byte(config), &conf)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	println(string(msg))

	reader := bufio.NewReader(os.Stdin)
	for {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Print(wd)
		fmt.Print(" [jailsh]$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if input == "" {
			continue
		}
		allow := false
		for _, s := range conf.Allowed {
			if strings.HasPrefix(input, s) {
				allow = true
			}
		}
		if !allow {
			fmt.Println("Command not allowed")
			continue
		}
		if err = execInput(input, conf); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
func execInput(input string, conf Config) error {
	home := os.Getenv("HOME")
	input = strings.TrimSuffix(input, "\n")

	args := strings.Split(input, " ")

	switch args[0] {
	case "exit":
		os.Exit(0)
	case "cd":
		if len(args) < 2 {
			return os.Chdir(home)
		}
		return os.Chdir(args[1])
	case "printallowed":
		for i, s := range conf.Allowed {
			fmt.Println(i, s)
		}
		return nil
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
