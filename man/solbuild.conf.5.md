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
configuration files. 

## EXAMPLE


## COPYRIGHT

 * Copyright © 2016 Ikey Doherty, License: CC-BY-SA-3.0


## SEE ALSO


`solbuild(1)`

https://github.com/toml-lang/toml

## NOTES

Creative Commons Attribution-ShareAlike 3.0 Unported

 * http://creativecommons.org/licenses/by-sa/3.0/
