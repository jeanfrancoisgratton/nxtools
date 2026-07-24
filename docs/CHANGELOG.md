| Release | Date       | Comments                                                                                                                                                                             |
|---------|------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1.1.4   | 2026.07.24 | Raw format support could not properly sort various versions of a package as it lacks support for the `/service/rest/v1/search` endpoint                                              |
| 1.1.3   | 2026.07.23 | Fixed private/public signing keypair generation for APK packages                                                                                                                     |
| 1.1.2   | _n/a_      | First attempt at signing APK repo                                                                                                                                                    |
| 1.1.1   | 2026.07.23 | Fixed the default root path in Alpine repos                                                                                                                                          |
| 1.1.0   | 2026.07.20 | Added `assets latest` to get the latest version of `PACKAGE` in `REPO`                                                                                                               |
| 1.0.2   | 2026.07.20 | cicd nearly automated                                                                                                                                                                |
| 1.0.1   | 2026.07.20 | fixed apkbuild upload specs                                                                                                                                                          |
| 1.0.0   | 2026.07.18 | First stable (1.0.0) release<br>Added Alpine/APK repo format<br>REST client honours NEXUS_* environment variables<br>Added unit tests across packages; fixed yum deploy-policy check | 
| 0.91.00 | 2026.05.30 | Added missing wiring of completionCmd to rootCmd                                                                                                                                     | 
| 0.85.00 | 2026.03.25 | Added support to the docker repo format<br>Added a repo migration feature                                                                                                            | 
| 0.80.00 | 2026.03.25 | `repo reindex` can now be called from `repo upload`                                                                                                                                  |
| 0.75.00 | 2026.03.25 | Moved the `repo reindex` command to another endpoint<br>Yum and maven2 support added                                                                                                 |
| 0.72.00 | 2026.03.24 | Added support for APT and *generic* repo formats                                                                                                                                     |
| 0.50.00 | 2026.03.14 | Completed blobs and assets                                                                                                                                                           |
| 0.10.00 | 2026.01.28 | Initial version.                                                                                                                                                                     |

