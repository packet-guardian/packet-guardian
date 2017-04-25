# Contributing to Packet Guardian

Want to help contribute to Packet Guardian? Great! Any and all help is certainly appreciated, whether it's code, documentation, or spelling corrections.

If you are planning to contribute a significant change, please draft a design document (or start a conversation) and post it to the [issue tracker](https://github.com/packet-guardian/packet-guardian/issues). This will allow other developers and users to discuss the proposed change.

## Filing issues

When filing an issue, make sure to answer these six questions:

1. What version of Packet Guardian are you using (`pg -version`)?
2. What version of Go are you using? (`go version`)?
3. What operating system and processor architecture are you using?
4. What did you do?
5. What did you expect to see?
6. What did you see instead?

## Contributing code

This repository follows the guidelines of [git flow](http://nvie.com/posts/a-successful-git-branching-model/). Fork this repo and create your own feature branch off of develop. Any new dependencies should be vendored using govendor. When submitting a PR, base it against the develop branch, not master.

### Copyright

By submitting code to this project, you agree to release your contribution under the BSD 3-clause or a less restrictive license. This will provide the best flexibility and compatibility with the project. All new files should have the following copyright header at the top:

```
// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
```

### Guidelines

Consider the following guidelines when preparing to submit a patch:

* Follow standard Go conventions (document any new exported types, funcs, etc.., ensuring proper punctuation).
* Ensure that you test your code. Any patches sent in for new / fixed functionality must include tests in order to be merged.
* If you plan on making any major changes, create an issue before sending a patch. This will allow for proper discussion beforehand.
* Keep any os / arch specific code contained to os / arch specific files. Packet Guardian may leverage Go's filename based conditional compilation, i.e do not put Linux specific functionality in a non Linux specific file.
* Ensure all Go code has been run through `go fmt`.
* Ensure all other code follows surrounding coding conventions.
* Ensure commits are small and atomic. Meaning, each commit should do one thing and one thing only. There's no need to squash commits before making a PR.

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why. If an issue exists that the commit
fixes, make sure to include it in the commit message.

```
Added feature X

This change adds feature X to package Y.

Fixes #38
```

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various
git tools.
