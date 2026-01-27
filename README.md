# Zeno

A Go-based benchmark analysis tool

## Installation

```bash
go install github.com/mateusfdl/zeno@latest
```

build from source

```bash
git clone https://github.com/mateusfdl/zeno.git
cd zeno
go build
```

## Usage

### Parse bench ouput

Parse Go bench output from stdin and stores it as json

```bash
go test -bench=. -benchmem | zeno parse -o results.json
```

Add metadata to the run

```bash
go test -bench=. | zeno parse --version=v1.0.0 --tags=ci -o results.json
```

Append to an existing file

```bash
go test -bench=. | zeno parse --append -o history.json
```

### Merge bench files

Merge multiple json files

```bash
zeno merge -o combined.json file1.json file2.json file3.json
```

Sort by date and deduplicate:

```bash
zeno merge --unique --sort-desc -o all.json *.json
```

### Compare runs

Compare two benchmark runs

```bash
zeno compare baseline.json current.json
```

Use a custom regression threshold

```bash
zeno compare --threshold=2.5 before.json after.json
```

output as json

```bash
zeno compare --format=json old.json new.json
```

### Views (TUI)

```bash
zeno view -f results.json

zeno view -f current.json --compare baseline.json

go test -bench=. -benchmem | zeno view

go test -bench=. -benchmem | zeno view --web
```

## Stored format

Benchmark data is stored as the following JSON:

```json
[
  {
    "version": "v1.0.0",
    "date": 1609459200,
    "tags": ["foo"],
    "suites": [
      {
        "go": "go1.23",
        "goos": "linux",
        "goarch": "amd64",
        "pkg": "github.com/example/mypackage",
        "benchmarks": [
          {
            "name": "BlazinglySlowFn",
            "runs": 1000000,
            "nsPerOp": 120.5,
            "mem": {
              "bytesPerOp": 16.0,
              "allocsPerOp": 1.0,
              "mbPerSec": 12.5
            },
            "custom": {}
          }
        ]
      }
    ]
  }
]
```
### Tracking performance over time/runs

```bash
go test -bench=. -benchmem | zeno parse --version=$(git rev-parse HEAD) -o bench-new.json

zeno compare --threshold=1.0 \ $(git show HEAD~1:bench-baseline.json) bench-new.json
```

## License

MIT License - feel free to use in your projects!

## Contributing

Contributions are welcome! Please feel free to open an Issue or submit a Pull Request.

Happy hacking :) 

## Acknowledgments
- Inspired by [gobenchdata](https://github.com/bobheadxi/gobenchdata)
- Inspired by [vizb](https://github.com/arielril/vizb)
