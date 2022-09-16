# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.3.0](https://github.com/keptn-contrib/job-executor-service/compare/0.2.5...0.3.0) (2022-09-16)


### ⚠ BREAKING CHANGES

* Reimplement service with go-sdk (#351)

### Features

* Reimplement service with go-sdk ([#351](https://github.com/keptn-contrib/job-executor-service/issues/351)) ([91c9409](https://github.com/keptn-contrib/job-executor-service/commit/91c940971dd01d641ddeb16ea9044e21520dd799))
* Send error cloud event ([#354](https://github.com/keptn-contrib/job-executor-service/issues/354)) ([e247419](https://github.com/keptn-contrib/job-executor-service/commit/e247419c6f10ac4f2fa623b1a57bee3aabaf9914))


### Other

* Delete NoIngress integration test ([#353](https://github.com/keptn-contrib/job-executor-service/issues/353)) ([cbe07b7](https://github.com/keptn-contrib/job-executor-service/commit/cbe07b792562c6400517d8eaca3445ef0de0eca8))
* Release notes for 0.2.5 ([#350](https://github.com/keptn-contrib/job-executor-service/issues/350)) ([f772d0c](https://github.com/keptn-contrib/job-executor-service/commit/f772d0c1282da8ba6414403c40a720516880dfcf))
* Update dependencies ([#356](https://github.com/keptn-contrib/job-executor-service/issues/356)) ([6b613b1](https://github.com/keptn-contrib/job-executor-service/commit/6b613b1aed0910e8cbb50c0eb4cb3d658f401caf))
* Update to Go 1.18 ([#357](https://github.com/keptn-contrib/job-executor-service/issues/357)) ([1353261](https://github.com/keptn-contrib/job-executor-service/commit/1353261872708207b55d152db462105d92c3a746))

### [0.2.5](https://github.com/keptn-contrib/job-executor-service/compare/0.2.4...0.2.5) (2022-08-29)


### Features

* Keptn 0.17 compatibility ([#310](https://github.com/keptn-contrib/job-executor-service/issues/310)) ([e9aefac](https://github.com/keptn-contrib/job-executor-service/commit/e9aefacb3a68817c2f2885df6da171efe770d3e0))

This release introduces `API_PROXY_MAX_PAYLOAD_BYTES_KB`, an environment variable for the Keptn distributor. It is currently set to 128 (Kilobyte) and defines the maximum payload size that job-executor can send back to Keptn, i.e., the log output of your job (+ a bit of overhead). Should your job create more log output, you have to increase the `API_PROXY_MAX_PAYLOAD_BYTES_KB` environment variable in the Kubernetes manifest.


### Docs

* Update compatibility matrix ([#345](https://github.com/keptn-contrib/job-executor-service/issues/345)) ([eeffbaf](https://github.com/keptn-contrib/job-executor-service/commit/eeffbaf4c85fbb6fadb42ab97346353e8f5da4d4))

### [0.2.4](https://github.com/keptn-contrib/job-executor-service/compare/0.2.3...0.2.4) (2022-08-11)


### Features

* Ability to specify annotations for jobs ([#302](https://github.com/keptn-contrib/job-executor-service/issues/302)) ([487c2a3](https://github.com/keptn-contrib/job-executor-service/commit/487c2a3ce4f769406af4faf0cd184015ba62186b))
* Add global and stage job configuration lookup ([#338](https://github.com/keptn-contrib/job-executor-service/issues/338)) ([c146449](https://github.com/keptn-contrib/job-executor-service/commit/c1464492b09039bc8365516a7c6910b98ed591b9))
* Add Keptn dev version to integration tests ([#314](https://github.com/keptn-contrib/job-executor-service/issues/314)) ([db5686a](https://github.com/keptn-contrib/job-executor-service/commit/db5686a4bb100557fb440520d2161af85a354568))
* Post integration test summary to GH workflow ([#312](https://github.com/keptn-contrib/job-executor-service/issues/312)) ([956422f](https://github.com/keptn-contrib/job-executor-service/commit/956422f8005d5c288dfa2ff42711808b04b026d6))
* Use helm build action ([#305](https://github.com/keptn-contrib/job-executor-service/issues/305)) ([5507b45](https://github.com/keptn-contrib/job-executor-service/commit/5507b45ea3c83431dd6de783f5de2e4880e6fe1e))
* Utilize gitCommitID from cloud events to fetch resources  ([#303](https://github.com/keptn-contrib/job-executor-service/issues/303)) ([5187bd9](https://github.com/keptn-contrib/job-executor-service/commit/5187bd9b4654efa45182ef0b414ea2af0191e792))


### Bug Fixes

* Auto-detection of Keptn 0.17.0 ([#319](https://github.com/keptn-contrib/job-executor-service/issues/319)) ([c9667b1](https://github.com/keptn-contrib/job-executor-service/commit/c9667b17d12effd45489e7dae4d57cb248468115))
* Handling of directories in the init container  ([#309](https://github.com/keptn-contrib/job-executor-service/issues/309)) ([7bd83dc](https://github.com/keptn-contrib/job-executor-service/commit/7bd83dcd3ec62c577dd1279773b854b81c03481e))
* Sending error logs to all registered jes instances in uniform ([#334](https://github.com/keptn-contrib/job-executor-service/issues/334)) ([17ee4e2](https://github.com/keptn-contrib/job-executor-service/commit/17ee4e2df6a2564d695738612482f8e3a8562b23))


### Other

* Improve integration tests log output ([#320](https://github.com/keptn-contrib/job-executor-service/issues/320)) ([8a9bcbd](https://github.com/keptn-contrib/job-executor-service/commit/8a9bcbd678ec9ba92647f282abedb4a04f33b4a8))
* Remove kubernetes-utils dependency ([#304](https://github.com/keptn-contrib/job-executor-service/issues/304)) ([c69cbec](https://github.com/keptn-contrib/job-executor-service/commit/c69cbecdaf025207244ee10d79e27636657efcff))
* Update pipeline to be compatible with Keptn 0.17 ([#327](https://github.com/keptn-contrib/job-executor-service/issues/327)) ([cd649c8](https://github.com/keptn-contrib/job-executor-service/commit/cd649c88851dd00d17da920182ecac532081389c))


### Docs

* Improve upgrade guide and breaking change documentation ([#331](https://github.com/keptn-contrib/job-executor-service/issues/331)) ([ebfdd65](https://github.com/keptn-contrib/job-executor-service/commit/ebfdd65d565d9c2ff0be8535c58f305351beb3a6))
* Update JES version in installation docs ([#300](https://github.com/keptn-contrib/job-executor-service/issues/300)) ([934728f](https://github.com/keptn-contrib/job-executor-service/commit/934728f30b495ed5e04c608d3245eecfce9290fb))

### [0.2.3](https://github.com/keptn-contrib/job-executor-service/compare/0.2.2...0.2.3) (2022-07-01)


### Features

* Upgrade dependencies to Keptn 0.16 ([#295](https://github.com/keptn-contrib/job-executor-service/issues/295)) ([7f71c62](https://github.com/keptn-contrib/job-executor-service/commit/7f71c620a297cb0f6ec40b73c51bc508a4891f54))


### Bug Fixes

* Integration tests debug archive ([#291](https://github.com/keptn-contrib/job-executor-service/issues/291)) ([4cdb40b](https://github.com/keptn-contrib/job-executor-service/commit/4cdb40b7f25559f535130e9daa0532cade482a75))


### Docs

* document ingress/egress job-executor-service network policies ([#293](https://github.com/keptn-contrib/job-executor-service/issues/293)) ([ecccad7](https://github.com/keptn-contrib/job-executor-service/commit/ecccad7f4ba64fb321120f5670b63cc975c3f53a))

### [0.2.2](https://github.com/keptn-contrib/job-executor-service/compare/0.2.1...0.2.2) (2022-06-21)


### Features

* Network policy for jobs ([#276](https://github.com/keptn-contrib/job-executor-service/issues/276)) ([51cb291](https://github.com/keptn-contrib/job-executor-service/commit/51cb2913d9718340202ca9c42ad5a56a87af724b))
* Upgrade to Keptn 0.15.1 ([#282](https://github.com/keptn-contrib/job-executor-service/issues/282)) ([a558325](https://github.com/keptn-contrib/job-executor-service/commit/a55832504505102dbe62e5c1860368717d073fd2))


### Bug Fixes

* fix typo in network policy helm template ([#281](https://github.com/keptn-contrib/job-executor-service/issues/281)) ([bd0d6e4](https://github.com/keptn-contrib/job-executor-service/commit/bd0d6e4a62f9ecae4f17d33b070116134cc60ce0))

### [0.2.1](https://github.com/keptn-contrib/job-executor-service/compare/0.2.0...0.2.1) (2022-06-21)


### Features

* Add job labels ([#240](https://github.com/keptn-contrib/job-executor-service/issues/240)) ([5c6911d](https://github.com/keptn-contrib/job-executor-service/commit/5c6911da7e189b8ffd662d9c8ea792a7f79c0b28))
* Add output to go test in pipeline ([#237](https://github.com/keptn-contrib/job-executor-service/issues/237)) ([effe9fe](https://github.com/keptn-contrib/job-executor-service/commit/effe9fe275ea1160cd75fef901fae047cdb637f3))
* Enforce minimum job TTL value ([#241](https://github.com/keptn-contrib/job-executor-service/issues/241)) ([4064ee9](https://github.com/keptn-contrib/job-executor-service/commit/4064ee99ce90b2582d07f08dee4696650257c84e))
* Include logs of all containers in error message ([#214](https://github.com/keptn-contrib/job-executor-service/issues/214)) ([a58c2cb](https://github.com/keptn-contrib/job-executor-service/commit/a58c2cb1c0a8c27ee08598f610e0dbabe7c2815b))
* limit job executor service network access ([6da2cac](https://github.com/keptn-contrib/job-executor-service/commit/6da2cac119758c1e843f26f896fa4b4c8762174a))
* limit job run time ([2045058](https://github.com/keptn-contrib/job-executor-service/commit/20450583fbc65a65948347fc2eb5c43d1709c9af))
* OAuth authentication mode ([#265](https://github.com/keptn-contrib/job-executor-service/issues/265)) ([1126cf5](https://github.com/keptn-contrib/job-executor-service/commit/1126cf5ff439625a3e3f5cc366cee13559eb9f9a))
* Upgrade to Keptn 0.14 ([#275](https://github.com/keptn-contrib/job-executor-service/issues/275)) ([642e2a9](https://github.com/keptn-contrib/job-executor-service/commit/642e2a9c493eae43bae525353595352c3ff61607))


### Bug Fixes

* Add output of failed events to logs ([#249](https://github.com/keptn-contrib/job-executor-service/issues/249)) ([2ca699e](https://github.com/keptn-contrib/job-executor-service/commit/2ca699edd9ba18bd747ecd51f9b572c78b39da7a))
* separate ingress and egress network policy ([#273](https://github.com/keptn-contrib/job-executor-service/issues/273)) ([4a6c013](https://github.com/keptn-contrib/job-executor-service/commit/4a6c013f2dfd4d7c41d0fe296aed034e2e000829))


### Other

* set helm chart version to 0.0.0-dev ([#248](https://github.com/keptn-contrib/job-executor-service/issues/248)) ([0a26c13](https://github.com/keptn-contrib/job-executor-service/commit/0a26c1393d2789c9f98301d1b7bac360a7f9baaf))
* Update Keptn versions in integration tests ([#247](https://github.com/keptn-contrib/job-executor-service/issues/247)) ([3408bee](https://github.com/keptn-contrib/job-executor-service/commit/3408beeedc33fa439c056c620f5b32b166971787))


### Docs

* Add OAuth installation instructions ([#274](https://github.com/keptn-contrib/job-executor-service/issues/274)) ([092e2a7](https://github.com/keptn-contrib/job-executor-service/commit/092e2a788779bacc3c5d2d59f2c00725929b410c))
* polish installation upgrade instructions ([d072f83](https://github.com/keptn-contrib/job-executor-service/commit/d072f832b88ddda095027c93f35eb518b0340de8))
* remove `Always send finished event` documentation and configmap settings ([68818a9](https://github.com/keptn-contrib/job-executor-service/commit/68818a93cd2aa5e92d06822f171db007f44c96eb))
* Update chart README to include latest changes in values.yaml ([#279](https://github.com/keptn-contrib/job-executor-service/issues/279)) ([6c4201d](https://github.com/keptn-contrib/job-executor-service/commit/6c4201d95acf08bcede5ef31b20dfc6fe057e6dc))

## [0.2.0](https://github.com/keptn-contrib/job-executor-service/compare/0.1.8...0.2.0) (2022-05-04)

:tada: This release focuses on :closed_lock_with_key: security hardening, quality assurance and refactoring.

### ⚠ BREAKING CHANGES

- :joystick:  The `enableKubernetesApiAccess` flag is removed in favor of the `serviceAccount` configuration for jobs
- :lock: The job-executor-service is moved into it's own namespace (e.g.: keptn-jes) to isolate the jobs from other Keptn services
- :key: A valid Keptn API token and Keptn API endpoint need to be configured when installing job-executor-service (it is no longer possible to connect directly to Keptn's nats-cluster)
- :robot:  A more restrictive service account is used for jobs by default
- :egg: The default value for `remotecontrolPlane.api.protocol` has been set to `http` (was `https` before). Please take special care when upgrading and specify the desired protocol.


### Features

* Add allowlist for job images ([#213](https://github.com/keptn-contrib/job-executor-service/issues/213)) ([f3febab](https://github.com/keptn-contrib/job-executor-service/commit/f3febab8ed8791f550e5f109cf205d5b632eb263))
* Add Keptn auto-detection ([#227](https://github.com/keptn-contrib/job-executor-service/issues/227)) ([741c876](https://github.com/keptn-contrib/job-executor-service/commit/741c876c3e8da52a9df89790726234b9bc078dcf))
* Create a security context for the job-executor-service  ([#205](https://github.com/keptn-contrib/job-executor-service/issues/205)) ([17b58a7](https://github.com/keptn-contrib/job-executor-service/commit/17b58a7ce5d905f5d07478b5148c663beb216b7c))
* Introduce serviceAccount for job workloads ([#223](https://github.com/keptn-contrib/job-executor-service/issues/223)) ([1192649](https://github.com/keptn-contrib/job-executor-service/commit/119264941f2081d9552a2917aba695c74f3fcccf))
* Job security context ([#221](https://github.com/keptn-contrib/job-executor-service/issues/221)) ([9185e8e](https://github.com/keptn-contrib/job-executor-service/commit/9185e8e3ec1ec1dbdf6070ded245c102208d362f))
* Move job-executor-service to it's own namespace ([#207](https://github.com/keptn-contrib/job-executor-service/issues/207)) ([8139bd5](https://github.com/keptn-contrib/job-executor-service/commit/8139bd5d228bca1686fbaff752da62290584e141))
* Restrict service account of jobs ([#204](https://github.com/keptn-contrib/job-executor-service/issues/204)) ([07dd337](https://github.com/keptn-contrib/job-executor-service/commit/07dd33713383d264b9f2627aad96df0737e3b975))
* Send error log when error occurs before starting any job ([5768b46](https://github.com/keptn-contrib/job-executor-service/commit/5768b46cf4ea5a90024db59bddd2fbbfaa30364e))
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
