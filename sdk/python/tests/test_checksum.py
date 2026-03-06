"""Tests for bkn.checksum API."""

from __future__ import annotations

import tempfile
from pathlib import Path

import pytest


def test_generate_and_verify_checksum():
    from bkn import generate_checksum_file, verify_checksum_file

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "objects").mkdir()
        (root / "objects" / "pod.bkn").write_text(
            "---\ntype: object\nid: pod\nname: Pod\nnetwork: k8s\n---\n\n## Object: pod\n**Pod**\n",
            encoding="utf-8",
        )
        content = generate_checksum_file(root)
        assert "sha256:" in content
        assert "objects/pod.bkn" in content
        assert "*" in content
        ok, errs = verify_checksum_file(root)
        assert ok, errs
        assert len(errs) == 0


def test_verify_fails_when_file_changed():
    from bkn import generate_checksum_file, verify_checksum_file

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "test.bkn").write_text(
            "---\ntype: object\nid: x\n---\n\n## Object: x\n",
            encoding="utf-8",
        )
        generate_checksum_file(root)
        (root / "test.bkn").write_text(
            "---\ntype: object\nid: x\n---\n\n## Object: x\nModified\n",
            encoding="utf-8",
        )
        ok, errs = verify_checksum_file(root)
        assert not ok
        assert any("Mismatch" in e for e in errs)
