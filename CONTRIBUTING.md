# Contributing

Contributions to VPP-Agent are welcome. We use the standard pull request
model. You can either pick an open issue and assign it to yourself or open
a new issue and discuss your feature.

In any case, before submitting your pull request please check the 
[Coding style](CODINGSTYLE.md) and cover the newly added code with tests 
and documentation.

The tool used for managing third-party dependencies is 
[Glide](https://github.com/Masterminds/glide). After adding or updating a
dependency in `glide.yaml` run `make install-dep` to download the specified
dependencies into the vendor folder. Please make sure that each dependency 
in the `glide.yaml` has a specific `version` defined (a specific commit ID
or a git tag).
