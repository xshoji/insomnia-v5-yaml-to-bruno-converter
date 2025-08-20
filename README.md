# insomnia-v5-yaml-to-bruno-converter

insomnia-v5-yaml-to-bruno-converter

# Usage

```
./itb -h
Usage: ./itb [OPTIONS] [-h, --help]

Description:
  This tool converts Insomnia-exported files (v5 YAML) into a Bruno collection files.

Options:
  -f string    (REQ) Path to Insomnia exported file
  -n string    (REQ) Name of bruno collection
  -o string    (REQ) Output directory
```

# Example

```
./itb -f /tmp/insomnia-export.1755624285825/Scratch-Pad-wrk_scratchpad.yaml -o /tmp/bruno_collection/ -n "My bruno collection"
```

# Build

```bash
APP="/tmp/itb"; MAIN="main.go"; GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "${APP}" "${MAIN}"; chmod +x "${APP}"
```


# Release

Release flow of this repository is integrated with github action. Git tag pushing triggers release job.

```
# Release
git tag v0.0.2 && git push --tags
```

```
# Delete tag
echo "v0.0.1" |xargs -I{} bash -c "git tag -d {} && git push origin :{}"
```

```
# Delete tag and recreate new tag and push
echo "v0.0.1" |xargs -I{} bash -c "git tag -d {} && git push origin :{}; git tag {} -m \"Release beta version.\"; git push --tags"
```
