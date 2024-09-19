# Overview

LBX is software for displaying photos online. It is intended to be a self-hosted alternative to services such as Smugmug and Photobucket, and to serve as a “forever home” for an individual or family’s photo (and video) collection.

# Design goals

## Durability

* It must be easy to set up and to migrate.
* It must require minimal maintenance as technologies evolve.
* It must have low ongoing costs (the cost of required storage + epsilon).
* It must support stable, user-defined paths for albums and media.
* Code should be publicly available (as source code and a Docker image) so that people with sufficient knowledge can run and deploy it on a system with a public IP address.

## Efficient file-based workflow
LBX must make it trivial to manage and display media stored on a local (henceforth “master”) filesystem, such as exports from Adobe Lightroom. It must be easy to upload or update entire directory trees of media.

## Beauty

LBX should provide a clutter-free and responsive interface.

# Primary design requirements
The design goals dictate the following requirements.

## Simplicity
LBX must have minimal technology dependencies and moving parts. To the extent that it uses frameworks or depends on external systems (e.g., databases), those should be mainstream and well-supported.

## Storage flexibility
LBX must support both Unix-like filesystems and S3-like bucket storage for media.

## Storage scalability
LBX should provide consistently low-latency responses for large photo collections (1000s albums, 100Ks photos, ~1TB+ storage) and large single albums (1000s of photos).

However, since it is intended for private media collections, peak throughput need not be very high. We expect <100 concurrent users and <1000 concurrent image loads.

(Concretely, LBX should run on a single VM or serverless instance with a lot of storage.)

## Easy migration
It should be possible to package and move an LBX instance from one cloud vendor to another or to switch between file system storage and bucket storage easily and in a manner that preserves media paths.

## Backup
LBX metadata must be easily backed up via common mechanisms such as rsync or arq.

LBX is not responsible for media backup. An LBX installation is already a backup of media on the master filesystem.
