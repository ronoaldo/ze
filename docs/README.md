# Development Process

This directory contains documentation regarding the development process of the `ze` agent. The primary goal of the project was to create an AI agent powered 100% by local models.

## Project Evolution

Development followed an iterative "bootstrap" approach, starting with a "version zero" document to build a minimum viable version of `ze` capable of "self-improvement."

From that point on, the agent was used to implement improvements into its own codebase - such as creating new tools, adding logging, and supporting Markdown display. This cycle of development - where the agent is used to build and refine itself - simultaneously served to define and format the tool's usage patterns.

## Local LLM Optimization

The design of `ze` was specifically optimized for the use of local LLMs, prioritizing:
- **Tool Simplicity:** Reducing the number of tools to prevent model confusion.
- **Prompt Refinement:** Precise adjustment of descriptions to ensure correct capability usage.
- **Context Management:** Preventing common errors, such as the loss of reasoning history during interaction with the model.

## Workflow and Documentation

The documentation within this directory tracks each phase of this evolution. The established workflow allowed the agent to operate almost autonomously for new features:

1. **Direction:** The user points to a specification (e.g., "read feature X in docs/03-features.md").
2. **Planning:** The agent traces a detailed development plan.
3. **Implementation:** The agent executes the plan step-by-step.

This method allows for the systematic addition of new functions while maintaining project consistency.

## Timeline

- [01 - Refactoring of Tools](01-refatoracao-de-ferramentas.md): Refactoring of tools to ensure modularity, individual tests per tool, and improvements in TUI log displays.
- [02 - UI and UX Improvements](02.0-melhorias-de-ui.md): Significant UI and UX improvements, including multiline commands, color deactivation, enhanced diff display, and stdin support.
    - [02.1 - InputHandler Refactoring](02.1-refatoracao.md): Plan to extract input logic into a testable component, enabling TDD for the multiline feature.
    - [02.2 - Git Diff Refactoring](02.2-diff.md): Plan to fix `git diff` statistics display, handling multiple files, binary files, and untracked file counts.
    - [02.3 - Headless Mode](02.3-headless.md): Plan to enable `ze` use in non-interactive environments (pipes, scripts) through automatic TTY detection.
- [03 - New Features](03-features.md): Implementation of new features including shell command execution, WebFetch tool, Markdown rendering, full Go documentation, session persistence, and execution logs.
