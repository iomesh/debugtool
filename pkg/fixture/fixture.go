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
package fixture

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/enescakir/emoji"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/iomesh/debugtool/pkg/checker"
	"github.com/iomesh/debugtool/pkg/constant"
	"github.com/iomesh/debugtool/pkg/kutils"
)

type Fixture struct {
	checker.Checker
}

var fixture Fixture
var once sync.Once

func GetInstance() Fixture {
	once.Do(func() {
		fixture = NewFixture("fixture")
	})
	return fixture
}

func NewFixture(LoggerName string) Fixture {
	return Fixture{
		Checker: checker.Newchecker(LoggerName),
	}
}

func (f Fixture) EnsureBasicDsDeployed() error {
	f.SpinnerStart()

	// create debug namespace if not exist
	debugNamespace := &corev1.Namespace{}
	debugNamespaceLookupKey := types.NamespacedName{
		Name: constant.DebugNamespace,
	}
	if err := f.Client.Get(context.TODO(), debugNamespaceLookupKey, debugNamespace); err != nil {
		if err := f.Client.Create(context.TODO(), kutils.NewNamespace(constant.DebugNamespace)); err != nil {
			f.SpinnerStop(emoji.CrossMark)
			return fmt.Errorf("Create debug namespace: %v", err)
		}
	}

	// create basic checker daemonset if not exist
	ds := &appsv1.DaemonSet{}
	dsLookupKey := types.NamespacedName{
		Name:      constant.BasicCheckerDSName,
		Namespace: constant.DebugNamespace,
	}
	if err := f.Client.Get(context.TODO(), dsLookupKey, ds); err == nil {
		f.SpinnerStop(emoji.CheckMarkButton)
		return nil
	}

	ds, err := f.BasicCheckerDaemonSet(constant.DebugNamespace, constant.BasicCheckerDSName)
	if err != nil {
		f.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("Create basic checker daemonset: %v", err)
	}

	if err := f.Client.Create(context.TODO(), ds); err != nil {
		f.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("Create basic checker daemonset: %v", err)
	}
	if err := kutils.WaitDaemonSetReady(f.Client, constant.DebugNamespace, constant.BasicCheckerDSName); err != nil {
		f.SpinnerStop(emoji.CrossMark)
		return fmt.Errorf("Wait basic checker daemonset ready %v", err)
	}

	f.SpinnerStop(emoji.CheckMarkButton)
	fmt.Println("")
	return nil
}

func (f Fixture) Cleanup() error {
	ns := &corev1.Namespace{}
	nsLookupKey := types.NamespacedName{
		Name: constant.DebugNamespace,
	}
	// delete debug ns if exist
	if err := f.Client.Get(context.TODO(), nsLookupKey, ns); err != nil {
		return nil
	}
	if err := f.Client.Delete(context.TODO(), ns); err != nil {
		return fmt.Errorf("Delete iomesh debug tool namespace: %v", err)
	}

	return nil
}

func (f Fixture) BasicCheckerDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	dataCIRD := os.Getenv("IOMESH_DATA_CIDR")
	_, _, err := net.ParseCIDR(dataCIRD)
	if err != nil {
		return nil, errors.New("IOMESH_DATA_CIDR is not the correct cidr format. example: IOMESH_DATA_CIDR=192.168.1.0/24")
	}
	ds := kutils.NewDaemonSet(namespace, name)
	labels := map[string]string{
		"app": constant.BasicCheckerLabel,
	}
	ds.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	ds.Spec.Template.ObjectMeta.Labels = labels

	container := corev1.Container{
		Name:  constant.BasicCheckerLabel,
		Image: constant.DebugToolsImage,
		Env: []corev1.EnvVar{
			{
				Name:  "DATA_CIDR",
				Value: "",
			},
		},
	}
	ds.Spec.Template.Spec.Containers = append(ds.Spec.Template.Spec.Containers, container)

	return ds, nil
}

func (f Fixture) SpinnerStart() {
	f.Spinner.Suffix = " Preparing debug environment"
	f.Spinner.Start()
}

func (f Fixture) SpinnerStop(emoji emoji.Emoji) {
	f.Spinner.FinalMSG = fmt.Sprintf("%v Preparing debug environment\n", emoji)
	f.Spinner.Stop()
}
