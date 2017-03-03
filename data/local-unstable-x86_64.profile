#
# local-unstable-x86_64 configuration
#
# Build Solus packages using the unstable repository image.
# This is the default profile for the Solus build server and developers.
#
# Do not make changes to this file. solbuild is implemented in a stateless
# fashion, and will load files in a layered mechanism. If you wish to edit
# this profile, copy to /etc/solbuild/.
#
# It is generally advisable to create a *new* profile name in /etc, because
# we will load /etc/ before /usr/share. Thus, profiles with the same name
# in /etc/ are loaded *first* and will override this profile.
#
# Of course, if that's what you intended to do, then by all means, do so.

image = "unstable-x86_64"

# If you have a local repo providing packages that exist in the main
# repository already, you should remove the repo, and re-add it *after*
# your local repository:
remove_repos = ['Solus']
add_repos = ['Local','Solus']

# A local repo with automatic indexing
[repo.Local]
uri = "/var/lib/solbuild/local"
local = true
autoindex = true

# Re-add the Solus unstable repo
[repo.Solus]
uri = "https://packages.solus-project.com/unstable/eopkg-index.xml.xz"
