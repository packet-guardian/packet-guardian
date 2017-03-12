# Packet Guardian

[![Build Status](https://travis-ci.org/usi-lfkeitel/packet-guardian.svg?branch=master)](https://travis-ci.org/usi-lfkeitel/packet-guardian)
[![GoDoc](https://godoc.org/github.com/usi-lfkeitel/packet-guardian/src?status.svg)](https://godoc.org/github.com/usi-lfkeitel/packet-guardian/src)
[![GitHub issues](https://img.shields.io/github/issues/usi-lfkeitel/packet-guardian.svg)](https://github.com/usi-lfkeitel/packet-guardian/issues)
[![GitHub stars](https://img.shields.io/github/stars/usi-lfkeitel/packet-guardian.svg)](https://github.com/usi-lfkeitel/packet-guardian/stargazers)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/usi-lfkeitel/packet-guardian/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/usi-lfkeitel/packet-guardian)](https://goreportcard.com/report/github.com/usi-lfkeitel/packet-guardian)

## About Packet Guardian

- [Documentation](docs)
- [Building](docs/building.md)
- [Contributing](CONTRIBUTING.md)


Packet Guardian is an easy to use captive portal for wired or wireless networks. It works in conjunction with a local DNS server and integrated DHCP server to redirect clients to a registration page where they can log in or register as a guest to gain access to a network. The configuration provides a lot of customization to fit any environment.

The source code for the DHCP server has been separated and moved to its own repository. You can find the project [here](https://github.com/usi-lfkeitel/pg-dhcp).
