# PG-DHCP

[![GoDoc](https://godoc.org/github.com/usi-lfkeitel/pg-dhcp?status.svg)](https://godoc.org/github.com/usi-lfkeitel/pg-dhcp)
[![GitHub issues](https://img.shields.io/github/issues/usi-lfkeitel/pg-dhcp.svg)](https://github.com/usi-lfkeitel/pg-dhcp/issues)
[![GitHub stars](https://img.shields.io/github/stars/usi-lfkeitel/pg-dhcp.svg)](https://github.com/usi-lfkeitel/pg-dhcp/stargazers)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/usi-lfkeitel/pg-dhcp/master/LICENSE)

This is the DHCP server package backing the Packet Guardian captive portal. It has been separated into it's own repository to make development a bit easier, and to provide a better focus to the origin project. This package may be used completely independent of Packet Guardian.

Features:

- RFC2131 DHCP protocol
- The most used options are implement, more to come
- Seperation of registered vs unregistered devices (known/unknown)
- Storage independent (the calling project is responsible for storage)

[Configuration File Format](configurationFileFormat.md)
