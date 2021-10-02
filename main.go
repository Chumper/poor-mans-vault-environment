package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/vault/api"
)

const (
	defaultVaultAddr    string = "http://localhost:8200"
	defaultKeepSecrets  bool   = false
	defaultDebug        bool   = false
	vaultEnvironmentKey string = "vault:"
	// environment variables
	VAULT_ADDR        string = "VAULT_ADDR"
	VAULT_ROLE_ID     string = "VAULT_ROLE_ID"
	VAULT_SECRET_ID   string = "VAULT_SECRET_ID"
	VAULT_TOKEN       string = "VAULT_TOKEN"
	PMVE_KEEP_SECRETS string = "PMVE_KEEP_SECRETS"
	PMVE_DEBUG        string = "PMVE_DEBUG"
)

type Config struct {
	vaultAddr     string
	vaultRoleId   string
	vaultSecretId string
	vaultToken    string
	keepSecrets   bool
	debug         bool
	httpClient    *http.Client
}

func NewConfig() *Config {
	return &Config{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func main() {
	config := CreateConfig()
	client := CreateVaultClient(config)

	SetupEnvironment(config)
	ReplaceEnvironment(client, config.debug)

	ReplaceProcess()
}

func CreateConfig() *Config {
	config := NewConfig()
	ParseEnv(config)

	if len(config.vaultToken) > 0 {
		// token set, so use that
		return config
	}

	if len(config.vaultRoleId) == 0 {
		fmt.Printf("No Vault role ID found. Set it in the env as %s\n", VAULT_ROLE_ID)
		os.Exit(1)
	}

	if len(config.vaultSecretId) == 0 {
		fmt.Printf("No Vault secret ID found. Set it in the env as %s\n", VAULT_SECRET_ID)
		os.Exit(1)
	}

	return config
}

func CreateVaultClient(config *Config) *api.Client {
	client, err := api.NewClient(&api.Config{Address: config.vaultAddr, HttpClient: config.httpClient})
	if err != nil {
		fmt.Printf("Could not create a new Vault Client: %s\n", err)
		os.Exit(1)
	}

	if len(config.vaultToken) == 0 {
		resp, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
			"role_id":   config.vaultRoleId,
			"secret_id": config.vaultSecretId,
		})
		if err != nil {
			fmt.Printf("Login failed: %s\n", err)
			os.Exit(1)
		}
		client.SetToken(resp.Auth.ClientToken)
	} else {
		client.SetToken(config.vaultToken)
	}

	return client
}

func SetupEnvironment(config *Config) {
	if config.keepSecrets {
		os.Setenv(VAULT_ADDR, config.vaultAddr)
		os.Setenv(VAULT_ROLE_ID, config.vaultRoleId)
		os.Setenv(VAULT_SECRET_ID, config.vaultSecretId)
		os.Setenv(VAULT_TOKEN, config.vaultToken)
	} else {
		os.Unsetenv(VAULT_ADDR)
		os.Unsetenv(VAULT_ROLE_ID)
		os.Unsetenv(VAULT_SECRET_ID)
		os.Unsetenv(VAULT_TOKEN)
	}
}

func ReplaceEnvironment(client *api.Client, debug bool) {

	// loop over env vars and check for vault urn vault:/path/to/secret:field_to_read
	for _, element := range os.Environ() {
		env := strings.Split(element, "=")
		envKey, envVal := env[0], env[1]
		if strings.HasPrefix(envVal, vaultEnvironmentKey) {

			// read value in path
			envVal = strings.Replace(envVal, vaultEnvironmentKey, "", 1)
			pathSplit := strings.Split(envVal, "#")
			vaultPath, vaultKey := pathSplit[0], pathSplit[1]

			secret, err := client.Logical().Read(vaultPath)
			if err != nil {
				fmt.Printf("Could not read Vault path: %s\n", err)
				os.Exit(1)
			}
			if secret.Warnings != nil {
				fmt.Printf("Could not extract secret: %s\n", secret.Warnings[0])
				os.Exit(1)
			}

			content := secret.Data["data"].(map[string]interface{})
			if content == nil {
				fmt.Printf("Could not extract secret: Data not available\n")
				os.Exit(1)
			}
			secretValue, ok := content[vaultKey].(string)
			if ok {
				os.Setenv(envKey, secretValue)
				if debug {
					fmt.Printf("pmve: replacing vaule for %s\n", envKey)
				}
			} else {
				fmt.Printf("Could not decode secret: Field not found\n")
				os.Exit(1)
			}
		}
	}
}

func ParseEnv(config *Config) {
	config.vaultAddr = LookupEnvOrString(VAULT_ADDR, defaultVaultAddr)
	config.vaultToken = LookupEnvOrString(VAULT_TOKEN, "")
	config.vaultRoleId = LookupEnvOrString(VAULT_ROLE_ID, "")
	config.vaultSecretId = LookupEnvOrString(VAULT_SECRET_ID, "")
	config.keepSecrets = LookupEnvOrBool(PMVE_KEEP_SECRETS, defaultKeepSecrets)
	config.debug = LookupEnvOrBool(PMVE_DEBUG, defaultDebug)
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		bool_val, err := strconv.ParseBool(val)
		if err != nil {
			return defaultVal
		}
		return bool_val
	}
	return defaultVal
}

func ReplaceProcess() {
	flag.Parse()
	if len(os.Args) == 1 {
		return
	}
	cmd, err := exec.LookPath(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err := syscall.Exec(cmd, flag.Args(), os.Environ()); err != nil {
		panic(err)
	}
}
