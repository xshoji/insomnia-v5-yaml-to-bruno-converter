# insomnia-v5-yaml-to-bruno-converter

insomnia-v5-yaml-to-bruno-converter

## Features

Compared to Bruno's built-in Insomnia YAML import functionality, this tool provides enhanced capabilities:
- **Request order preservation**: Maintains the original order of requests and folders from your Insomnia collection
- **Environment migration support**: Fully converts and migrates Insomnia environments to Bruno environment files



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

Pushing a version tag runs the `goreleaser` workflow to build and publish the
GitHub release.

```bash
# Release
v=v0.0.2; git tag "${v}" && git push origin "refs/tags/${v}"

# Move the tag to HEAD and re-run the release
v=v0.0.2; git tag -f "${v}" && git push -f origin "refs/tags/${v}"
```

Deleting a tag turns its release into a draft. Move the tag instead to keep the
release and its existing notes while replacing its assets. Checksums change;
fix an already published version with a new version instead.
