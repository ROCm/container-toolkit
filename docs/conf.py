"""Configuration file for the Sphinx documentation builder."""
import os
from pathlib import Path

html_baseurl = os.environ.get("READTHEDOCS_CANONICAL_URL", "instinct.docs.amd.com")
html_context = {}
if os.environ.get("READTHEDOCS", "") == "True":
    html_context["READTHEDOCS"] = True

project = "AMD Container Runtime Toolkit"

version = "1.3.0"
rocm_version = '6.4'
rocm_directory_version = '6.4.0' # in 6.0 rocm was located in /opt/rocm-6.0.0
amdgpu_version = '6.4' # directory in https://repo.radeon.com/rocm/apt/ and https://repo.radeon.com/amdgpu-install/
amdgpu_install_version = '6.4.60400-1'
release = version
html_title = project
author = "Advanced Micro Devices, Inc."
copyright = "Copyright (c) 2024 Advanced Micro Devices, Inc. All rights reserved."

# Supported linux version numbers
ubuntu_version_numbers = [('24.04', 'noble'), ('22.04', 'jammy')]
debian_version_numbers = [('12', 'jammy')]
rhel_release_version_numbers = ['9', '8']
rhel_version_numbers = ['9.5', '9.4', '8.10']
sles_version_numbers = ['15.6']
ol_release_version_numbers = ['9', '8']
ol_version_numbers = [('9.5'), ('8.10')]
azl_version_numbers = ['3.0']

html_context.update({
    "ubuntu_version_numbers": ubuntu_version_numbers,
    "debian_version_numbers": debian_version_numbers,
    "sles_version_numbers": sles_version_numbers,
    "rhel_release_version_numbers": rhel_release_version_numbers,
    "rhel_version_numbers": rhel_version_numbers,
    "ol_release_version_numbers": ol_release_version_numbers,
    "ol_version_numbers": ol_version_numbers,
    "azl_version_numbers": azl_version_numbers,
})

# Required settings
html_theme = "rocm_docs_theme"
html_theme_options = {
    "flavor": "instinct",
    "link_main_doc": True,
    "use_download_button": True,
    # Add any additional theme options here
}
extensions = [
    "rocm_docs",
    "sphinxcontrib.datatemplates",
    "sphinx_substitution_extensions",
]

# Table of contents
external_toc_path = "./sphinx/_toc.yml"

exclude_patterns = ['.venv']

# Add the following replacements to every RST file.
rst_prolog = f"""
.. |rocm_version| replace:: {rocm_version}
.. |amdgpu_version| replace:: {amdgpu_version}
.. |rocm_directory_version| replace:: {rocm_directory_version}
.. |amdgpu_install_version| replace:: {amdgpu_install_version}
"""

import re

EXCLUDED_DIRS = {
    "_build",
    "_templates",
    "_static",
    ".git",
    ".venv",
}

MARKUP_PREFIXES = (
    ":::",
    "```{",
    "```",
    ":img-top:",
    ":class",
    ":link:",
    ":link-type:",
    ":shadow:",
    ":columns:",
    ":padding:",
    ":gutter:",
    ":open:",
    ":name:",
    ":header-rows:",
    ":alt:",
    "+++",
    "<",
    "-->",
    "{bdg-",
)

# Matches lines like "align: center", "alt:", "name: foo" (directive options
# not starting with a colon, common in MyST figure/table fences)
_BARE_DIRECTIVE_RE = re.compile(r"^[a-z][a-z_-]*:\s*\S*$")

# Matches MyST/RST anchor labels like "(some-label)="
_ANCHOR_LABEL_RE = re.compile(r"^\(\w[\w-]*\)=$")

MIN_PROSE_LINES = 10


def should_skip(path: Path) -> bool:
    return any(part in EXCLUDED_DIRS for part in path.parts)


def is_prose_line(line: str) -> bool:
    stripped = line.strip()
    if not stripped:
        return False
    if stripped.startswith(MARKUP_PREFIXES):
        return False
    # Drop bare directive-option lines (e.g. "align: center", "alt:")
    if _BARE_DIRECTIVE_RE.match(stripped):
        return False
    # Drop MyST/RST anchor labels (e.g. "(some-label)=")
    if _ANCHOR_LABEL_RE.match(stripped):
        return False
    # Drop lines that contain an HTML tag anywhere (e.g. ".</p>")
    if re.search(r"</?[a-zA-Z]", stripped):
        return False
    return True


def generate_combined_markdown(app, exception):
    if exception:
        return

    docs_root = Path(app.srcdir)
    output_file = Path(app.outdir) / "llms.txt"
    base_file = docs_root / "llms.txt"

    combined = []

    if base_file.exists():
        base_text = base_file.read_text(encoding="utf-8").rstrip().rstrip("-").rstrip()
        combined.append(base_text)
    else:
        combined.append("# AMD Container Toolkit")

    all_files = sorted(docs_root.rglob("*.md"))

    for doc_file in all_files:
        if should_skip(doc_file):
            continue

        if doc_file == base_file:
            continue

        try:
            content = doc_file.read_text(encoding="utf-8")
        except Exception:
            continue

        lines = content.splitlines()
        prose_lines = [line for line in lines if is_prose_line(line)]

        if len(prose_lines) < MIN_PROSE_LINES:
            continue

        relative = doc_file.relative_to(docs_root)
        cleaned = "\n".join(
            line for line in lines
            if line.strip() == "" or is_prose_line(line)
        )

        combined.append(f"\n\n---\n\n# {relative}\n")
        combined.append(cleaned.strip())

    output_file.write_text(
        "\n".join(combined) + "\n",
        encoding="utf-8",
    )

def setup(app):
    app.connect("build-finished", generate_combined_markdown)
