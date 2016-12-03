solbuild.conf(5) -- solbuild configuration
==========================================

## NAME

    solbuild.conf - configuration for solbuild
    
## SYNOPSIS

    /usr/share/solbuild/*.conf
    
    /etc/solbuild/*.conf


## DESCRIPTION

`solbuild(1)` uses configuration files from the above mentioned directories to
configure various aspects of the `solbuild` defaults.

All configuration files must be valid prior to `solbuild(1)` launching, as it
will load and validate them all into a merged configuration. Using a layered
approach, `solbuild` will first read from the global vendor directory,
`/usr/share/solbuild`, before finally loading from the system directory,
`/etc/solbuild`.

`solbuild(1)` is capable of running without configuration, and this method
permits a stateless implementation whereby vendor & system administrator
configurations are respected in the correct order.

## CONFIGURATION FORMAT

`solbuild` uses the `TOML` configuration format for all of it's own
configuration files. This is a strongly typed configuration format, whereby
strict validation occurs against expected key types.

* `default_profile`

    Set the default profile used by `solbuild(1)`. This must have a string value,
    and will be used by `solbuild(1)` in the absence of the `-p`,`--profile`
    flag.

* `enable_tmpfs`

    Instruct `solbuild(1)` to use tmpfs mounts by default for all builds. Note
    that even if this is disabled, as it is by default, you may still override
    this at runtime with the `-t`,`--tmpfs` flag.

* `tmpfs_size`

    Set the default tmpfs size used by `solbuild(1)` when tmpfs builds are
    enabled. An empty value, the default, will mean an unbounded size to
    the tmpfs. This value should be a string value, with the same syntax
    that one would pass to `mount(8)`.

    See `solbuild(1)` for more details on the `-t`,`--tmpfs` option behaviour.


## EXAMPLE

    # Set the default profile, a string value assignment
    default_profile = "main-x86_64"

    # Set tmpfs enabled by default, a boolean value assignment
    enable_tmpfs = true


## COPYRIGHT

 * Copyright © 2016 Ikey Doherty, License: CC-BY-SA-3.0


## SEE ALSO


`solbuild(1)`

https://github.com/toml-lang/toml

## NOTES

Creative Commons Attribution-ShareAlike 3.0 Unported

 * http://creativecommons.org/licenses/by-sa/3.0/
