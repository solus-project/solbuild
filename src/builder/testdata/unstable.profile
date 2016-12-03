# TODO: Flesh out the profile configuration =P
image = "unstable-x86_64"

# Remove all the repos from the base image
# remove_repos = ['*']

# Remove just a single repo from the base image
# remove_repos = ['Solus']

# Restrict enabled repos to just one repo
add_repos = ["Solus"]

# Example of adding a remote repo
[repo.Solus]
uri = "http://packages.solus-project.com/unstable/eopkg-index.xml.xz"

# Add a local repository by bind mounting it into chroot on each build
[repo.Local]
uri = "/var/lib/myrepo"
local = true

# A local repo with automatic indexing
[repo.LocalIndexed]
uri = "/var/lib/myOtherRepo"
local = true
autoindex = true
