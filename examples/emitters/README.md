# Automi emitter examples

Automi emitters are components that are placed at the start of a stream. They
are used to emit values unto a stream.  The following shows examples for supported Automi emitters.

* [channel](./channel) - Example to show how to emit items from a Go channel unto the stream.
* [csv](./csv) - Example that shows how to use a CSV file as a source for a stream.
* [reader](./reader) - Example that shows how to use an `io.Reader` to emit items on a stream.
* [scanner](./scanner) - Example that shows how to emit scanned items from an `io.Reader` unto the stream.
* [slice0](./slice0) - Shows how to emit items from a slice unto a stream.
* [slice1](./slice1) - Shows how to emit items of custom types from a slice unto the stream.