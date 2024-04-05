package main

import (
	purevpnwg "github.com/Rikpat/purevpnwg/cmd"
	"github.com/Rikpat/purevpnwg/pkg/util"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
)

func main() {
	ctx := kong.Parse(&purevpnwg.CLI, kong.Configuration(kongyaml.Loader, "/etc/purevpnwg/config.yml", "~/.purevpnwg.yml", util.CONFIG_FILE))
	err := ctx.Run(&purevpnwg.Context{Config: purevpnwg.CLI.Config})
	ctx.FatalIfErrorf(err)
}
