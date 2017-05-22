btrfscue
========

btrfscue is an advanced data recovery tool for the BTRFS filesystem. Despite
being a state of the art filesystem, at the time when I started writing this
(Q2 2011), BTRFS did not have a stable fsck tool that is capable of restoring
a filesystem to a mountable state after a power failure or system crash.
Recently, this situation has somewhat improved with the `btrfs restore`
command. Unlike this official tool, btrfscue is designed to be able to restore
data from disk images that were obtained from faulty storage devices or if all
superblocks were overwritten inadvertently.

Being a recovery tool, btrfscue works best on disk images and will write
recovered data to a directory. It can thus be used to convert BTRFS filesystems
to any other filesystem supported by the host OS. It will also recover recently
deleted files and directories and aid in BTRFS filesystem forensics.


Table of Contents
-----------------

   * [btrfscue](#btrfscue)
      * [Development State](#development-state)
      * [Requirements](#requirements)
      * [Recommended Tools](#recommended-tools)
      * [How to Build](#how-to-build)
         * [Build using Make](#build-using-make)
      * [Packages](#packages)
      * [Usage](#usage)
      * [Copyright/License](#copyrightlicense)


Development State
-----------------

As the version number 0.3 implies, this software is pretty much in alpha state.
In fact, the repository you're looking at now is a complete rewrite of an
earlier attempt that was written in C++ as early as 2011 (so don't let the
copyright years fool you :)).

This works:
  - Heuristic detection of filesystem identifiers
  - Dump meta data to file
  - Listing of files and directories in the metadata
  - FUSE-mounting a "rescue" view of the metadata

This definitely does not work:
  - Actually restoring files bigger than the filesystem block size
  - Running on big-endian machines
  - BTRFS RAID levels, multi-device FS. These are planned for later.


Requirements
------------

  - Go 1.8 or higher
  - Git version 1.7 or later
  - Optional: CDBS (to build the Debian packages)
  - Optional: GNU Make


Recommended Tools
-----------------

  - btrfs-tools to inspect the faulty filesystem before attempting data recovery
  - ddrescue to create disk images of faulty drives (download from
    http://www.gnu.org/software/ddrescue/ddrescue.html or, on Debian, install
    the package gddrescue)


How to Build
------------

General way to build from source via `go get`:
```
go get blichmann.eu/code/btrfscue
```

### Build using Make

To build from a specific revision/branch/tag, not using `go get`:
```bash
mkdir -p btrfscue && cd btrfscue
git clone https://github.com/cblichmann/btrfscue.git .
# Optional: checkout a specific rev./branch/tag using i.e. git checkout
make
```

You may want to create a symlink to the binary somewhere in your path.


Packages
--------

At the moment, only building Debian packages is supported. Just run `make deb`
to build.


Usage
-----

btrfscue command-line syntax is generally as follows:
```
btrfscue SUBCOMMAND OPTION...
```

Data recovery with btrfscue is divided in stages:

  1. If you suspect physical damage, use a tool like ddrescue to dump the
     contents of the damaged filesystem to another disk. Otherwise, the
     standard `dd` utility will do just fine. The following steps assume the
     disk image is named DISKIMAGE.
     If you don't have enough physical storage space, btrfscue will also
     directly work with the device file. However, **THIS IS NOT RECOMMENDED
     IN CASE OF SUSPECTED PHYSICAL DAMAGE**. Although btrfscue never writes
     to the device, it may stress the drive too much and may render further
     recovery attempts impossible. This is even true of damaged SSDs since
     the flash controller may decide at any time to shutdown the device for
     good.

  2. Build a list of possible ids to help identify the filesystem id for the
     filesystem that is to be restored by applying a heuristic. This will
     output a list of filesystem ids along with the number of times the
     respective id was found while sampling the disk image.
     ```
     btrfscue identify DISKIMAGE
     ```
  3. Save metadata for later analysis. This may take a long time to finish
     as the whole image is being scanned. You need to specify the filesystem
     to look for by using the --id parameter with a filesystem id FSID.
     ```
     btrfscue recon --id FSID --metadata metadata.db DISKIMAGE
     ```
  4. Inspect the metadata dump to help decide what to restore later.
     ```
     btrfscue --metadata metadata.db ls /
     ...
     ```
     Alternatively, if you're on Linux or macOS, you can FUSE-mount a "rescue"
     of the filesystem metadata:
     ```
     btrfscue --metadata metadata.db mount MOUNTPOINT
     ```
     Explore the metadata from another shell. Type CTRL+C to unmount.

  5. Restore the actual data. This is work-in-progress. You can use the mount
     command to copy files that are no bigger than the filesystem block size.


Copyright/License
-----------------

btrfscue version 0.4
Copyright (c)2011-2017 Christian Blichmann <mail@blichmann.eu>

btrfscue is licensed under a two-clause BSD license, see the LICENSE file
for details.
