package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Host struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Port     uint16 `json:"port"`
}

type Tunnel struct {
	LocalHost  string `json:"local_host"`
	LocalPort  uint16 `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort uint16 `json:"remote_port"`
}

type Config struct {
	Hosts   map[string]Host   `json:"hosts"`
	Tunnels map[string]Tunnel `json:"tunnels"`
}

func main() {
	file, err := os.OpenFile("portman.json", os.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Failed to open config file %s", err)
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file %s", err)
	}

	for tunnelName := range config.Tunnels {
		go sshTunnel(&config, tunnelName)
	}

	for {
	}
}

func sshTunnel(config *Config, tunnelName string) {
	tunnel := config.Tunnels[tunnelName]

	host, found := config.Hosts[tunnel.RemoteHost]
	if !found {
		log.Printf("%s | Could not find host \"%s\".", tunnelName, tunnel.RemoteHost)
		return
	}

	port := fmt.Sprintf("%d", host.Port)
	remote := fmt.Sprintf("%d:%s:%d", tunnel.RemotePort, tunnel.LocalHost, tunnel.LocalPort)

	// Manpage: https://linuxcommand.org/lc3_man_pages/ssh1.
	arg := []string{
		"-N",

		"-l", host.Username,
		"-p", port,
		"-R", remote,
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=1",
		host.Host,
	}

	log.Printf("%s | Starting ssh subprocess with arguments \"%s\"", tunnelName, strings.Join(arg, " "))

	command := exec.Command("ssh", arg...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stdout

	err := command.Start()
	if err != nil {
		log.Printf("%s | Failed to run ssh command: %s", tunnelName, err)
		return
	}

	// TODO: Add proper error handling
	_ = command.Wait()

	log.Printf("%s | SSH subprocess exited with exit code %d", tunnelName, command.ProcessState.ExitCode())
	sshTunnel(config, tunnelName)
}
