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

	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"

	"github.com/iomesh/debugtool/pkg/network/cni"
	"github.com/iomesh/debugtool/pkg/network/hostnetwork"
)

// networkCmd represents the network command
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Verify connectivity and bandwidth of cni and hostnetwork",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Network %v\n", emoji.ElectricPlug)
		cniChecker := cni.NewCNIChecker()
		if err := cniChecker.CheckConnectivity(); err != nil {
			return err
		}

		hostNetworkChecker := hostnetwork.NewHostNetworkChecker()
		if err := hostNetworkChecker.GetBandwidth(); err != nil {
			return err
		}
		fmt.Println("")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(networkCmd)
}
