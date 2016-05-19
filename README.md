# fullcover

Sometimes `go tool cover` does not cut it. For example, if you need to grab coverage from Google App Engine,
you cannot simply create output files.

`fullcover` has you covered: It communicates coverage information via HTTP, thereby enabling you to
throw it over most fences.

## Usage

Installation:
```
go get github.com/Drahflow/fullcover
```

Rewriting code to report coverage:
```
fullcover -mode=remote -connection=localhost:10001 -o generated.go your-source.go
```

Getting your coverage report:
```
fullcover -connection=:10001 -daemon

go build generated.go && ./generated

firefox http://localhost:10001/
```

Stopping the collection daemon:
```
wget -O - http://localhost:10001/quit
```

## Planned features

* Reports with real time animation of covered code
* Allow monkey-patching function returns to get those darn `if err != nil` blocks covered
