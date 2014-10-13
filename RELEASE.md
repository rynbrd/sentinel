## v1.0.5 - Logger Bug Fix
* Fix logger output setup on start.

## v1.0.4 - Watcher Bug Fix
* Validate watcher configuration.

## v1.0.3 - Watcher Bugs
* Do not exit when a watcher fails.

## v1.0.2 - Template Bug Fixes
* Renderer templates in temporary files under the destination directory. Fixes
  cross-volume link error.
* Fail template rendering if destination hash generation fails.

## v1.0.1 - Bug Fixes
+ Execute all watchers on start when run as a service.
* Clean all key path elements replacing `-` with `_`.
* Fix bug which causes hang on stop forcing the use of a SIGKILL.

## v1.0.0 - Initial Release
+ Trigger template rendering and command execution.
+ Watch and retrieve context for multiple keys.
+ Only execute command when rendered templates change.
+ Support multple templates.
