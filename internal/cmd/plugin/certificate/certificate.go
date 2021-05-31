/*
This file is part of Cloud Native PostgreSQL.

Copyright (C) 2019-2021 EnterpriseDB Corporation.
*/

// Package certificate implement the kubectl-cnp certificate command
package certificate

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/EnterpriseDB/cloud-native-postgresql/api/v1"
	"github.com/EnterpriseDB/cloud-native-postgresql/internal/cmd/plugin"
	"github.com/EnterpriseDB/cloud-native-postgresql/pkg/certs"
)

// Params are the required information to create an user secret
type Params struct {
	Name        string
	Namespace   string
	User        string
	ClusterName string
}

// Generate generates a Kubernetes secret suitable to allow certificate authentication
// for a PostgreSQL user
func Generate(ctx context.Context, params Params, dryRun bool, format plugin.OutputFormat) error {
	var secret corev1.Secret

	err := plugin.Client.Get(
		ctx,
		client.ObjectKey{Namespace: params.Namespace, Name: params.ClusterName + apiv1.CaSecretSuffix},
		&secret)
	if err != nil {
		return err
	}

	caPair, err := certs.ParseCASecret(&secret)
	if err != nil {
		return err
	}

	userPair, err := caPair.CreateAndSignPair(params.User, certs.CertTypeClient, nil)
	if err != nil {
		return err
	}

	userSecret := userPair.GenerateServerSecret(params.Namespace, params.Name)
	err = plugin.Print(userSecret, format)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	err = plugin.Client.Create(ctx, userSecret)
	if err != nil {
		return err
	}

	fmt.Printf("secret/%v created\n", userSecret.Name)
	return nil
}