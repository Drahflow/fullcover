# fullcover

Sometimes `go tool cover` does not cut it. For example, if you need to grab coverage from Google App Engine,
you cannot simply create output files.

`fullcover` has you covered: It communicates coverage information via HTTP, thereby enabling you to
throw it over most fences.

## Usage

Installation:
```
go install github.com/Drahflow/fullcover
```

Rewriting code to report coverage:
```
go tool fullcover -mode=remote -conection=localhost:10001 -o generated.go source.go
```

Getting your coverage report:
```
go tool fullcover -connection=:10001 -daemon

go build generated.go && ./generated

firefox http://localhost:10001/
```

## Planned features

* Reports with real time animation of covered code
* Allow monkey-patching function returns to get those darn `if err != nil` blocks covered
