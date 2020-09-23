// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/log"
)

func lxknsservice(cmd *cobra.Command, _ []string) error {
	if silent, _ := cmd.PersistentFlags().GetBool("silent"); silent {
		log.SetLevel(log.ErrorLevel)
	}
	if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("debug logging enabled")
	}
	// And now for the real meat.
	log.Infof("this is the lxkns Linux-kernel namespaces discovery service version %s", lxkns.SemVersion)
	log.Infof("https://github.com/thediveo/lxkns")
	addr, _ := cmd.PersistentFlags().GetString("http")
	if _, err := startServer(addr); err != nil {
		log.Errorf("cannot start service, error: %s", err.Error())
		os.Exit(1)
	}
	stopit := make(chan os.Signal, 1)
	signal.Notify(stopit, syscall.SIGINT)
	signal.Notify(stopit, syscall.SIGTERM)
	signal.Notify(stopit, syscall.SIGQUIT)
	<-stopit
	maxwait, _ := cmd.PersistentFlags().GetDuration("shutdown")
	stopServer(maxwait)
	return nil
}

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lxkns",
		Short:   "lxkns provides Linux-kernel namespace discovery as a service",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.BeforeCommand()
		},
		RunE: lxknsservice,
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().Bool("debug", false, "enables debugging output")
	rootCmd.PersistentFlags().Bool("silent", false, "silences everything below the error level")
	rootCmd.PersistentFlags().String("http", "[::]:5010", "HTTP service address")
	rootCmd.PersistentFlags().Duration("shutdown", 15*time.Second, "graceful shutdown duration limit")
	cli.AddFlags(rootCmd)
	return
}
