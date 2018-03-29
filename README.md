# Packet Guardian

[![Build Status](https://travis-ci.org/packet-guardian/packet-guardian.svg?branch=develop)](https://travis-ci.org/packet-guardian/packet-guardian)
[![GoDoc](https://godoc.org/github.com/packet-guardian/packet-guardian/src?status.svg)](https://godoc.org/github.com/packet-guardian/packet-guardian/src)
[![GitHub issues](https://img.shields.io/github/issues/packet-guardian/packet-guardian.svg)](https://github.com/packet-guardian/packet-guardian/issues)
[![GitHub stars](https://img.shields.io/github/stars/packet-guardian/packet-guardian.svg)](https://github.com/packet-guardian/packet-guardian/stargazers)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/packet-guardian/packet-guardian/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/packet-guardian/packet-guardian)](https://goreportcard.com/report/github.com/packet-guardian/packet-guardian)

## About Packet Guardian

- [Documentation](docs)
- [Building](docs/building.md)
- [Contributing](CONTRIBUTING.md)


Packet Guardian is an easy to use captive portal for wired or wireless networks. It works in conjunction with a local DNS server and integrated DHCP server to redirect clients to a registration page where they can log in or register as a guest to gain access to a network. The configuration provides a lot of customization to fit any environment.

The source code for the DHCP server has been separated and moved to its own repository. You can find the project [here](https://github.com/packet-guardian/pg-dhcp).
