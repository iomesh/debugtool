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
package hostnetwork

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/enescakir/emoji"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/iomesh/debugtool/pkg/checker"
	"github.com/iomesh/debugtool/pkg/constant"
	"github.com/iomesh/debugtool/pkg/kutils"
)

type HostNetworkChecker struct {
	checker.Checker
}

func NewHostNetworkChecker() *HostNetworkChecker {
	checker := &HostNetworkChecker{
		Checker: checker.Newchecker("HostNetworkChecker"),
	}
	return checker
}

func (hc HostNetworkChecker) GetBandwidth() error {
	hc.SpinnerStart()

	ds, err := hc.HostNetworkCheckerDaemonSet(constant.DebugNamespace, constant.HostNetworkCheckerDSName)
	if err != nil {
		hc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("create basic checker daemonset fail: %v", err)
	}

	if err := hc.Client.Create(context.TODO(), ds); err != nil {
		hc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("create HostNetworkCheckerDaemonSet: %v", err)
	}
	if err := kutils.WaitDaemonSetReady(hc.Client, constant.DebugNamespace, constant.HostNetworkCheckerDSName); err != nil {
		hc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("HostNetworkCheckerDaemonSet create: %v", err)
	}
	podList := &corev1.PodList{}
	err = hc.Client.List(context.TODO(), podList, &client.ListOptions{
		Namespace: constant.DebugNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": constant.HostNetworkCheckerLabel,
		}),
	})
	if err != nil {
		hc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("List debugtool's: %v", err)
	}

	if len(podList.Items) < 2 {
		hc.SpinnerStop(emoji.CrossMark)
		return errors.New("Num of nodes less than 2")
	}
	results := []CheckResult{}
	for iperfClientIdx := 0; iperfClientIdx < len(podList.Items)-1; iperfClientIdx++ {
		for iperfServerIdx := iperfClientIdx + 1; iperfServerIdx < len(podList.Items); iperfServerIdx++ {
			clientPod := podList.Items[iperfClientIdx]
			serverPod := podList.Items[iperfServerIdx]

			getClientIperfListenIPCmd := "cat /opt/iperf_bind_addr"
			output, err := hc.RunCmdInPod(clientPod.Name, constant.DebugNamespace, getClientIperfListenIPCmd)
			if err != nil {
				hc.SpinnerStop(emoji.CrossMark)
				return fmt.Errorf("Get pod %s iperf ip: %v", clientPod.Name, err)
			}
			clientIperfIP := strings.TrimSpace(output)

			getServerIperfListenIPCmd := "cat /opt/iperf_bind_addr"
			output, err = hc.RunCmdInPod(serverPod.Name, constant.DebugNamespace, getServerIperfListenIPCmd)
			if err != nil {
				hc.SpinnerStop(emoji.CrossMark)
				return fmt.Errorf("Get pod %s iperf ip: %v", serverPod.Name, err)
			}
			serverIperfIP := strings.TrimSpace(output)

			doIperfCmd := fmt.Sprintf("iperf3 -c %s -t 5 | grep sender | awk '{print $7/8*1024}'", serverIperfIP)
			hc.Log.V(5).Info(doIperfCmd)
			output, err = hc.RunCmdInPod(clientPod.Name, constant.DebugNamespace, doIperfCmd)
			if err != nil {
				hc.SpinnerStop(emoji.CrossMark)
				return fmt.Errorf("Pod %s can't connect to Pod %s, check if HostNetwork is configured correctly", clientPod.Name, serverPod.Name)
			}
			bandwidth, _ := strconv.ParseFloat(strings.TrimSpace(output), 32)
			results = append(results, CheckResult{
				SourceIP:      clientIperfIP,
				DestinationIP: serverIperfIP,
				BandwidthMB:   float32(bandwidth),
			})
		}
	}
	hc.SpinnerStop(emoji.CheckMarkButton)
	for _, result := range results {
		result.Println()
	}
	return nil
}

func (hc HostNetworkChecker) HostNetworkCheckerDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	dataCIRD := os.Getenv("IOMESH_DATA_CIDR")
	_, _, err := net.ParseCIDR(dataCIRD)
	if err != nil {
		return nil, errors.New("IOMESH_DATA_CIDR is not the correct cidr format. example: IOMESH_DATA_CIDR=192.168.1.0/24")
	}
	ds := kutils.NewDaemonSet(namespace, name)
	labels := map[string]string{
		"app": constant.HostNetworkCheckerLabel,
	}
	ds.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	ds.Spec.Template.ObjectMeta.Labels = labels
	ds.Spec.Template.Spec.HostNetwork = true

	container := corev1.Container{
		Name:  constant.HostNetworkCheckerLabel,
		Image: constant.DebugToolsImage,
		Env: []corev1.EnvVar{
			{
				Name:  "DATA_CIDR",
				Value: dataCIRD,
			},
		},
	}
	ds.Spec.Template.Spec.Containers = append(ds.Spec.Template.Spec.Containers, container)

	return ds, nil
}

func (hc HostNetworkChecker) SpinnerStart() {
	hc.Spinner.Suffix = " Measuring hostnetwork bandwidth"
	hc.Spinner.Start()
}

func (hc HostNetworkChecker) SpinnerStop(emoji emoji.Emoji) {
	hc.Spinner.FinalMSG = fmt.Sprintf("%v Measuring hostnetwork bandwidth\n", emoji)
	hc.Spinner.Stop()
}
