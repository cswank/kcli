# kcli
Kcli is a kafka read only command line browser.

## Install

Binaries are provided [here](https://github.com/cswank/kcli/releases/tag/1.0.0) (windows
is not tested).  If you have go installed you can do:

    $ go get -u github.com/cswank/kcli

## Usage

    $ kcli --help
    usage: kcli [<flags>]

    Flags:
          --help       Show context-sensitive help (also try --help-long and --help-man).
      -a, --addresses=localhost:9092 ...
                       comma seperated list of kafka addresses
      -l, --logs=LOGS  for debugging, set the log output to a file

Once you start kcli type 'h' to see the help menu:

<img src="./docs/help.png" width="620"/>

If you don't like the colors you can set KCLI_COLOR[1,2,3] to one of:

* black
* red
* green
* yellow
* blue
* magenta
* cyan
* white

For example:

    $ KCLI_COLOR1=red KCLI_COLOR2=yellow kcli

See it in action at [asciinema](https://asciinema.org/a/110096)

[![asciicast](https://asciinema.org/a/110096.png)]()


