# Video Editor

Video file editing with filesystem abstraction.

- FileSystem interface for testability (ReadFile, WriteFile)
- Editor interface: ReplaceFrame returns modified bytes
- NewEditor factory injects dependencies
- OSFileSystem wraps os helpers
- File permissions: 0o644 (owner rw, group/others r)
