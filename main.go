package main

import (
	"flag"

	"github.com/ohoonice/protoc-gen-sql/internal"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	omitempty   = flag.Bool("omitempty", true, "omit if google.api is empty")
	outdir      *string
	database    *string
)

func main() {
	flag.Parse()
	if *showVersion {
		//fmt.Printf("protoc-gen-sql %v\n", release)
		return
	}

	var flags flag.FlagSet
	outdir = flags.String("outdir", "", "path to sql dir")
	database = flags.String("database", "", "database")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			internal.GenerateFile(gen, f, *outdir, *database)
		}
		return nil
	})
}
