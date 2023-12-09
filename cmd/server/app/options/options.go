package options

import (
	"flag"
	"github.com/practice/multi_resource/pkg/options"

	"github.com/pkg/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
)

type Options struct {
	Server *options.ServerOptions
	MySQL  *options.MySQLOptions
	Logs   *logs.Options
}

func NewOptions() *Options {
	return &Options{
		Server: options.NewServerOptions(),
		MySQL:  options.NewMySQLOptions(),
		Logs:   logs.NewOptions(),
	}
}

func (o *Options) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	fss.FlagSet("generic").AddGoFlagSet(flag.CommandLine)

	logs.AddGoFlags(flag.CommandLine)

	o.Server.AddFlags(fss.FlagSet("server"))
	o.MySQL.AddFlags(fss.FlagSet("mysql"))
	return fss
}

func (o *Options) Complete() error {

	if err := o.Server.Complete(); err != nil {
		return err
	}
	if err := o.MySQL.Complete(); err != nil {
		return err
	}

	return nil
}

func (o *Options) Validate() error {
	var errs []error

	errs = append(errs, o.Server.Validate()...)
	errs = append(errs, o.MySQL.Validate()...)

	if len(errs) == 0 {
		return nil
	}

	wrapped := errors.New("options validate error")
	for _, err := range errs {
		wrapped = errors.WithMessage(wrapped, err.Error())
	}
	return wrapped
}
