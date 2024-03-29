/*
Copyright 2021 The IOMesh Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/iomesh/debugtool/pkg/fixture"
)

var f = fixture.GetInstance()

var rootCmd = &cobra.Command{
	Use: "debug",
	Long: `IOMesh debugtool is used to detect whether the k8s environment meets
the installation conditions before installing IOMesh.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}
		return f.EnsureBasicDsDeployed()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := networkCmd.RunE(cmd, args); err != nil {
			return err
		}
		if err := infraCmd.RunE(cmd, args); err != nil {
			return err
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}
		return f.Cleanup()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
