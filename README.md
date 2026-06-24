# Inventory SDK

This repo contains the SDK for Plakar Control-Plane Inventories.
Inventories are plugins that list resources (computes, buckets,
databases, etc...) in a cloud provider, or similar.

This repository contains both the plugin and the consumer part.  The
plugin (in `sdk/server/`) part is what needs to be used by inventory
plugins, while the consumer (`sdk/client`) is to receive the
information provided from an inventory plugin.

To ease the development, use `cmd/run` to test your inventory:

	$ make
	$ ./run -o param1=foo -o param2=bar ./path/to/inventory-executable

this should print a series of JSON-objects, inventory entries, that
your inventory is providing.


## Example Inventory

A complete example that can used for scaffolding purposes can be found
in the [example][./example] directory.

To try it out:

	$ go build -v ./example/cmd/example-inventory
	$ make
	$ ./run ./example-inventory

