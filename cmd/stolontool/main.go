/*
Copyright (C) 2018 Gravitational, Inc.

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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravitational/stolon-app/internal/stolontool/pkg/cluster"
	"github.com/gravitational/stolon-app/internal/stolontool/pkg/defaults"

	"github.com/fatih/color"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	clusterConfig cluster.Config
	ctx           context.Context

	envs = map[string]string{
		"ETCD_CERT":         "etcd-cert-file",
		"ETCD_KEY":          "etcd-key-file",
		"ETCD_CACERT":       "etcd-ca-file",
		"ETCD_ENDPOINTS":    "etcd-endpoints",
		"POSTGRES_PASSWORD": "postgres-password",
		"APP_VERSION":       "app-version",
		"RIG_CHANGESET":     "changeset",
		"NODE_NAME":         "nodename",
	}

	stolontoolCmd = &cobra.Command{
		Use:   "",
		Short: "PostgreSQL major versions upgrade tool for Stolon cluster",
		Run: func(ccmd *cobra.Command, args []string) {
			ccmd.HelpFunc()(ccmd, args)
		},
	}
)

func main() {
	if err := stolontoolCmd.Execute(); err != nil {
		log.Error(trace.DebugReport(err))
		printError(err)
		os.Exit(255)
	}
}
func init() {
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.KubeConfig, "kubeconfig", "",
		"Kubernetes client config file")
	stolontoolCmd.PersistentFlags().StringVarP(&clusterConfig.Namespace, "namespace", "n",
		defaults.Namespace, "Kubernetes namespace for Stolon application")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.KeepersPodSelector, "keepers-selector",
		defaults.KeepersPodSelector, "Label to select keeper pods")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.SentinelsPodSelector, "sentinels-selector",
		defaults.SentinelsPodSelector, "Label to select sentinel pods")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.EtcdEndpoints, "etcd-endpoints",
		defaults.EtcdEndpoints, "Etcd server endpoints(ENV variable 'ETCD_ENDPOINTS')")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.EtcdCertFile, "etcd-cert-file", "",
		"Path to TLS certificate for connecting to etcd(ENV variable 'ETCD_CERT')")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.EtcdKeyFile, "etcd-key-file", "",
		"Path to TLS key for connecting to etcd(ENV variable 'ETCD_KEY')")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.EtcdCAFile, "etcd-ca-file", "",
		"Path to TLS CA for connecting to etcd(ENV variable 'ETCD_CACERT')")
	stolontoolCmd.PersistentFlags().StringVar(&clusterConfig.Name, "cluster-name",
		defaults.ClusterName, "Stolon cluster name")

	bindFlagEnv(stolontoolCmd.PersistentFlags())

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.TODO())
	go func() {
		exitSignals := make(chan os.Signal, 1)
		signal.Notify(exitSignals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

		select {
		case sig := <-exitSignals:
			log.Infof("Caught signal: %v.", sig)
			cancel()
		}
	}()

}

// bindFlagEnv binds environment variables to command flags
func bindFlagEnv(flagSet *flag.FlagSet) {
	for env, flag := range envs {
		cmdFlag := flagSet.Lookup(flag)
		if cmdFlag != nil {
			if value := os.Getenv(env); value != "" {
				cmdFlag.Value.Set(value)
			}
		}
	}
}

// printError prints the error message to the console
func printError(err error) {
	color.Red("[ERROR]: %v\n", trace.UserMessage(err))
}