solbuild(1) -- Solus package builder
=================================


## SYNOPSIS

`solbuild [subcommand] <flags>`


## DESCRIPTION

`solbuild(1)` is a `chroot(2)` based package build system, used to safely and
efficiently build Solus packages from source.

`solbuild(1)` makes use of `OverlayFS` to provide a simple caching system, whereby
a base image (provided by the Solus project) is used as the bottom-most, read-only
layer, and changes are made in temporary upper layers.

When building `package.yml` files (`ypkg`), the tool will also disable all
networking within the environment, apart from the loopback device. This is
intended to prevent uncontrolled build environments in which a package may
be fetching external, unverified sources, during the build.

This behaviour can be turned off on a package basis, by setting the `networking`
key to `true` within the YML file. This should only be used when it is completely
unavoidable, however, as the container mechanism is there for a reason. Trust.

With both build types, legacy and `ypkg`, the tool will enter an isolated namespace
using the `unshare(2)` system call. It intends to provide a highly controlled
build environment, and providing a robust container in which to build packages
intended for use in production.

## OPTIONS

These options apply to all subcommands within `solbuild(1)`.

 * `-h`, `--help`

   Help provides an explanation for any command or subcommand. Without any
   specified subcommands it will list the main subcommands for the application.

 * `-p`, `--profile`

   Set the build configuration profile to use with all operations.

 * `-d`, `--debug`

   Enable extra logging messages with debug level, useful to assist in further
   introspection of the environment setup and teardown..


## SUBCOMMANDS


`build [package.yml] | [pspec.xml]`

    Build the given package in a chroot environment, and upon success,
    store those packages in the current directory.

    If you do not pass a package file as an argument to `build`, it will look
    for the files in the current working directory. The priority is always given
    to `package.yml` files, falling back to `pspec.xml`, the legacy build format.

 * `-t`, `--tmpfs`:

        Instruct `solbuild(1)` to use a `tmpfs` mount as the bottom most point
        in the chroot layer system. This can drastically improve build times,
        as most of the changes are happening purely in memory. If running on
        a memory constrained device, please consider setting an appropriate
        upper constraint. See the next flag for more details.

 *  `-m`, `--memory`

        Set the contraint size for `tmpfs` mounts used by `solbuild(1)`. This is
        only useful in conjunction with the `-t` option.

`chroot [package.yml] | [pspec.xml]`

    Interactively chroot into the package's build environment, to enable
    further inspection when issues aren't immediately resolvable, i.e. pkg-config
    dependencies.

`init`

    Initialise a solbuild profile so that it can be used for subsequent
    builds. You must perform this step if you wish to do any kind of useful
    operations with `solbuild(1)`.

    The init command respects the global `--profile` option, however you
    may pass the name of the profile as an argument instead if you wish.

 *  `-u`, `--update`

        Passing the update flag will cause `solbuild(1)` to automatically update
        the base image, after it has successfully initialised it.

`update [profile]`

    Update the base image of the specified solbuild profile, helping to
    minimize the build times in future updates with this profile.

    The update command respects the global `--profile` option, however you
    may pass the name of the profile as an argument instead if you wish.

`version`

    Print the version and copyright notice of `solbuild(1)` and exit.


## EXIT STATUS

On success, 0 is returned. A non-zero return code signals a failure.


## COPYRIGHT

 * Copyright © 2016 Ikey Doherty, License: CC-BY-SA-3.0


## SEE ALSO


https://github.com/solus-project/solbuild

https://github.com/solus-project/ypkg


## NOTES

Creative Commons Attribution-ShareAlike 3.0 Unported

 * http://creativecommons.org/licenses/by-sa/3.0/
