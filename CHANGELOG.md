# GoLang Module Release Notes

## 1.8.2 2020-10-19

* Added `server_flavor` config option.

## 1.8.1 2020-06-15

* Added internal release metadata support.

## 1.8.0 2020-06-15

* Deprecated the `AltResponseCodes` concept in favor of using all codes 300-599 as "blocking"
* Added HTTP redirect support

## 1.7.1 2020-04-06

* Updated the response recorder to implement the io.ReaderFrom interface
* Fixed some linter issues with missing comments on exported functions

## 1.7.0 2020-03-11

* Cleaned up configuration and added an `AltResponseCodes` option to configure
  alternative (other than 406) response codes that can be used for blocking

## 1.6.5 2020-01-06

* Updated the `http.ResponseWriter` wrapper to allow `CloseNotify()` calls to pass through

## 1.6.4 2019-11-06

* Updated helloworld example to be more configurable allowing it to be used in other example documentation
* Added the ability to support inspecting gRPC (protobuf) content

## 1.6.3 2019-09-12

* Added custom header extractor to the post request

## 1.6.2 2019-08-25

* Added support for a custom header extractor fn

## 1.6.1 2019-06-13

* Cleaned up internal code

## 1.6.0 2019-05-30

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
