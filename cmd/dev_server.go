package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/dev_server"
)

// initCmd represents the init command
var devServerCmd = &cobra.Command{
	Use:   "dev-server [cluster ID]",
	Short: "Run a vault server in development mode with kubernetes PKI created.",
	Run: func(cmd *cobra.Command, args []string) {

		logger := logrus.New()

		i, err := RootCmd.PersistentFlags().GetInt("log-level")
		if err != nil {
			logrus.Fatalf("failed to get log level of flag: %s", err)
		}
		if i < 0 || i > 2 {
			logrus.Fatalf("not a valid log level")
		}
		switch i {
		case 0:
			logger.Level = logrus.FatalLevel
		case 1:
			logger.Level = logrus.InfoLevel
		case 2:
			logger.Level = logrus.DebugLevel
		}

		log := logrus.NewEntry(logger)

		if len(args) < 1 {
			logrus.Fatalf("no cluster ID was given")
		}

		wait, err := cmd.PersistentFlags().GetBool(dev_server.FlagWaitSignal)
		if err != nil {
			logrus.Fatalf("error finding wait value: %v", err)
		}

		port, err := cmd.PersistentFlags().GetInt(dev_server.FlagPortNumber)
		if err != nil {
			logrus.Fatalf("error finding port value: %v", err)
		}

		if port > 65536 {
			logrus.Fatalf("invalid port %d > 65536", port)
		}
		if port < 1 {
			logrus.Fatalf("invalid port %d < 1", port)
		}

		v := dev_server.New(log)
		v.Vault.SetPort(port)
		if err := v.Vault.Start(); err != nil {
			logrus.Fatalf("unable to initialise dev vault: %s", err)
		}

		if err := v.Run(cmd, args); err != nil {
			logrus.Fatal(err)
		}

		for n, t := range v.Kubernetes.InitTokens() {
			logrus.Infof(n + "-init_token := " + t)
		}

		if wait {
			waitSignal(v)
		}
	},
}

func init() {
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	devServerCmd.Flag(dev_server.FlagMaxValidityCA).Shorthand = "c"

	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	devServerCmd.Flag(dev_server.FlagMaxValidityAdmin).Shorthand = "d"

	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")
	devServerCmd.Flag(dev_server.FlagMaxValidityComponents).Shorthand = "s"

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenEtcd, "", "Set init-token-etcd   (Default to new token)")
	devServerCmd.Flag(dev_server.FlagInitTokenEtcd).Shorthand = "e"

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenWorker, "", "Set init-token-worker (Default to new token)")
	devServerCmd.Flag(dev_server.FlagInitTokenWorker).Shorthand = "o"

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenMaster, "", "Set init-token-master (Default to new token)")
	devServerCmd.Flag(dev_server.FlagInitTokenMaster).Shorthand = "m"

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenAll, "", "Set init-token-all    (Default to new token)")
	devServerCmd.Flag(dev_server.FlagInitTokenAll).Shorthand = "a"

	devServerCmd.PersistentFlags().Bool(dev_server.FlagWaitSignal, true, "Wait for TERM + QUIT signal has been given before termination")
	devServerCmd.Flag(dev_server.FlagWaitSignal).Shorthand = "w"

	devServerCmd.PersistentFlags().Int(dev_server.FlagPortNumber, 8200, "Set the port number to connect to vault")
	devServerCmd.Flag(dev_server.FlagPortNumber).Shorthand = "t"

	RootCmd.AddCommand(devServerCmd)
}

func waitSignal(v *dev_server.DevVault) {
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exit_chan := make(chan int)

	go func() {
		for {
			s := <-signal_chan
			switch s {
			case syscall.SIGTERM:
				exit_chan <- 0

			case syscall.SIGQUIT:
				exit_chan <- 0
			}
		}
	}()

	<-exit_chan
	v.Vault.Stop()
}
