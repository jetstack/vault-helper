package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/coreos/go-systemd/daemon"
	"github.com/jetstack-experimental/vault-helper/pkg/dev_server"
)

// initCmd represents the init command
var devServerCmd = &cobra.Command{
	Use:   "dev-server [cluster ID]",
	Short: "Run a vault server in development mode with kubernetes PKI created.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		if len(args) < 1 {
			log.Fatalf("no cluster ID was given")
		}

		wait, err := cmd.PersistentFlags().GetBool(dev_server.FlagWaitSignal)
		if err != nil {
			log.Fatalf("error finding wait value: %v", err)
		}

		port, err := cmd.PersistentFlags().GetInt(dev_server.FlagPortNumber)
		if err != nil {
			log.Fatalf("error finding port value: %v", err)
		}
		if port > 65536 {
			log.Fatalf("invalid port %d > 65536", port)
		}
		if port < 1 {
			log.Fatalf("invalid port %d < 1", port)
		}
		if port < 1 {
			logrus.Fatalf("invalid port %d < 1", port)
		}

		v := dev_server.New(log)
		v.Vault.SetPort(port)

		if err := v.Run(cmd, args); err != nil {
			log.Fatal(err)
		}

		for n, t := range v.Kubernetes.InitTokens() {
			log.Infof(n + "-init_token := " + t)
		}

		daemon.SdNotify(false, "READY=1")

		if wait {
			waitSignal(v)
		}
	},
}

func init() {
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenEtcd, "", "Set init-token-etcd   (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenWorker, "", "Set init-token-worker (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenMaster, "", "Set init-token-master (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenAll, "", "Set init-token-all    (Default to new token)")

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
