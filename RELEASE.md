## v1.1 - Bug Fixes
+ Execute all watchers on start when run as a service.
* Clean all key path elements replacing `-` with `_`.
* Fix bug which causes hang on stop forcing the use of a SIGKILL.

## v1.0 - Initial Release
+ Trigger template rendering and command execution.
+ Watch and retrieve context for multiple keys.
+ Only execute command when rendered templates change.
+ Support multple templates.
