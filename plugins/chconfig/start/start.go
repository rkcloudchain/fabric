package start

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var logger = flogging.MustGetLogger("chconfig.start")

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chconfig REST server",
	Long:  "Start the chconfig REST server",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var loggingLevel string
		if viper.GetString("logging_level") != "" {
			loggingLevel = viper.GetString("logging_level")
		} else {
			loggingLevel = viper.GetString("logging.level")
		}
		flogging.InitFromSpec(loggingLevel)
		return common.InitCrypto(viper.Get("msp.path").(string), viper.Get("msp.id").(string), "bccsp")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer(fmt.Sprintf("%s:%d", viper.Get("listen.address"), viper.Get("listen.port")))
	},
}

// Cmd returns the cobra command for start
func Cmd() *cobra.Command {
	addFlags(startCmd)
	return startCmd
}

func addFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()

	flags.String("hostname", "0.0.0.0", "The hostname or IP on which the REST server will listen")
	flags.Int("port", 8059, "The port on which the REST server will listen")
	flags.String("mspdir", "msp", "The MSP path")
	flags.String("mspid", "SampleOrg", "The MSP id")
	flags.StringP("orderer", "o", "", "Ordering service endpoint")
	flags.Bool("tls", false, "Use TLS when communicating with the orderer endpoint")
	flags.String("cafile", "", "Path to file containing PEM-encoded trusted certificate(s) for the ordering endpoint")
	flags.Duration("connTimeout", 3*time.Second, "Timeout for client to connect")
	flags.Bool("clientauth", false, "Use mutual TLS when communicating with the orderer endpoint")
	flags.String("keyfile", "", "Path to file containing PEM-encoded private key to use for mutual TLS communication with the orderer endpoint")
	flags.String("certfile", "", "Path to file containing PEM-encoded X509 public key to use for mutual TLS communication with the orderer endpoint")

	bindFlags(flags)
}

func bindFlags(flags *pflag.FlagSet) {
	bindFlag("listen.address", "hostname", flags)
	bindFlag("listen.port", "port", flags)
	bindFlag("msp.path", "mspdir", flags)
	bindFlag("msp.id", "mspid", flags)
	bindFlag("orderer.address", "orderer", flags)
	bindFlag("orderer.tls.enabled", "tls", flags)
	bindFlag("orderer.tls.rootcert.file", "cafile", flags)
	bindFlag("orderer.client.connTimeout", "connTimeout", flags)
	bindFlag("orderer.tls.clientAuthRequired", "clientauth", flags)
	bindFlag("orderer.tls.clientKey.file", "keyfile", flags)
	bindFlag("orderer.tls.clientCert.file", "certfile", flags)
}

func bindFlag(key, name string, flags *pflag.FlagSet) {
	flag := flags.Lookup(name)
	if flag == nil {
		panic(errors.Errorf("failed to lookup '%s'", name))
	}
	viper.BindPFlag(key, flag)
}

func startServer(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	logger.Infof("Serving HTTP requests on %s", listener.Addr())
	err = http.Serve(listener, router())
	if err != nil {
		logger.Errorf("Error starting server:[%s]\n", err)
		return err
	}

	return nil
}

func router() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/chconfig/update-from-configs", computeUpdateFromConfigs).Methods(http.MethodPost)
	r.HandleFunc("/chconfig/sign-config-tx", signConfigTx).Methods(http.MethodPost)
	r.HandleFunc("/chconfig/update", update).Methods(http.MethodPost)

	return r
}
