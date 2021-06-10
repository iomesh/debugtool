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
package dns

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
	"github.com/iomesh/debugtool/pkg/kutils"
)

type DNSChecker struct {
	checker.Checker
}

func NewDNSChecker() *DNSChecker {
	return &DNSChecker{
		Checker: checker.Newchecker("DNSChecker"),
	}
}

func (dc DNSChecker) Check() error {
	dc.Spinner.Start()
	// create debug service
	service := kutils.NewService(constant.DebugNamespace, "iomesh-debug")
	service.Spec.Ports = []corev1.ServicePort{
		{
			Port: 5201,
		},
	}
	err := dc.Client.Create(context.TODO(), service)
	if err != nil {
		dc.SpinnerStop(emoji.CrossMark)
		return err
	}

	// check nslookup debug service
	podList := &corev1.PodList{}
	err = dc.Client.List(context.TODO(), podList, &client.ListOptions{
		Namespace: constant.DebugNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": constant.BasicCheckerLabel,
		}),
	})
	if err != nil {
		dc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("List basic checker pods: %v", err)
	}
	if len(podList.Items) < 2 {
		dc.SpinnerStop(emoji.CrossMark)
		return errors.New("Num of nodes less than 2")
	}

	checkDNSCmd := "host iomesh-debug"
	_, err = dc.RunCmdInPod(podList.Items[0].Name, constant.DebugNamespace, checkDNSCmd)
	if err != nil {
		dc.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("Can't resolute service iomesh-debug, DNS service not working")
	}
	dc.SpinnerStop(emoji.CheckMarkButton)
	return nil
}

func (dc DNSChecker) SpinnerStart() {
	dc.Spinner.Suffix = " Checking Coredns working"
	dc.Spinner.Start()
}

func (dc DNSChecker) SpinnerStop(emoji emoji.Emoji) {
	dc.Spinner.FinalMSG = fmt.Sprintf("%v Checking Coredns working\n", emoji)
	dc.Spinner.Stop()
}
