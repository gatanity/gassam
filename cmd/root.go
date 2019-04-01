package cmd

import (
	"context"
	"fmt"
	"github.com/cybozu/arws/aws"
	"github.com/cybozu/arws/config"
	"github.com/cybozu/arws/defaults"
	"github.com/cybozu/arws/idp"
	"github.com/cybozu/arws/prompt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// Execute runs root command
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var configure bool

	cmd := &cobra.Command{
		Use:   "arws",
		Short: "arws simplifies AssumeRoleWithSAML with CLI",
		Long:  `It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.`,
		Run: func(_ *cobra.Command, args []string) {
			if configure {
				err := configureSettings()
				if err != nil {
					log.Panic(err)
				}
				return
			}

			cfg, err := config.NewConfig()
			if err != nil {
				log.Panic(err)
			}

			request, err := aws.CreateSAMLRequest(cfg.AppIDURI)
			if err != nil {
				log.Panic(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			azure := idp.NewAzure(request, cfg.AzureTenantID)
			base64Response, err := azure.Authenticate(ctx)
			if err != nil {
				log.Panic(err)
			}

			response, err := aws.ParseSAMLResponse(base64Response)
			if err != nil {
				log.Panic(err)
			}

			roleArn, principalArn, err := aws.ExtractRoleArnAndPrincipalArn(*response)
			if err != nil {
				log.Panic(err)
			}

			credentials, err := aws.AssumeRoleWithSAML(ctx, roleArn, principalArn, base64Response)
			if err != nil {
				log.Panic(err)
			}

			err = aws.SaveCredentials("default", *credentials)
			if err != nil {
				log.Panic(err)
			}
		},
	}
	cmd.PersistentFlags().BoolVarP(&configure, "configure", "c", false, "configure initial settings")

	return cmd
}

func configureSettings() error {
	p := prompt.NewPrompt()
	cfg := config.Config{}

	var err error
	cfg.AzureTenantID, err = p.AskString("Azure Tenant ID", nil)
	if err != nil {
		return err
	}

	cfg.AppIDURI, err = p.AskString("App ID URI", nil)
	if err != nil {
		return err
	}

	cfg.DefaultSessionDurationHours, err = p.AskInt("Default Session Duration Hours (1-12)", &prompt.Options{
		ValidateFunc: func(val string) error {
			duration, err := strconv.Atoi(val)
			if err != nil || duration < 1 || 12 < duration {
				return fmt.Errorf("default session duration hours must be between 1 and 12: %s", val)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	cfg.ChromeUserDataDir, err = p.AskString("Chrome User Data Directory", &prompt.Options{
		Default: filepath.Join(defaults.UserHomeDir(), ".config", "arws", "chrome-user-data"),
	})
	if err != nil {
		return err
	}

	return config.Save(cfg)
}