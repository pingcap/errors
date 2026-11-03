## Usage

```shell script
# eg: ./errdoc-gen --source devel/pingap/tidb --module github.com/pingcap/tidb --output devel/pingap/tidb/errors.toml
./errdoc-gen --source /path/to/source/code --module ${module-name} --output /path/to/errors.toml
```

Only run errdoc-gen on trusted source trees. It compiles and runs a generated Go program that imports the scanned packages, executing their variable initializers and `init()` functions, including those of their dependencies. Input validation does not sandbox this execution.

`--module` and package import paths must match `[A-Za-z0-9_]+([.-][A-Za-z0-9_]+)*(/[A-Za-z0-9_]+([.-][A-Za-z0-9_]+)*)*`. This accepts domains, hyphenated names, and version numbers while rejecting empty components and traversal. `--ignore` accepts comma-separated paths using the same pattern, with an optional `./` prefix and trailing `/`, relative to `--source`; surrounding whitespace and blank entries are ignored. Parse failures, invalid paths, and scans with no error declarations or no normalized errors fail without overwriting the output.
