# kcli
Kcli is a kafka read only command line browser.

## Install

Binaries are provided [here](https://github.com/cswank/kcli/releases/tag/1.1.0) (windows
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

<img src="./docs/help.png" width="396"/>

### Jumping
You can use the jump command (C-j) to set the current offset of a partition
or topic.  Jumping on a partition is simple: the number you enter becomes
the current offset.  Jumping on a topic is a bit different.  The number you
enter sets the current offset of each partition relative to either the 1st
offset or last offset.  If the number you enter (N) is positive then the current
offset becomes first offset + N.  If the number you enter is negative
then the current offset becomes last offset + N.

### Searching
You can search for a string on either a partition or topic.  When you search
on a partition then the current offset is set to the first message that
contains the search string.  When you search on a topic then only the topics
that contain a match are printed to the screen and their current offset is
set to the first message that contains that match.

If you have partitions that have large amounts of data then it can take a
long time to search through all the partitions.  It is sometimes useful
to use the partition jump functionality described above to speed up your
search if you have an idea where the message might be.  If you know the message
you are searching for is fairly recent then you can use a negative jump to set
each offset close to then last offset.  The search will then start from those
offsets.

### Screen Colors

If you don't like the defaul colors you can set KCLI_COLOR[1,2,3] to one of:

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

See it in action at [asciinema](https://asciinema.org/a/7wobtflusydvswbsn1d320o1g)

[![asciicast](https://asciinema.org/a/7wobtflusydvswbsn1d320o1g.png)](https://asciinema.org/a/7wobtflusydvswbsn1d320o1g)

NOTE: If you are connecting to a local kafka that is running in a docker container
using wurstmeister/kafka you may have the env KAFKA_ADVERTISED_HOST_NAME set to
a name that is used by other containers that need to connect to kafka.  This will
cause kcli to not be able to read from kafka.  A hacky fix is to edit your /etc/hosts
file and add another name to the 127.0.0.1 network interface.  For example, if

    KAFKA_ADVERTISED_HOST_NAME=kafka

Then the 127.0.0.1 line /etc/hosts should look like:

    127.0.0.1       localhost kafka
