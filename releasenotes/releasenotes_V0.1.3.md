# Release Notes v0.1.3

Compatible with Keptn 0.8.6

## New Features
* Allow array of strings for command, add args that are also passed through to the kubernetes job (#31)
* Provide a cli tool that validates job configurations (#33)
* Support env variables from string (#34, #36)
* Allow setting the working directory of a kubernetes job (#38)
* Configurable job timeout (#40, #43)

## Fixed Issues
* Fix kubernetes labels used by distributor for uniform registration (#32)

## Known Limitations
* Uniform registration for remote execution planes over https doesn't work (https://github.com/keptn/keptn/issues/4516)