# <img src="./images/nxtools_logo.png" alt="nxtools logo" height="256" width="512" />
___

This tool is a CLI-driven client to Nexus Repository Manager 3 servers.<br>It will allow:
- authentication
- blob store ops (list, delete, create)
- user + role ops (list, delete, create, edit)
- repo ops (list, delete, create, edit, upload+download package)
- more to come

## Build/install requirements

You have three alternatives:
- [Install from source](#Install from source)
- [Install from a binary package](#Install from a binary package)
- [Build your own APK, DEB, RPM packages, then manually install those packages](#Build your own package)

Installing from source requires a bit more work in the sense that GO has to be installed on your system

### Install from source
1. Clone/fork the repo : either `git clone https://github.com/jeanfrancoisgratton/nxtools` or `git clone https://git.famillegratton.net:3000/devops/nxtools`
2. Ensure that you have the proper GO version, as stated in the `go.version` file in the root of the repo. Your GO version should be equal or higher than the one in that file. To ensure, run `go version`
3. cd to `src`, and then run: `./updateBuildDeps.sh`, to ensure that all build dependencies are up to date; this might be overkill, but I always run it nonetheless
4. run `./build.sh`. By default the binary will be created in /opt (check the dir's permission ahead of running it). Examine that script, you can taylor the output as you see fit

### Install from a binary package
The simplest way : just go in the RELEASES tab of the repo, select your format, download it, and then install throught you package manager

### Build your own package
The scripts and files (__alpine/, __debian, nxtools.spec) are there for my own ease of work; I usually build my tools using "builder containers" for each format: `apkbuilder`, `debbuilder`, `rpmbuilder`
I'll leave you with homeworks, and will show you how to roughly reproduce my environment

**The three methods below assume that you have forked (not just cloned) the repo somewhere**

#### ALPINE (APKBUILDER)
1. In an Alpine container or VM, you need the following packages: `abuild-doc pax-utils git alpine-sdk`. Some other packages might be needed, depending on the config in __alpine/APKBUILD
2. From the `__alpine`, run: `abuild -r`

This should give you an Alpine package

#### DEBIAN (DEBBUILDER)
1. cd to `__debian`
2. Besides binutils, you do not need any specific package, and of course the required GO version. Have a look at `../go.version`, and `./1.install-build-deps.sh`.
3. Run `./2.build_binary.sh`
4. Copy the .deb file in a safe space, then run `./restore_repo.sh`

#### RPM (RPMBUILDER)
**FORK OR COPY the repo, do not CLONE** it; there's a step there that would fail, otherwise (see step #4)
1. Ensure that tito is installed; the easy way is with pip: `pip install tito`
2. Ensure that all other build deps are installed; from the nxtools root directory, run: `./rpmbuild-deps.sh`
3. Run the following: `tito tag --keep-version`
4. Run the following: `git push --follow-tags origin` --> **This has to be done from a forked repo, otherwise if you point at my own repo, it will likely fail**
5. Run the following: `tito build --rpm` : the result will be in /tmp/tito/ copy the files (SRPM, RPM) in a safe place

## Blob operations
We support add, remove and list operations; update operations are not yet implemented. The current blob subcommands are:
<img src="./images/blobs_-h.png" alt="nxtools blobs -h"/>

### List blobs
Very simply: `nxtools blob ls`<br><br>
<img src="./images/blobs_ls.png" alt="nxtools blobs ls"/>

*A note about the Path column* : the column will not show any data for paths with relative values (that is: the blobstore path uses its default value, inside Nexus' $DATA_DIR)


### Delete blobs
`nxtools blob rm $BLOBSTORE_NAME`, as shown below<br><br>
<img src="./images/blobs_ls-rm-ls.png" alt="nxtools blobs rm"/>


### Create blob stores

## Repositories operations
We support add, remove, list operations

### List repos
Again, very simply: `nxtools repos ls`
**IMAGE TO COME**
