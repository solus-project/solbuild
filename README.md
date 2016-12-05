solbuild
--------

[![Report](https://goreportcard.com/badge/github.com/solus-project/solbuild)](https://goreportcard.com/report/github.com/solus-project/solbuild) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`solbuild` is a `chroot` based package build system, used to safely and efficiently build Solus packages from source, in a highly controlled and isolated environment. This tool succeeds the `evobuild` tool, originally in Evolve OS, which is now known as Solus. The very core concept of the layered builder has always remained the same, this is the next .. evolution.. of the tool.

`solbuild` makes use of `OverlayFS` to provide a simple caching system, whereby a base image (provided by the Solus project) is used as the bottom-most, read-only layer, and changes are made in temporary upper layers. Currently the project provides two base images for the default profiles shipped with `solbuild`:

 - **main-x86_64**: Built using the stable Solus repositories, suitable for production deployments for `shannon` users.
 - **unstable-x86_64**: Built using the unstable Solus repositories, ideal for developers, and the Solus build machinery for the repo waterfall prior to `shannon` inclusion.

When building `package.yml` files ([ypkg](https://github.com/solus-project/ypkg)), the tool will also disable all networking within the environment, apart from the loopback device. This is intended to prevent uncontrolled build environments in which a package may be fetching external, unverified sources, during the build.

`solbuild` also allows developers to control the repositories used by configuring the profiles:

 - Remove any base image repo
 - Add any given repo, remote or local. Local repos are bind mounted and can be automatically indexed by `solbuild`.

`solbuild` performs heavy caching throughout, with source archives being stored in unique hash based directories globally, and the `ccache` being retained after each build through bind mounts. A single package cache is retained to speed up subsequent builds, and users may speed this up further by updating their base images.

As a last speed booster, `solbuild` allows you to perform builds in memory via the `--tmpfs` option.

**Note**: `solbuild` is designed in such a way that you *do not need to be running Solus*. You can build packages for Solus from any compatible host.

solbuild is a [Solus project](https://solus-project.com/).

![logo](https://build.solus-project.com/logo.png)

Pending items for Solus switch
------------------------------

There are a couple of items left to address before we switch the Solus build infrastructure over to `solbuild`:

- [ ] Add CLI flag to disable colour output
- [ ] Add rewriting of `eopkg.conf` to conditionally disable `dbginfo` packages, i.e. for the kernel package.
- [ ] Add `tmpfs` option to `ypkg` to disable a package from building in tmpfs when the default config specifies to use tmpfs. i.e. for LibreOffice.

Getting started
----------------

**Solus Users**

    sudo eopkg up
    sudo eopkg it solbuild

    # If you only ever want to use the unstable repo by default
    sudo eopkg it solbuild-config-unstable

**Everyone else**

    git clone https://github.com/solus-project/solbuild.git
    cd solbuild
    make ensure_modules
    make
    sudo make install

You may wish to use the provided tarballs, which include vendored dependencies.
Distributions are free to nuke the src/vendor directory from the distributed
tarball and use their own golang dependencies if appropriate.

**Initialising the root**

Run the following command to fetch and install the base image. If you wish
to change the profile, use the `-p` flag (`unstable-x86_64` or `main-x86_64`)
The `-u` flag will automatically update the image.

    sudo solbuild init -u

**Updating the image**

    # Update the default profile
    sudo solbuild update

    # Update a specific profile
    sudo solbuild update unstable-x86_64

**Building packages**

    # Build the first package found in the current directory
    sudo solbuild build

    # Build a specific path
    sudo solbuild build ../mypackages/package.yml

    # Build for unstable profile
    sudo solbuild -p unstable-x86_64 build

See the `solbuild help` command for more details, or `solbuild(1)` manpage.

Requirements
------------

 - golang (tested with 1.7.4)
 - `libgit2` (Also require `git` at runtime for submodules)
 - `curl` command

Your kernel must support the `overlayfs` filesystem.
Git is required as `solbuild` supports the `git|` source type of ypkg files. Additionally, `solbuild` will try to generate a package changelog from the git history where the YPKG file is found. This is used within Solus to create a changelog dynamically from the git tags, and automatically marking security updates, etc.

License
-------

Copyright © 2016 Solus Project

`solbuild` is available under the terms of the Apache-2.0 license
