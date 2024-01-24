package options

import (
	"crypto/tls"
	"fmt"
	"github.com/myoperator/multiclusteroperator/pkg/server"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
)

type ServerOptions struct {
	Port       int
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
	fs.IntVar(&o.Port, "address", 8888,
		"Address of http server to bind on. Default to 8888.")
	fs.IntVar(&o.HealthPort, "HealthPort", 29999,
		"Address of health http server to bind on. Default to 29999.")
	fs.IntVar(&o.CtlPort, "ctl-port", 8888,
		"Port used for ctl configuration server.")

	fs.StringVar(&o.ConfigPath, "config", "resources/config_home_test.yaml",
		"kubeconfig path for k8s cluster")
	fs.StringVar(&o.CertDir, "cert-dir", "",
		"The directory of cert to use, use tls.crt and tls.key as certificates. Default to disable https.")

	fs.BoolVar(&o.DebugMode, "debug-mode", false, "Debug Mode")

}

// Complete TODO: 实现赋值逻辑
func (o *ServerOptions) Complete() error {
	return nil
}

// Validate 验证逻辑
func (o *ServerOptions) Validate() []error {
	var errs []error

	// 验证 Address 字段是否在有效范围内
	if o.Port <= 0 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("Invalid server address. Must be within the range of 1-65535"))
	}

	// 验证 HealthPort 字段是否在有效范围内
	if o.HealthPort <= 0 || o.HealthPort > 65535 {
		errs = append(errs, fmt.Errorf("Invalid health port. Must be within the range of 1-65535"))
	}

	// 验证 CtlPort 字段是否在有效范围内
	if o.CtlPort <= 0 || o.CtlPort > 65535 {
		errs = append(errs, fmt.Errorf("Invalid ctl port. Must be within the range of 1-65535"))
	}

	// 验证 ConfigPath 字段是否为空
	if o.ConfigPath == "" {
		errs = append(errs, fmt.Errorf("Config path is required"))
	}

	// 验证 CertDir 字段是否符合要求
	if o.CertDir != "" {
		// 在这里可以添加更多的证书目录验证逻辑
		// 例如，验证证书文件是否存在、证书文件格式是否正确等
		if _, err := os.Stat(o.CertDir); os.IsNotExist(err) {
			errs = append(errs, errors.New("Cert directory does not exist"))
		}
	}

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
	return server.NewServer(o.Port, tlsConfig), nil
}
