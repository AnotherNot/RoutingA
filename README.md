# RoutingA

It is a parser lib that can help your routing rules more readable and understandable for users. 

## example

[parser_test.go](parser_test.go)

[v2rayA wiki](https://github.com/v2rayA/v2rayA/wiki/RoutingA)

## detail

[design and implementation details](https://github.com/v2rayA/v2rayA/wiki/%E5%BC%80%E5%8F%91%E7%9B%B8%E5%85%B3#RoutingA)

## process(...)

A condition that names the program behind a connection rather than an address:

```
process(firefox) -> proxy
process(name: "chrom*", folder: /opt/apps) -> direct
process(path: /usr/bin/curl) && port(443) -> proxy
```

`process` takes the executable's file name as a bare value, or as `name:`,
`path:` (a full path to the file) and `folder:` (a directory the executable
lives in or below). One of the three lists matching is enough. Values may hold
the shell-style wildcards `path.Match` understands: `*` stays inside one path
element, `?` is one character.

The parser already accepted `process(...)` as any other function; this fork
makes it part of the language, so a consumer reads a typed value instead of
matching the name and guessing the parameter shape:

```go
m, isProcess, err := condition.ProcessMatcher()
if err != nil {
    // the rule names something this version cannot enforce
}
```

`err` is set for a parameter other than `name`, `path` and `folder`, and for
`process()` with no value. Both are refusals rather than warnings: a condition
that cannot be enforced must not be dropped, because a rule without it matches
more than its author wrote.

## Fork

This is [AnotherNot/RoutingA](https://github.com/AnotherNot/RoutingA), a fork of
[v2rayA/RoutingA](https://github.com/v2rayA/RoutingA) used by a fork of v2rayA.
The module path stays `github.com/v2rayA/RoutingA`: a module path is an address,
not a label, and the fork is consumed through a `replace` directive. Everything
outside `process(...)` is upstream's, under the same licence.

## LICENSE

[MIT LICENSE](LICENSE)
