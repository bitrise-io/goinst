# goinst

Go Install command line tools.

Works in isolated, temporary GO workspaces.

Does not require GOPATH to be configured,
the only dependency is `Go`, and the VCS tools the packages use (e.g. `git`).


## Usage

```
goinst PACKAGE
```

e.g.

```
goinst github.com/bitrise-tools/betriis
```

Works similar to `go get -u PACKAGE`, but in a temporary, isolated workspace,
instead of a system `GOPATH`. This means that __it will not modify the content
of `GOPATH`__ if you have one configured.

The generated binaries are moved into `/usr/local/bin` by default.

Safe to use with `sudo` (`sudo goinst PACKAGE`), as it won't modify anything outside
its isolated workspace (other than copying the generated binaries), and will
delete the workspace once it's finished.
