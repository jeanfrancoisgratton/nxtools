# Nexus REST endpoints used by nxtools

The examples below use `https://nexus:8081` as the base URL; substitute your own
host and port (these are read from the environment file, or from the `NEXUS_HOST`
environment variable).

- Base URL (webUI): `https://nexus:8081`
- API base URL: `https://nexus:8081/service/rest/v1/`

All paths below are relative to the API base URL.

## Repositories
- `repositories` — list every repository
- `repositories/{format}/{type}` — create a repository (e.g. `repositories/yum/hosted`)
- `repositories/{repositoryName}` — delete a repository

## Assets & components
- `assets` — list/query assets in a repository
- `search` — search components (used by `assets pkginfo` and repo migration)
- `components` — upload a component/asset
- `formats/upload-specs` — per-format upload field specifications

## Blob stores
- `blobstores` — list blob stores
- `blobstores/file` — create a file-based blob store
- `blobstores/{name}/quota-status` — blob store quota status

## Tasks
- `tasks` — list scheduled tasks (used by `repo reindex`, which triggers the
  pre-existing `_reindex_{REPONAME}` task)
