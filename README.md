solbuild
--------

[![Report](https://goreportcard.com/badge/github.com/solus-project/solbuild)](https://goreportcard.com/report/github.com/solus-project/solbuild) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Solus package builder. This is still a **Work In Progress**.

solbuild is a [Solus project](https://solus-project.com/).

![logo](https://build.solus-project.com/logo.png)

**evobuild (legacy) workflow**:

 - Mount overlayfs with profile directory as lower node
 - Copy aux files into root
 - Bring up dbus/eopkg/etc
 - Upgrade & install `system.devel`
 - Bind mount system wide cache directory
 - Chroot, run build as root
 - If successful, copy newly built files out of chroot into current directory

**Issues with the `evobuild` approach**:

The build process is delicate and very much a simple script. We allow the chroot'd
process to download and install both build dependencies and the sources themselves.
We continue to allow networking within this chroot, which in turn allows any package
to download it's own additional tarballs without warning (`LibreOffice`, anyone?)

On top of this, all builds are performed as root. This allows the process within
the chroot to break the chroot itself quite badly. While it's not the worst problem
in the world (this is why we `chroot` after all) it's not **good**. Instead, we'll
operate with "normal" permissions, and allow `ypkg` to utilise `fakeroot` to complete
it's tasks.

**Note**:

For legacy `pspec.xml` format builds, not all of these steps are possible, however
this format will no longer be supported within Solus with the advent of the `sol`
package manager. For now we'll disable some steps when building old style packages.

**solbuild proposed workflow**:

 - Enter new namespace (`unshare`) - preserve networking here
 - Mount overlayfs from *configuration-based* profile
 - Bring up services
 - Add any required repositories
 - Upgrade & install `system.devel`
 - Copy aux files
 - Fetch sources & cache in system, bind mount **individual sources**
 - Request installation of build dependencies
 - Now `unshare` networking
 - Begin build in chroot namespace as unprivileged user
 - If successful, copy newly built files out of chroot into current directory

**TODO**:

 - [x] Port `update` and `chroot` to manager interface
 - [x] Restore `eopkg build` support for legacy format.
 - [x] Restore `ccache` bind mount support
 - [ ] Add `tmpfs` support for builds
 - [ ] Add custom repo support
 - [ ] Add new `networking` key to `ypkg` files to disable network isolation
 - [ ] Add profile concept *based on backing images*
 - [ ] Add config file support
 - [ ] Add documentation
 - [ ] Seal the deal, v1

License
-------

Copyright © 2016 Solus Project

`solbuild` is available under the terms of the Apache-2.0 license
