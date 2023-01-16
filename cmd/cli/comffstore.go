package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/comfforts/comff-stores/internal/agent"
	"github.com/comfforts/comff-stores/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

type cli struct {
	cfg cfg
}

func main() {
	cli := cli{}

	cmd := &cobra.Command{
		Use:     "comffstores",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func setupFlags(cmd *cobra.Command) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	cmd.Flags().String("config-file", "", "path to config file.")

	cmd.Flags().String("run-dir", "", "agent run directory")
	cmd.Flags().String("data-dir", "data", "agent run directory")
	cmd.Flags().String("node-name", hostname, "Unique server ID.")
	cmd.Flags().String("bind-addr", "127.0.0.1:50050", "Address to bind serf on.")
	cmd.Flags().Int("rpc-port", 50051, "Port for RPC clients (and raft) connections.")
	cmd.Flags().StringSlice("peer-join-addrs", nil, "Serf addresses to join.")
	cmd.Flags().Bool("bootstrap", false, "Bootstrap the cluster.")

	cmd.Flags().String("policies-path", "", "Path to ACL model and policy file directory.")
	cmd.Flags().String("certs-path", "", "Path to TLS certs.")

	return viper.BindPFlags(cmd.Flags())
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	cfgFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	viper.SetConfigFile(cfgFile)
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	runDir := viper.GetString("run-dir")
	c.cfg.Config.RunDir = runDir
	c.cfg.Config.DataDir = path.Join(runDir, viper.GetString("data-dir"))
	c.cfg.Config.NodeName = viper.GetString("node-name")
	c.cfg.Config.BindAddr = viper.GetString("bind-addr")
	c.cfg.Config.RPCPort = viper.GetInt("rpc-port")
	c.cfg.Config.PeerNodeAddrs = viper.GetStringSlice("peer-join-addrs")
	c.cfg.Config.Bootstrap = viper.GetBool("bootstrap")

	policiesPath := viper.GetString("policies-path")
	c.cfg.Config.ACLModelFile = path.Join(runDir, policiesPath, config.PolicyFile(config.ACLModelFile))
	c.cfg.Config.ACLPolicyFile = path.Join(runDir, policiesPath, config.PolicyFile(config.ACLPolicyFile))

	certsPath := viper.GetString("certs-path")
	c.cfg.ServerTLSConfig.CertFile = filepath.Join(runDir, certsPath, config.CertFile(config.ServerCertFile))
	c.cfg.ServerTLSConfig.KeyFile = filepath.Join(runDir, certsPath, config.CertFile(config.ServerKeyFile))
	c.cfg.ServerTLSConfig.CAFile = filepath.Join(runDir, certsPath, config.CertFile(config.CAFile))

	c.cfg.PeerTLSConfig.CertFile = filepath.Join(runDir, certsPath, config.CertFile(config.ClientCertFile))
	c.cfg.PeerTLSConfig.KeyFile = filepath.Join(runDir, certsPath, config.CertFile(config.ClientKeyFile))
	c.cfg.PeerTLSConfig.CAFile = filepath.Join(runDir, certsPath, config.CertFile(config.CAFile))

	if c.cfg.ServerTLSConfig.CertFile != "" &&
		c.cfg.ServerTLSConfig.KeyFile != "" {
		c.cfg.ServerTLSConfig.Server = true
		c.cfg.Config.ServerTLSConfig, err = config.SetupTLSConfig(
			c.cfg.ServerTLSConfig,
		)
		if err != nil {
			return err
		}
	}

	if c.cfg.PeerTLSConfig.CertFile != "" &&
		c.cfg.PeerTLSConfig.KeyFile != "" {
		c.cfg.Config.PeerTLSConfig, err = config.SetupTLSConfig(
			c.cfg.PeerTLSConfig,
		)
		if err != nil {
			return err
		}

	}
	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	fmt.Printf("Agent config: %v\n", c.cfg)
	fmt.Printf("Agent is bootstrap: %v\n", c.cfg.Bootstrap)
	fmt.Printf("Agent run directory: %v\n", c.cfg.RunDir)
	fmt.Printf("Agent data directory: %v\n", c.cfg.DataDir)
	// return nil
	agent, err := agent.NewAgent(c.cfg.Config)
	if err != nil {
		return err
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	return agent.Shutdown()
}
