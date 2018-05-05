/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

import (
	"fmt"
	"net"
)

// mock command
type MockCommand struct {
	c chan bool
}

func NewMockCommand() Command {
	return &MockCommand{c: make(chan bool)}
}

func (c *MockCommand) Run(cmds string, args []string) error {
	fmt.Println("mock.Run")

	timeout := NormalTimeout(1000 * 3600)
	select {
	case <-timeout.C:
		fmt.Println("mock.Run.Timeout[3seconds]")
	case <-c.c:
		fmt.Println("mock.Run.channel.closed")
	}
	return nil
}

func (c *MockCommand) Scan(substr string, times int) error {
	fmt.Println("mock.Scan")
	return nil
}

func (c *MockCommand) Kill() error {
	fmt.Println("mock.Kill")
	close(c.c)
	return nil
}

func (c *MockCommand) RunCommand(cmds string, args []string) (string, error) {
	return "ok", nil
}

func (c *MockCommand) RunCommandWithTimeout(to int, cmds string, args []string) (string, error) {
	return "", nil
}

// mock command
type MockACommand struct {
}

func NewMockACommand() Command {
	return &MockACommand{}
}

func (c *MockACommand) Run(cmds string, args []string) error {
	fmt.Println("mock.Run")
	return nil
}

func (c *MockACommand) Scan(substr string, times int) error {
	fmt.Println("mock.Scan")
	return nil
}

func (c *MockACommand) Kill() error {
	fmt.Println("mock.Kill")
	return nil
}

func (c *MockACommand) RunCommand(cmds string, args []string) (string, error) {
	fmt.Println("mock.RunCommand")
	return "ok", nil
}

func (c *MockACommand) RunCommandWithTimeout(to int, cmds string, args []string) (string, error) {
	return "", nil
}

// mock command
type MockBCommand struct {
}

func NewMockBCommand() Command {
	return &MockBCommand{}
}

func (c *MockBCommand) Run(cmds string, args []string) error {
	fmt.Println("mock.Run")
	return nil
}

func (c *MockBCommand) Scan(substr string, times int) error {
	fmt.Println("mock.Scan")
	return nil
}

func (c *MockBCommand) Kill() error {
	fmt.Println("mock.Kill")
	return nil
}

func (c *MockBCommand) RunCommand(cmds string, args []string) (string, error) {
	fmt.Println("mock.RunCommand")
	return "ok", fmt.Errorf("runcommmand.mock.error")
}

func (c *MockBCommand) RunCommandWithTimeout(to int, cmds string, args []string) (string, error) {
	return "", nil
}

// get local  ip for test only
func GetLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() {
					// ipv4
					if v.IP.To4() != nil {
						return v.IP.String(), nil
					}
				}
			}
		}
	}
	return "127.0.0.1", nil
}
