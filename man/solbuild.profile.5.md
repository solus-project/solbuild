solbuild.profile(5) -- solbuild profiles
==========================================

## NAME

    solbuild.profile - Profile definition for solbuild
    
## SYNOPSIS

    /usr/share/solbuild/*.profile
    
    /etc/solbuild/*.profile


## DESCRIPTION

`solbuild(1)` uses configuration files from the above mentioned directories to
define profiles used for builds. A `solbuild` profile is automatically named
to the basename of the file, without the `.profile` suffix.

As an example, if we have the file `/etc/solbuild/test.profile`, the name of
the profile in `solbuild(1)` would be **test**. With the layered stateless
approach in solbuild, any named profile in the system config directory `/etc/`
will take priority over the named profiles in the vendor directory. These
profiles are not merged, the one in `/etc/` will "replace" the one in the
vendor directory, `/usr/share/solbuild`.


## CONFIGURATION FORMAT

`solbuild` uses the `TOML` configuration format for all of it's own
configuration files. This is a strongly typed configuration format, whereby
strict validation occurs against expected key types.

* `image`

    Set the backing image to one of the (currently Solus) provided backing
    images. Valid values include:

        * `main-x86_64`
        * `unstable-x86_64`

    A string value is expected for this key.

* `remove_repos`

    This key expects an array of strings for the repo names to remove from the
    existing base image during builds. Currently the Solus provided images all
    use the name **Solus** for the repo they ship with.

    Setting this to a value of `['*']` will indicate removal of all repos.

* `add_repos`

    This key expects an array of strings for the repo names defined in this
    profile to add to the image. The default unset value, i.e. an absence
    of this key, or the special value `['*']` will enable all of the repos
    in the profile.

    This option may be useful for testing repos and conditionally disabling
    them for testing, without having to remove them from the file.

* `[repo.$Name]`

    A repository is defined with this key, where `$Name` is replaced with the
    name you intend to assign to the repository. By default, a repo definition
    is configured for a remote repository.

    * `[repo.$Name]` `uri`

        Set this to the remote repository URL, including the `eopkg-index.xml.xz`
        If the repository is a **local** one, you must include the path to the
        directory, with no suffix.

    * `[repo.$Name]` `local`

        Set this to true to configure `solbuild(1)` to add a local repository
        to the build. The build process will bind-mount the `uri` configured
        directory into the build and make it available.

    * `[repo.$Name]` `autoindex`

        Set this to true to instruct `solbuild(1)` to automatically reindex this
        local repository while in the container. This may be useful if you do
        not have the appropriate host side tools.

        `solbuild(1)` will only index the files once, at startup, before it has
        performed the upgrade and component validation. Once your build has
        completed, and your `*.eopkg` files are deposited in your current directory,
        you can simply copy them to your local repository directory, and then
        `solbuild` will be able to use them immediately in your next build.


## EXAMPLE

    # Use the unstable backing image for this profile
    image = "unstable-x86_64"

    # Restrict adding the repos to the Solus repo only
    add_repos = ['Solus']

    # Example of adding a remote repo
    [repo.Solus]
    uri = "https://packages.solus-project.com/unstable/eopkg-index.xml.xz"

    # Add a local repository by bind mounting it into chroot on each build
    [repo.Local]
    uri = "/var/lib/myrepo"
    local = true

    # If you have a local repo providing packages that exist in the main
    # repository already, you should remove the repo, and re-add it *after*
    # your local repository:
    remove_repos = ['Solus']
    add_repos = ['Local','Solus']



## COPYRIGHT

 * Copyright © 2016 Ikey Doherty, License: CC-BY-SA-3.0


## SEE ALSO


`solbuild(1)`, `solbuild.conf(5)`

https://github.com/toml-lang/toml

## NOTES

Creative Commons Attribution-ShareAlike 3.0 Unported

 * http://creativecommons.org/licenses/by-sa/3.0/
