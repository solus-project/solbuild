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

solbuild is a [Solus project](https://solus-project.com/).

![logo](https://build.solus-project.com/logo.png)

**TODO**:

 - [x] Port `update` and `chroot` to manager interface
 - [x] Restore `eopkg build` support for legacy format.
 - [x] Restore `ccache` bind mount support
 - [x] Add `tmpfs` support for builds
 - [x] Restore `.solus/packager` support
 - [x] Add an `--update,-u` flag to `init` to automatically update it
 - [x] Add profile concept *based on backing images*
 - [x] Add config file support
 - [x] Add custom repo support
 - [x] Add new `networking` key to `ypkg` files to disable network isolation
 - [x] Add locking of `Overlay` and `BackingImage` storage
 - [x] Restore `history.xml` generation for `ypkg` builds in git
 - [x] Restore support for `ypkg` git sources
 - [x] Restore `git submodule` support
 - [x] Add documentation
 - [x] Add `index` command
 - [ ] Seal the deal, v1

**Future considerations**:

 - Abstract the package manager stuff, port it back into [libosdev](https://github.com/solus-project/libosdev)
 - Add `sol` support when the time is right..
 - When `sol` support has landed, always mount `overlayfs` with `nosuid`
 - Generate the builder base images using [USpin](https://github.com/solus-project/uspin)


License
-------

Copyright © 2016 Solus Project

`solbuild` is available under the terms of the Apache-2.0 license
