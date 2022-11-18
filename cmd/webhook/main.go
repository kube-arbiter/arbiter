/*
Copyright 2022 The Arbiter Authors.

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
	"flag"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	"github.com/kube-arbiter/arbiter/pkg/webhook/overcommit"
)

func main() {
	port := flag.Int("port", 9443, "Port to listen")
	tlsKeyFile := flag.String("tls-key", "/etc/webhook/certs/tls.key", "TLS key file path")
	tlsCertFile := flag.String("tls-cert", "/etc/webhook/certs/tls.crt", "TLS certificate file path")
	overcommitName := flag.String("overcommit-name", "overcommit", "overcommit config name")
	klog.InitFlags(flag.CommandLine)
	flag.Parse()

	overcommit.OvercommitName = *overcommitName
	addr := fmt.Sprintf(":%d", *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/overcommit/mutate", overcommit.Handle)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	klog.Infof("Listening on %s", addr)
	if err := server.ListenAndServeTLS(*tlsCertFile, *tlsKeyFile); err != nil {
		klog.Fatalf("Error listening: %v", err)
	}
}
