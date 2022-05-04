# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.2.0](https://github.com/keptn-contrib/job-executor-service/compare/0.1.8...0.2.0) (2022-05-04)


### ⚠ BREAKING CHANGES

* The `enableKubernetesApiAccess` flag is removed in favor of the serviceAccount configuration for jobs
* - The job-executor-service is moved into it's own namespace (e.g.: keptn-jes) to isolate the jobs from other Keptn services
- A valid Keptn API token is needed for the job-executor-service to function properly
- A more restrictive service account is used for jobs

Signed-off-by: Raphael Ludwig <raphael.ludwig@dynatrace.com>

### Features

* Add allowlist for job images ([#213](https://github.com/keptn-contrib/job-executor-service/issues/213)) ([f3febab](https://github.com/keptn-contrib/job-executor-service/commit/f3febab8ed8791f550e5f109cf205d5b632eb263))
* Add Keptn auto-detection ([#227](https://github.com/keptn-contrib/job-executor-service/issues/227)) ([741c876](https://github.com/keptn-contrib/job-executor-service/commit/741c876c3e8da52a9df89790726234b9bc078dcf))
* Create a security context for the job-executor-service  ([#205](https://github.com/keptn-contrib/job-executor-service/issues/205)) ([17b58a7](https://github.com/keptn-contrib/job-executor-service/commit/17b58a7ce5d905f5d07478b5148c663beb216b7c))
* Introduce serviceAccount for job workloads ([#223](https://github.com/keptn-contrib/job-executor-service/issues/223)) ([1192649](https://github.com/keptn-contrib/job-executor-service/commit/119264941f2081d9552a2917aba695c74f3fcccf))
* Job security context ([#221](https://github.com/keptn-contrib/job-executor-service/issues/221)) ([9185e8e](https://github.com/keptn-contrib/job-executor-service/commit/9185e8e3ec1ec1dbdf6070ded245c102208d362f))
* Move job-executor-service to it's own namespace ([#207](https://github.com/keptn-contrib/job-executor-service/issues/207)) ([8139bd5](https://github.com/keptn-contrib/job-executor-service/commit/8139bd5d228bca1686fbaff752da62290584e141))
* Restrict service account of jobs ([#204](https://github.com/keptn-contrib/job-executor-service/issues/204)) ([07dd337](https://github.com/keptn-contrib/job-executor-service/commit/07dd33713383d264b9f2627aad96df0737e3b975))
* send error log when error occurs before starting any job ([5768b46](https://github.com/keptn-contrib/job-executor-service/commit/5768b46cf4ea5a90024db59bddd2fbbfaa30364e))
* Upgrade to Keptn 0.13 ([#228](https://github.com/keptn-contrib/job-executor-service/issues/228)) ([c287632](https://github.com/keptn-contrib/job-executor-service/commit/c287632ccc865e33b8798cfb365d6f83a21fb49a))


### Docs

* fixed documentation for the usage of labels ([#216](https://github.com/keptn-contrib/job-executor-service/issues/216)) ([cb3f9f5](https://github.com/keptn-contrib/job-executor-service/commit/cb3f9f534f146e9123c0597b2e682266ce6a6b91))
* Provide incompatibility warning for Keptn 0.14.x ([#218](https://github.com/keptn-contrib/job-executor-service/issues/218)) ([4cb3380](https://github.com/keptn-contrib/job-executor-service/commit/4cb3380b8f529dfe0486ab053696d26cebec0ed3))
* updated compatibility matrix ([#211](https://github.com/keptn-contrib/job-executor-service/issues/211)) ([0ef5cb2](https://github.com/keptn-contrib/job-executor-service/commit/0ef5cb2e89318962a0da1add00c49194ff875277))


### Refactoring

* remove prometheus dependency ([#232](https://github.com/keptn-contrib/job-executor-service/issues/232)) ([5ab969c](https://github.com/keptn-contrib/job-executor-service/commit/5ab969c9ffe9d258860ef4b350c67d23cd43d514))
* separate event data mapping from handling and remove redundant EventHandler attributes ([dfe009d](https://github.com/keptn-contrib/job-executor-service/commit/dfe009d0ee1b01f8bc0fb54592f5c8a3f3cd0d9e))

### [0.1.8](https://github.com/keptn-contrib/job-executor-service/compare/0.1.7...0.1.8) (2022-03-30)


### Features

* Add labels to environment variables ([#185](https://github.com/keptn-contrib/job-executor-service/issues/185)) ([43cee8d](https://github.com/keptn-contrib/job-executor-service/commit/43cee8d1ed0b810c2fa554363faf4c8273a55487))
* Implement liveness and readiness endpoints ([#197](https://github.com/keptn-contrib/job-executor-service/issues/197)) ([0bcd9a6](https://github.com/keptn-contrib/job-executor-service/commit/0bcd9a6eb6bdaa5cd3c83c2ac74b8f31d2d8cfdd))
* Use pullPolicy ifNotPresent for initContainer ([#191](https://github.com/keptn-contrib/job-executor-service/issues/191)) ([#196](https://github.com/keptn-contrib/job-executor-service/issues/196)) ([71716b7](https://github.com/keptn-contrib/job-executor-service/commit/71716b7563adfe60d3a327c1f81d1fdb63dc1ebf))


### Bug Fixes

* Allow handling of event types other than *.triggered ([#182](https://github.com/keptn-contrib/job-executor-service/issues/182)) ([80a49d7](https://github.com/keptn-contrib/job-executor-service/commit/80a49d7769036f3e1b192e1b130dc7438d0be59d))


### Other

* Release notes for 0.1.7 ([#180](https://github.com/keptn-contrib/job-executor-service/issues/180)) ([146257e](https://github.com/keptn-contrib/job-executor-service/commit/146257e4849667af8e891e67d89b6a5801d515c6))


### Docs

* Add guidance for updating api-token, topic subscription to README ([#200](https://github.com/keptn-contrib/job-executor-service/issues/200)) ([6e99d57](https://github.com/keptn-contrib/job-executor-service/commit/6e99d57fa46a5abb90f66265631cb1350d20343b))
* added architecture diagrams ([#172](https://github.com/keptn-contrib/job-executor-service/issues/172)) ([1814bc0](https://github.com/keptn-contrib/job-executor-service/commit/1814bc0f80923dde2d0c00162ebb3f27bac95c30))
* Added migration guide for generic-executor-service ([#199](https://github.com/keptn-contrib/job-executor-service/issues/199)) ([f1eb946](https://github.com/keptn-contrib/job-executor-service/commit/f1eb946d671de336786e6554455e54b32dc9fe59))
* Moved releasenotes into CHANGELOG.md ([#203](https://github.com/keptn-contrib/job-executor-service/issues/203)) ([c8cbe5e](https://github.com/keptn-contrib/job-executor-service/commit/c8cbe5eb51c18b52b40aa1aac7d5cb04c7647f7b))

### [0.1.7](https://github.com/keptn-contrib/job-executor-service/compare/0.1.6...0.1.7) (2022-02-28)


### Features

* Update job-executor to Keptn 0.12 ([#173](https://github.com/keptn-contrib/job-executor-service/issues/173)) ([3393b00](https://github.com/keptn-contrib/job-executor-service/commit/3393b0039959c51bc889525ce06c9dd9ac039bdf))


### Docs

* Move features to FEATURES.md, restructure README ([#166](https://github.com/keptn-contrib/job-executor-service/issues/166)) ([4557ca2](https://github.com/keptn-contrib/job-executor-service/commit/4557ca2dc0dc5f9f4aeb69b986cc4487b324c8c2))


### Other

* prepare release 0.1.7 ([#175](https://github.com/keptn-contrib/job-executor-service/issues/175)) ([8f33747](https://github.com/keptn-contrib/job-executor-service/commit/8f337470279325df499dc54ee4eaeea24496d4ae))

### [0.1.6](https://github.com/keptn-contrib/job-executor-service/compare/0.1.5...0.1.6) (2022-01-25)


### Bug Fixes

* clean-up pods when deleting kubernetes jobs ([#154](https://github.com/keptn-contrib/job-executor-service/issues/154)) ([#155](https://github.com/keptn-contrib/job-executor-service/issues/155)) ([38c9f19](https://github.com/keptn-contrib/job-executor-service/commit/38c9f19886251b5bbdd06a6391db1c146d90ce29))


### Docs

* Added docs regarding remote execution plane ([#153](https://github.com/keptn-contrib/job-executor-service/issues/153)) ([6735d50](https://github.com/keptn-contrib/job-executor-service/commit/6735d504426de006137c055447d1b126860ba7b0))

### [0.1.5](https://github.com/keptn-contrib/job-executor-service/compare/0.1.4...0.1.5) (2022-01-13)


### Features

* **core:** Add kubernetes api access to jobs ([#146](https://github.com/keptn-contrib/job-executor-service/issues/146)) ([b89be8b](https://github.com/keptn-contrib/job-executor-service/commit/b89be8b27df2959bf08cbd548ee1f9266562f287))
* **core:** support imagePullPolicy for tasks([#127](https://github.com/keptn-contrib/job-executor-service/issues/127)) ([#135](https://github.com/keptn-contrib/job-executor-service/issues/135)) ([022cdfe](https://github.com/keptn-contrib/job-executor-service/commit/022cdfef9645521b5337dae82b017ca3c164c65b))
* **core:** upgrade to keptn 0.10.0 ([#107](https://github.com/keptn-contrib/job-executor-service/issues/107)) ([8045612](https://github.com/keptn-contrib/job-executor-service/commit/80456129634a6e55d9c8297b403b5cb8ed26066e))


### Docs

*  Document ttlSecondsAfterFinished ([#126](https://github.com/keptn-contrib/job-executor-service/issues/126)) ([#147](https://github.com/keptn-contrib/job-executor-service/issues/147)) ([d93645c](https://github.com/keptn-contrib/job-executor-service/commit/d93645c367ed024a3a3ed1eee93347c61cf7c70d))
* add short explanation on howto add the job/config.yaml to the keptn repository for a specific service ([62bdcd8](https://github.com/keptn-contrib/job-executor-service/commit/62bdcd8ee69f7dc2dcc4363dc7fcef33b4a85e1f))
* add short explanation on howto add the job/config.yaml to the keptn repository for a specific service ([#79](https://github.com/keptn-contrib/job-executor-service/issues/79)) ([44c0cba](https://github.com/keptn-contrib/job-executor-service/commit/44c0cbaa26ad67a220710431c48b1678dd8c90d2))


### Other

* add christian-kreuzberger-dtx as a codeowner ([#105](https://github.com/keptn-contrib/job-executor-service/issues/105)) ([68167ff](https://github.com/keptn-contrib/job-executor-service/commit/68167ff0c304570630ca56b0a510521f19863209))
* Add Gilbert Tanner, Gabriel Tanner and Paolo Chila as Codeowners ([#128](https://github.com/keptn-contrib/job-executor-service/issues/128)) ([b65263d](https://github.com/keptn-contrib/job-executor-service/commit/b65263d4f36240c456994d22a31f8ab425bc1c0e))
* Added new release integration workflow ([#150](https://github.com/keptn-contrib/job-executor-service/issues/150)) ([d4e5b98](https://github.com/keptn-contrib/job-executor-service/commit/d4e5b989517aea8a59b1ee897e8c142359f3202a))
* Added semantic PR check ([02807ed](https://github.com/keptn-contrib/job-executor-service/commit/02807edb79cd5629bba99a8598370e5db83fa8dc))
* Enable skaffold to build and deploy initcontainer ([#139](https://github.com/keptn-contrib/job-executor-service/issues/139)) ([1783653](https://github.com/keptn-contrib/job-executor-service/commit/1783653b449d4c420dba2223dd529aac571802fe))
* Introduce new CI and pre-release pipeline, level up helm charts ([#112](https://github.com/keptn-contrib/job-executor-service/issues/112)) ([cc1da80](https://github.com/keptn-contrib/job-executor-service/commit/cc1da80f4310f1a5d60d4b3d119cc6ca85d50a25))
* remove dependabot, add renovate ([#113](https://github.com/keptn-contrib/job-executor-service/issues/113)) ([c1fbc8b](https://github.com/keptn-contrib/job-executor-service/commit/c1fbc8b0ceb06ef1a7e0798f901039d7d3c7f2e4))
* restructure docs, add new installation method ([#149](https://github.com/keptn-contrib/job-executor-service/issues/149)) ([f5522d4](https://github.com/keptn-contrib/job-executor-service/commit/f5522d4968958fba84e0a4a84a5f24392ddfeae2))
* Use validate-semantic-pr workflow from keptn/gh-automation repo ([#103](https://github.com/keptn-contrib/job-executor-service/issues/103)) ([c4e7a97](https://github.com/keptn-contrib/job-executor-service/commit/c4e7a9725edd4dd0fed79971cc5184d55a5185db))


### [0.1.4](https://github.com/keptn-contrib/job-executor-service/compare/0.1.3...0.1.4) (2021-09-14)


Compatible with Keptn 0.9.0

### Features

* Add environment setting to always send a started/finished event on job config errors (#52, #57)
    * (thanks @thschue for the contribution)
* Event data formatting (#59, #63, #65)
    * (thanks @TannerGabriel for the contribution)
* Add start and end event metadata for test finished events (#19, #64)
* Add support for running jobs in a different namespace (#53, #67, #73)
    * (thanks @thschue for the contribution)
* With each release the helm chart is packaged and added to the assets (#77)

### Fixed Issues

* Uniform registration for remote execution planes over https doesn't work (https://github.com/keptn/keptn/issues/4516)
* Display correct timeout value in job timeout error message (#49, #61)


### [0.1.3](https://github.com/keptn-contrib/job-executor-service/compare/0.1.2...0.1.3) (2021-07-16)

Compatible with Keptn 0.8.6

Starting with this release a binary for checking job configurations is attached to each release (see https://github.com/keptn-sandbox/job-executor-service#how-to-validate-a-job-configuration)

### Features

* Allow array of strings for command, add args that are also passed through to the kubernetes job (#31)
* Provide a cli tool that validates job configurations (#33)
* Support env variables from string (#34, #36)
* Allow setting the working directory of a kubernetes job (#38)
* Configurable job timeout (#40, #43)

### Fixed Issues

* Fix kubernetes labels used by distributor for uniform registration (#32)

### Known Limitations

* Uniform registration for remote execution planes over https doesn't work (https://github.com/keptn/keptn/issues/4516)

### [0.1.2](https://github.com/keptn-contrib/job-executor-service/compare/0.1.1...0.1.2) (2021-06-24)


Compatible with Keptn 0.8.4

### Features

* Reference kubernetes secrets as environment variables in tasks (#8, #15)
* Configurable resource quotas (#18, #27)
* Specifying a directory under task files now imports all files of this directory (#28, #29)
* Configuration for uniform registration feature of distributor (#22)


### Known Limitations

* Uniform registration for remote execution planes over https doesn't work

### [0.1.1](https://github.com/keptn-contrib/job-executor-service/compare/0.1.0...0.1.1) (2021-06-07)

Fixes some issues with wrong tags or versions for images. Adds the possibility to execute actions in silent mode, meaning no `started` or `finished` events are sent.

### Features

* Silent mode for actions #6

### Fixed Issues

* Correct tags for images #5, #9

### 0.1.0

This is the initial implementation of job-executor-service, compatible with Keptn 0.8.3.
