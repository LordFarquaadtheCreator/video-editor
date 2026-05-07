# CLI Commands

Cobra-based CLI layer.

- Commands use `NewRootCmd(editor)` factory pattern
- Dependencies injected via interfaces (video.Editor, frame.Replacer)
- Error handling: wrap with context using `fmt.Errorf`
- Output to stdout/stderr via cmd.OutOrStdout()
