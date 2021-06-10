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
package cni

import (
	"context"
	"errors"
	"fmt"

	"github.com/enescakir/emoji"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/iomesh/debugtool/pkg/checker"
	"github.com/iomesh/debugtool/pkg/constant"
)

type CNIChecker struct {
	checker.Checker
}

func NewCNIChecker() *CNIChecker {
	return &CNIChecker{
		Checker: checker.Newchecker("CNIChecker"),
	}
}

func (cc CNIChecker) CheckConnectivity() error {
	cc.SpinnerStart()

	podList := &corev1.PodList{}
	err := cc.Client.List(context.TODO(), podList, &client.ListOptions{
		Namespace: constant.DebugNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": constant.BasicCheckerLabel,
		}),
	})
	if err != nil {
		cc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("List basic checker pods: %v", err)
	}

	if len(podList.Items) < 2 {
		cc.SpinnerStop(emoji.CrossMark)
		return errors.New("Num of nodes less than 2")
	}
	clientPod := podList.Items[0]
	for _, serverPod := range podList.Items {
		checkConnectivityCmd := fmt.Sprintf("nc -zv %s 5201", serverPod.Status.PodIP)
		_, err = cc.RunCmdInPod(clientPod.Name, constant.DebugNamespace, checkConnectivityCmd)
		if err != nil {
			cc.SpinnerStop(emoji.CrossMark)
			return fmt.Errorf("Pod %s can't connect to Pod %s, check if CNI is configured correctly", clientPod.Name, serverPod.Name)
		}
	}
	cc.SpinnerStop(emoji.CheckMarkButton)
	return nil
}

func (cc CNIChecker) GetBandwidth() string {
	return ""
}

func (cc CNIChecker) GetLatencyMS() int {
	return 0
}

func (cc CNIChecker) GetPacketLossRate() float32 {
	return 0.1
}

func (cc CNIChecker) SpinnerStart() {
	cc.Spinner.Suffix = " Checking CNI connectivity"
	cc.Spinner.Start()
}

func (cc CNIChecker) SpinnerStop(emoji emoji.Emoji) {
	cc.Spinner.FinalMSG = fmt.Sprintf("%v Checking CNI connectivity\n", emoji)
	cc.Spinner.Stop()
}
