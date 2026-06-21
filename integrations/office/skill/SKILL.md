---
name: office
description: >
  Create, read, and edit Word (.docx), Excel (.xlsx), and PowerPoint (.pptx)
  files with Python. Use for tasks that produce or inspect Office documents.
---

# office

Office files are OOXML; manipulate them with the preinstalled Python libraries.
Write a short script and run it -- they read, edit, and create.

- **Word** -> `docx` (python-docx)
- **Excel** -> `openpyxl`
- **PowerPoint** -> `pptx` (python-pptx)

**No renderer here:** you can't see the laid-out page or export a faithful PDF.
Build structurally and verify by re-reading the file; if a task genuinely needs
visual fidelity or PDF, say so rather than guess. `poppler-utils` and
`imagemagick` are on PATH for PDF/image extraction.
