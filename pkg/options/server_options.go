package options

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"github.com/practice/multi_resource/pkg/server"
	"github.com/spf13/pflag"
	"path/filepath"
)

type ServerOptions struct {
	Address    int
	CertDir    string
	ConfigPath string
	HealthPort int
	CtlPort    int
	DebugMode  bool
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}

func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.Address, "address", 8888,
		"Address of http server to bind on. Default to 8888.")
	fs.IntVar(&o.HealthPort, "HealthPort", 29999,
		"Address of health http server to bind on. Default to 29999.")
	fs.IntVar(&o.CtlPort, "CtlPort", 8888,
		"Port used for ctl configuration server.")

	fs.StringVar(&o.ConfigPath, "config", "resources/config_home_test.yaml",
		"kubeconfig path for k8s cluster")
	fs.StringVar(&o.CertDir, "cert-dir", "",
		"The directory of cert to use, use tls.crt and tls.key as certificates. Default to disable https.")

	fs.BoolVar(&o.DebugMode, "debug-mode", false,
		"Debug Mode")

}

func (o *ServerOptions) Complete() error {
	return nil
}

func (o *ServerOptions) Validate() []error {
	var errs []error

	return errs
}

func (o *ServerOptions) NewServer() (*server.Server, error) {
	var tlsConfig *tls.Config = nil
	if o.CertDir != "" {
		cert, err := tls.LoadX509KeyPair(filepath.Join(o.CertDir, "tls.crt"), filepath.Join(o.CertDir, "tls.key"))
		if err != nil {
			return nil, errors.WithMessage(err, "unable to load tls certificate")
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return server.NewServer(o.Address, tlsConfig), nil
}
