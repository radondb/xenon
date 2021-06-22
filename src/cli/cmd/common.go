/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"config"
	"model"
	"xbase/xlog"

	"github.com/spf13/cobra"
)

var (
	configPathFile = "config.path"
	log            = xlog.NewStdLog(xlog.Level(xlog.INFO))
)

func ErrorOK(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		log.Panic("%v", err)
	}
}

func RspOK(ret string) {
	if ret != model.OK {
		log.Panic("rsp[%v] != [OK]", ret)
	}
}

func GetConfig() (*config.Config, error) {
	fullPath := configPathFile

	// try to search in current dir
	// and try to in Abs dir if failed
	if _, err := os.Stat(fullPath); err != nil {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return nil, err
		}
		fullPath = fmt.Sprintf("%s/%s", dir, configPathFile)
	}

	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	conf, err := config.LoadConfig(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func SaveConfig(conf *config.Config) error {
	fullPath := configPathFile

	// try to search in current dir
	// and try to in Abs dir if failed
	if _, err := os.Stat(fullPath); err != nil {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		fullPath = fmt.Sprintf("%s/%s", dir, configPathFile)
	}

	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	if err := config.WriteConfig(strings.TrimSpace(string(data)), conf); err != nil {
		return err
	}

	return nil
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	_, err = root.ExecuteC()
	return buf.String(), err
}
