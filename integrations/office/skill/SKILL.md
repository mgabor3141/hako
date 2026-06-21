---
name: office
description: >
  Create, read, and edit Word (.docx), Excel (.xlsx), and PowerPoint (.pptx)
  files with Python. Use for any task that produces or inspects Office documents.
---

# office

Office files are just OOXML. You manipulate them with mature, pure-Python
libraries -- no Office suite, no conversion service. Write a short script and
run it.

## Setup (once per environment, then persists)

```sh
python -m pip install -r ~/.agents/skills/office/requirements.txt
```

Idempotent and quick; the packages live in the persistent home. Run it if an
`import` below fails.

## Which library

- **Word** -> `docx` (python-docx): `Document()`, `add_heading` / `add_paragraph`
  / `add_table`, then `.save("out.docx")`. Opens existing files to edit too.
- **Excel** -> `openpyxl`: `Workbook()` or `load_workbook(path)`, `ws["A1"] = ...`,
  formulas as plain strings (`"=SUM(A1:A9)"`), `.save(...)`.
- **PowerPoint** -> `pptx` (python-pptx): `Presentation()`, `slides.add_slide(layout)`,
  fill placeholders/shapes, `.save(...)`.

To read a document down to text/markdown quickly, `pip install
markitdown[docx,pptx,xlsx]` and run `markitdown file.docx`.

## Know the one tradeoff

There is **no renderer** here, so you cannot *see* the laid-out page or export a
faithful PDF. Therefore:

- Build structurally, then verify by **re-reading the file** (open it back and
  assert the text/sheet/slide is what you intended) rather than trusting it
  blind.
- If a task truly needs visual fidelity ("does it look right", PDF export), say
  so -- that needs a render engine hako does not bundle by default; don't guess.

`poppler-utils` (`pdftotext`, `pdftoppm`) and `imagemagick` are already on PATH
for PDF text/image extraction.
