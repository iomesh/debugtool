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
package checker

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Checker struct {
	Client  client.Client
	Log     logr.Logger
	Spinner *spinner.Spinner

	// using for run cmd in pod
	PodExecConfig *rest.Config
	ClientSet     *kubernetes.Clientset
}

func Newchecker(LoggerName string) Checker {
	checker := Checker{}
	log.SetLogger(zap.New())
	checker.Log = ctrl.Log.WithName(LoggerName)

	k8sClient, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		checker.Log.Error(err, "Connect to k8s cluster fail")
		os.Exit(1)
	}
	checker.Client = k8sClient

	checker.Spinner = spinner.New(spinner.CharSets[7], 100*time.Millisecond)

	// using for run cmd in pod
	checker.PodExecConfig = config.GetConfigOrDie()
	checker.PodExecConfig.GroupVersion = &schema.GroupVersion{
		Version: "v1",
	}
	checker.PodExecConfig.APIPath = "/api"
	checker.PodExecConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	checker.ClientSet, err = kubernetes.NewForConfig(checker.PodExecConfig)
	if err != nil {
		checker.Log.Error(err, "create clientset fail")
		os.Exit(1)
	}

	return checker
}

func (c Checker) RunCmdInPod(podName string, namespace string, cmd string) (string, error) {
	stdOutput := bytes.NewBuffer([]byte{})
	errOutput := bytes.NewBuffer([]byte{})
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				Out:    stdOutput,
				ErrOut: errOutput,
			},

			Namespace: namespace,
			PodName:   podName,
		},

		Command:   []string{"bash", "-c", cmd},
		Executor:  &exec.DefaultRemoteExecutor{},
		Config:    c.PodExecConfig,
		PodClient: c.ClientSet.CoreV1(),
	}

	if err := options.Validate(); err != nil {
		return "", fmt.Errorf("Validate ExecOptions: %v", err)
	}

	if err := options.Run(); err != nil {
		return "", fmt.Errorf("Run %s fail %v", cmd, err)
	}

	return stdOutput.String(), nil
}
