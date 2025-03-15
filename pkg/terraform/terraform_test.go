package terraform

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

var localstackContainer *localstack.LocalStackContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	localstackContainer, err = localstack.Run(ctx, "localstack/localstack:1.4.0")
	defer func() {
		if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
			fmt.Printf("failed to terminate container: %s", err.Error())
		}
	}()
	if err != nil {
		fmt.Printf("failed to start container: %s", err.Error())
		return
	}

	code := m.Run()
	os.Exit(code)
}

func getEndpoint(ctx context.Context) (string, error) {
	mappedPort, err := localstackContainer.MappedPort(ctx, "4566/tcp")
	if err != nil {
		return "", err
	}
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = provider.Close()
	}()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%d", host, mappedPort.Int()), nil
}

func TestTerraform_CompareVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		expectErr error
	}{
		{
			name:      "success",
			version:   "1.8.1",
			expectErr: nil,
		},
		{
			name:      "mismatch version",
			version:   "1.11.0",
			expectErr: errTerraformMissmatchVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &Params{
				Version: "1.8.1",
				WorkDir: "./testdata",
			}
			tf := New(params)
			err := tf.Setup(ctx)
			require.NoError(t, err)
			defer tf.Cleanup(ctx)
			err = tf.CompareVersion(ctx, tt.version)
			assert.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestTerraform_Plan_Apply(t *testing.T) {
	ctx := context.Background()

	endpoint, err := getEndpoint(ctx)
	require.NoError(t, err)

	params := &Params{
		Version: "1.8.1",
		WorkDir: "./testdata",
	}
	tf := New(params)
	err = tf.Setup(ctx)
	require.NoError(t, err)
	defer tf.Cleanup(ctx)

	init, err := tf.Init(ctx, &InitParams{
		BackendConfig: map[string]string{
			"path": "./plan.tfstate",
		},
	})
	require.NoError(t, err)
	require.False(t, init.HasError)

	plan, err := tf.Plan(ctx, &PlanParams{
		Out: "./test.tfplan",
		Vars: []string{
			"endpoint=" + endpoint,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Plan: 1 to add, 0 to change, 0 to destroy.", plan.Result)

	apply, err := tf.Apply(ctx, &ApplyParams{
		PlanFilePath: "./test.tfplan",
	})
	require.NoError(t, err)
	require.Equal(t, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.", apply.Result)
}
