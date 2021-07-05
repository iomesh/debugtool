module github.com/iomesh/debugtool

go 1.15

replace k8s.io/client-go => k8s.io/client-go v0.20.2

require (
	github.com/briandowns/spinner v1.15.0
	github.com/enescakir/emoji v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/iomesh/operator v0.9.8 // indirect
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.20.2
	sigs.k8s.io/controller-runtime v0.6.2
)
