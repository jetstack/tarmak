// Copyright Jetstack Ltd. See LICENSE for details.
package server

import (
	"fmt"
	"io"
	"net"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"
	admissionmetrics "k8s.io/apiserver/pkg/admission/metrics"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/admission/plugin/instanceinittime"
	"github.com/jetstack/tarmak/pkg/wing/admission/winginitializer"
	"github.com/jetstack/tarmak/pkg/wing/apiserver"
	clientset "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
	informers "github.com/jetstack/tarmak/pkg/wing/client/informers/externalversions"
	"github.com/jetstack/tarmak/pkg/wing/tags"
)

const defaultEtcdPathPrefix = "/registry/wing.tarmak.io"

type WingServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions

	SharedInformerFactory informers.SharedInformerFactory
	StdOut                io.Writer
	StdErr                io.Writer
}

func NewWingServerOptions(out, errOut io.Writer) *WingServerOptions {
	o := &WingServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion),
			genericoptions.NewProcessInfo("wing-apiserver", "wing"),
		),

		StdOut: out,
		StdErr: errOut,
	}

	return o
}

// NewCommandStartWingServer provides a CLI handler for 'start master' command
// with a default WingServerOptions.
func NewCommandStartWingServer(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := NewWingServerOptions(out, errOut)

	cmd := &cobra.Command{
		Short: "Launch a wing API server",
		Long:  "Launch a wing API server",
		RunE: func(c *cobra.Command, args []string) error {
			t, err := tags.New(nil)
			if err != nil {
				return err
			}

			if err := t.EnsureMachineTags(); err != nil {
				return err
			}

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunWingServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.Etcd.AddFlags(flags)
	o.RecommendedOptions.SecureServing.AddFlags(flags)

	return cmd
}

func (o WingServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *WingServerOptions) Complete() error {

	// start with empty admission plugins
	o.RecommendedOptions.Admission = &genericoptions.AdmissionOptions{
		Plugins: admission.NewPlugins(),
	}

	// register admission plugins
	instaceinittime.Register(o.RecommendedOptions.Admission.Plugins)

	// add admisison plugins to the RecommendedPluginOrder
	o.RecommendedOptions.Admission.RecommendedPluginOrder = append(o.RecommendedOptions.Admission.RecommendedPluginOrder, instaceinittime.PluginName)

	return nil
}

func (o *WingServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	o.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
		client, err := clientset.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return nil, err
		}
		informerFactory := informers.NewSharedInformerFactory(client, c.LoopbackClientConfig.Timeout)
		o.SharedInformerFactory = informerFactory
		return []admission.PluginInitializer{winginitializer.New(informerFactory)}, nil
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.Etcd.ApplyTo(&serverConfig.Config); err != nil {
		return nil, err
	}
	if err := o.RecommendedOptions.SecureServing.ApplyTo(&serverConfig.SecureServing, &serverConfig.LoopbackClientConfig); err != nil {
		return nil, err
	}

	a := o.RecommendedOptions.Admission
	if initializers, err := o.RecommendedOptions.ExtraAdmissionInitializers(serverConfig); err != nil {
		return nil, err
	} else {
		pluginNames := a.RecommendedPluginOrder
		configScheme := runtime.NewScheme()

		pluginsConfigProvider, err := admission.ReadAdmissionConfiguration(pluginNames, a.ConfigFile, configScheme)
		if err != nil {
			return nil, fmt.Errorf("failed to read plugin config: %v", err)
		}

		initializersChain := admission.PluginInitializers{}
		initializersChain = append(initializersChain, initializers...)
		if admissionChain, err := a.Plugins.NewFromPlugins(pluginNames, pluginsConfigProvider, initializersChain, admission.DecoratorFunc(admissionmetrics.WithControllerMetrics)); err != nil {
			return nil, err
		} else {
			serverConfig.AdmissionControl = admissionmetrics.WithStepMetrics(admissionChain)
		}
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}

	return config, nil
}

func (o WingServerOptions) RunWingServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
