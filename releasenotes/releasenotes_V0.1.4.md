# Release Notes v0.1.4

Compatible with Keptn 0.9.0

## New Features
* Add environment setting to always send a started/finished event on job config errors (#52, #57)
  * (thanks @thschue for the contribution)
* Event data formatting (#59, #63, #65)
  * (thanks @TannerGabriel for the contribution)
* Add start and end event metadata for test finished events (#19, #64)
* Add support for running jobs in a different namespace (#53, #67, #73)
  * (thanks @thschue for the contribution)
* With each release the helm chart is packaged and added to the assets (#77)

## Fixed Issues
* Uniform registration for remote execution planes over https doesn't work (https://github.com/keptn/keptn/issues/4516)
* Display correct timeout value in job timeout error message (#49, #61)

## Known Limitations
