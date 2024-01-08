package main

import (
	"log"

	"github.com/travisjeffery/mocker/pkg/mocker"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var c = mocker.Config{}

func init() {
	kingpin.Version("1.1.0")

	kingpin.Arg("source-file", "Source file containing interfaces to generate mocks from.").StringVar(&c.Src)
	kingpin.Arg("source-interfaces", "List of interface names to mock. Comma delimited.").StringsVar(&c.Itf)
	kingpin.Flag("destination", "File to write generated mocks in. Default is stdout.").Short('d').StringVar(&c.Dst)
	kingpin.Flag("package", "Name of the mock's package. Inferred by default.").Short('p').StringVar(&c.Pkg)
	kingpin.Flag("prefix", "Prefix to put in front of the generated interface mock names.").Short('P').Default("Mock").StringVar(&c.Pre)
	kingpin.Flag("suffix", "Suffix to put at the enf of the generated interface mock names.").Short('S').StringVar(&c.Suf)
	kingpin.Flag("import-path", "The full package import path for the generated code. The purpose of this flag is to prevent import cycles in the generated code by trying to include its own package.").Short('s').StringVar(&c.Slf)

	// to maintain backwards compatibility
	kingpin.Flag("dst", "").Hidden().StringVar(&c.Dst)
	kingpin.Flag("pkg", "").Hidden().StringVar(&c.Pkg)
	kingpin.Flag("selfpkg", "").Hidden().StringVar(&c.Slf)
}

func main() {
	kingpin.Parse()

	if err := mocker.Run(c); err != nil {
		log.Fatalf("mocker: failed to mock: %v", err)
	}
}
