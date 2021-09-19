package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConfig(t *testing.T) {
	os.Setenv("VAULT_ROLE_ID", "123")
	os.Setenv("VAULT_SECRET_ID", "123")
	os.Setenv("PMVE_KEEP_SECRETS", "true")
	os.Setenv("VAULT_ADDR", "123")

	config := CreateConfig()

	assert.Equal(t, "123", config.vaultRoleId)
	assert.Equal(t, "123", config.vaultSecretId)
	assert.Equal(t, "123", config.vaultAddr)
	assert.Equal(t, true, config.keepSecrets)
}

func TestCreateConfigWithToken(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "123")
	os.Setenv("PMVE_KEEP_SECRETS", "true")
	os.Setenv("VAULT_ADDR", "123")

	config := CreateConfig()

	assert.Empty(t, config.vaultRoleId)
	assert.Empty(t, config.vaultSecretId)
	assert.Equal(t, "123", config.vaultAddr)
	assert.Equal(t, "123", config.vaultToken)
	assert.Equal(t, true, config.keepSecrets)
}

func TestSetupEnvironment(t *testing.T) {
	config := &Config{
		keepSecrets:   true,
		vaultAddr:     "123",
		vaultRoleId:   "123",
		vaultSecretId: "123",
		vaultToken:    "123",
	}

	SetupEnvironment(config)

	assert.Equal(t, "123", os.Getenv(VAULT_ADDR))
	assert.Equal(t, "123", os.Getenv(VAULT_ROLE_ID))
	assert.Equal(t, "123", os.Getenv(VAULT_SECRET_ID))
	assert.Equal(t, "123", os.Getenv(VAULT_TOKEN))

	config.keepSecrets = false
	SetupEnvironment(config)

	assert.Empty(t, os.Getenv(VAULT_ADDR))
	assert.Empty(t, os.Getenv(VAULT_ROLE_ID))
	assert.Empty(t, os.Getenv(VAULT_SECRET_ID))
	assert.Empty(t, os.Getenv(VAULT_TOKEN))
}

// ERROR...
// # chumper.github.com/poor-mans-vault-environment.test
// github.com/sethvargo/go-limiter/memorystore.(*store).purge: relocation target runtime.walltime not defined
// github.com/sethvargo/go-limiter/memorystore.newBucket: relocation target runtime.walltime not defined
// github.com/sethvargo/go-limiter/memorystore.(*bucket).take: relocation target runtime.walltime not defined
// FAIL	chumper.github.com/poor-mans-vault-environment [build failed]
// FAIL

// func TestReplaceEnvironment(t *testing.T) {
// 	cluster, client := createVault(t)
// 	defer cluster.Cleanup()

// 	// Set up environment
// 	os.Setenv("TEST_SEC", "vault:/secret/foo#secret")
// 	ReplaceEnvironment(client)
// 	if os.Getenv("TEST_SEC") != "bar" {
// 		t.Fatalf("Secret TEST_SEC not replaced: %s", os.Getenv("TEST_SEC"))
// 	}
// }

// func createVault(t *testing.T) (*vault.TestCluster, *api.Client) {
// 	cluster := vault.NewTestCluster(t, &vault.CoreConfig{
// 		DevToken: "root",
// 	}, &vault.TestClusterOptions{
// 		HandlerFunc: vaulthttp.Handler,
// 	})
// 	cluster.Start()

// 	core := cluster.Cores[0].Core
// 	vault.TestWaitActive(t, core)
// 	client := cluster.Cores[0].Client

// 	return cluster, client

// 	// err := putSecret(client, map[string]interface{}{"foo": "bar"}, "secret")
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// data, err := client.Logical().Read("secret/data/secret")
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// if secret, ok := data.Data["foo"].(string); ok {
// 	// 	if secret != "bar" {
// 	// 		t.Fatalf("Wrong secret returned: %s", secret)
// 	// 	}
// 	// } else {
// 	// 	t.Fatal("Could not get secret")
// 	// }
// }
