# GoLang Module Release Notes

## Unreleased

* Updated list of inspectable XML content types
* Added `http.Flusher` interface when the underlying handler supports this interface
* Updated timeout to include time to connect to the agent
* Cleaned up docs/code/examples

## 1.5.0 2019-01-31

* Switched Update / Post RPC call to async
* Internal release for agent reverse proxy

## 1.4.3 2018-08-07

* Improved error and debug messages
* Exposed more functionality to allow easier extending


## 1.4.2 2018-06-15
* Improved handling of the `Host` request header
* Improved debugging output

## 1.4.1 2018-06-04
* Improved error and debug messages

## 1.4.0 2018-05-24

* Standardized release notes
* Added support for multipart/form-data post
* Extended architecture to allow more flexibility
* Updated response writer interface to allow for WebSocket use
* Removed default filters on CONNECT/OPTIONS methods - now inspected by default
* Standardized error page
* Updated to contact agent on init for faster module registration

## 1.3.1 2017-09-25

* Removed unused dependency
* Removed internal testing example

## 1.3.0 2017-09-19

* Improved internal testing
* Updated msgpack serialization

## 1.2.3 2017-09-11

* Standardized defaults across modules and document
* Bad release

## 1.2.2 2017-07-02

* Updated to use [signalsciences/tlstext](https://github.com/signalsciences/tlstext)

## 1.2.1 2017-03-21

* Added ability to send XML post bodies to agent
* Improved content-type processing

## 1.2.0 2017-03-06

* Improved performance
* Exposed internal datastructures and methods
  to allow alternative module implementations and
  performance tests

## 1.1.0 2017-02-28

* Fixed TCP vs. UDS configuration

## 0.1.0 2016-09-02

* Initial release
