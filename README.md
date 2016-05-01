btrfscue
========

btrfscue is an advanced data recovery tool for the BTRFS filesystem. Despite
being a state of the art filesystem, at the time when I started writing this (Q2
2011), BTRFS did not have a stable fsck tool that is capable of restoring a
filesystem to a mountable state after a power failure or system crash. More
recently, this situation has somewhat improved with the `btrfs restore`
command. Unlike this official tool, btrfscue is designed to be able to restore
data from disk images that were obtained from faulty storage devices or if all
superblocks were overwritten inadvertedly.

Being a recovery tool, btrfscue works best on disk images and writes recovered
data to a directory. It can thus be used to convert BTRFS filesystems to any
other filesystem supported by the host OS. It can also recover recently
deleted files and directories and aid in BTRFS filesystem forensics.


Development State
-----------------

As the version number 0.3 implies, this software is pretty much in alpha state.
This works:
  - Heuristic detection of filesystem identifiers
  - Dump meta data to file
  - Listing of files and directories in the metadata

This definitely does not work:
  - Running on big-endian machines
  - BTRFS RAID levels, multi-device FS


Requirements
------------

  - Go 1.5 or higher
  - SQLite3 via [go-sqlite3](https://github.com/mattn/go-sqlite3), tested with
    3.11.1. This also need a working C compiler.
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
go get github.com/cblichmann/btrfscue/src/btrfscue
```

### Build the old-fashioned Way

To build from a specific revision/branch/tag, not using `go get`:
```bash
mkdir -p btrfscue && cd btrfscue
git clone https://github.com/cblichmann/btrfscue.git .
# Optional: checkout a specific rev./branch/tag using i.e. git checkout
eval $(make env)
make
```

You may want to create a symlink to the binary somewhere in your path.


Packages
--------

At the moment, only building Debian packages is supported. Change to the
`debian` directory and build the package by running `debuild` as usual.


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
     to look out for by using the -u parameter with a filesystem id FSID.
     ```
     btrfscue recon --id FSID --metadata metadata.sqlite DISKIMAGE
     ```
  4. Inspect the metadata dump to help decide what to restore later.
     ```
     btrfscue --metadata metadata.sqlite ls /
     btrfscue --metadata metadata.sqlite ls /#sub:default/
     btrfscue --metadata metadata.sqlite ls /#snap:a_snapshot/
     btrfscue --metadata metadata.sqlite find --type f --name '*.cpp'
     ...
     ```
  5. Restore the actual data. For example, restore all of the files,
     directories and sub-volumes from the disk image, using the metadata.
     The files are restored to the current directory, with sub-volumes and
     snapshot data placed in separate sub-directories with the respective
     filesystem entity name.
     ```
     btrfscue --metadata metadata.sqlite recover DISKIMAGE /
     ```


List syntax
-----------

When using the list or find command, prepend `/#sub:NAME/` or `/#snap:NAME/`
to the path in the disk image to operate on the sub-volume or snapshot named
NAME, respectively.


Copyright/License
-----------------

btrfscue version 0.3
Copyright (c)2011-2016 Christian Blichmann <mail@blichmann.eu>

btrfscue is licensed under a two-clause BSD license, see the LICENSE file
for details.
