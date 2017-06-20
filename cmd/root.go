package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appConfig *viper.Viper
var cfgFile string

const cfgMode = "mode"
const cfgModeValueAWSKMSSSM = "aws-kms-ssm"
const cfgModeValueGoogleCloudKMSGCS = "google-cloud-kms-gcs"

const cfgGoogleCloudKMSProject = "google-cloud-kms-project"
const cfgGoogleCloudKMSLocation = "google-cloud-kms-location"
const cfgGoogleCloudKMSKeyRing = "google-cloud-kms-key-ring"
const cfgGoogleCloudKMSCryptoKey = "google-cloud-kms-crypto-key"

const cfgGoogleCloudStorageBucket = "google-cloud-storage-bucket"
const cfgGoogleCloudStoragePrefix = "google-cloud-storage-prefix"

const cfgAWSKMSKeyID = "aws-kms-key-id"
const cfgAWSSSMKeyPrefix = "aws-ssm-key-prefix"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vault-helper",
	Short: "Automates PKI tasks using Hashicorp's Vault as a backend.",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func configStringVar(key, defaultValue, description string) {
	RootCmd.PersistentFlags().String(key, defaultValue, description)
	appConfig.BindPFlag(key, RootCmd.PersistentFlags().Lookup(key))
}

func init() {
	appConfig = viper.New()
	appConfig.SetEnvPrefix("vault")
	appConfig.AutomaticEnv()

	// SelectMode
	configStringVar(
		cfgMode,
		cfgModeValueGoogleCloudKMSGCS,
		fmt.Sprintf("Select the mode to use '%s' => Google Cloud Storage with encryption using Google KMS; '%s' => AWS SSM parameter store using AWS KMS encryption", cfgModeValueGoogleCloudKMSGCS, cfgModeValueAWSKMSSSM),
	)

	// Google Cloud KMS flags
	configStringVar(cfgGoogleCloudKMSProject, "", "The Google Cloud KMS project to use")
	configStringVar(cfgGoogleCloudKMSLocation, "", "The Google Cloud KMS location to use (eg. 'global', 'europe-west1')")
	configStringVar(cfgGoogleCloudKMSKeyRing, "", "The name of the Google Cloud KMS key ring to use")
	configStringVar(cfgGoogleCloudKMSCryptoKey, "", "The name of the Google Cloud KMS crypt key to use")

	// Google Cloud Storage flags
	configStringVar(cfgGoogleCloudStorageBucket, "", "The name of the Google Cloud Storage bucket to store values in")
	configStringVar(cfgGoogleCloudStoragePrefix, "", "The prefix to use for values store in Google Cloud Storage")

	// AWS KMS Storage flags
	configStringVar("aws-kms-key-id", "", "The ID or ARN of the AWS KMS key to encrypt values")

	// AWS SSM Parameter Storage flags
	configStringVar("aws-ssm-key-prefix", "", "The Key Prefix for SSM Parameter store")

}
