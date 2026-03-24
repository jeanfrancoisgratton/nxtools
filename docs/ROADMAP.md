# ROADMAP, INCLUDED FEATURES

## Current version
```bash
# nxtools -v
nxtools version 0.75.00 (2026.03.24), Go version = v1.26.1
```
## Supported repo formats
=> a format is considered *supported* once we can create a repo in this format
- [ ] docker
- [ ] yum
- [x] apt
- [x] raw
- [x] helm
- [x] cargo
- [ ] maven2
- [x] npm
- [x] nuget
- [x] pypi
- [x] terraform
- [x] swift
___
## Supported blob stores
- [x] file-based
- [ ] AWS S3
- [ ] GCP
___
## Roadmap

| Task                                                | Slated for  | Actual release | Comments |
|-----------------------------------------------------|-------------|----------------|----------|
| [x] repo ls                                         | 0.10.00     |                |          | 
| [x] blob ls, create, remove                         | 0.20.00     |                |          |
| [x] upload asset                                    | 0.20.00     |                |          |
| [x] repo reindex                                    | 0.30.00     |                |          |
| [x] assets subcommands                              | 0.40.00     | 0.50.00        |          |
| [x] extended information on `repo ls`               | 0.60.00     |                |          |
| [x] json output on most functions                   | 0.60.00     |                |          |
| [ ] pypi, nuget, npm and maven2 repo format support | ~~0.60.00~~ |                | delayed  | 
