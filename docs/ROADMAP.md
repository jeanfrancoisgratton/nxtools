# ROADMAP, INCLUDED FEATURES

## Current version
```bash
# nxtools -v
nxtools version 0.30.00 (2026.03.10), Go version = go1.26.1
```
## Supported repo formats
- [x] docker
- [x] yum
- [x] apt
- [x] raw
- [x] helm
- [x] cargo
- [ ] maven
- [ ] npm
- [ ] nuget
- [ ] pypi
- [ ] terraform
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
