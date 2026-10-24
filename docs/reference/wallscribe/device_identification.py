#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
機器識別モジュール
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Optional

from models.config import DeviceType

from .textfsm_utils import run_textfsm


@dataclass
class IdentifiedDevice:
    """識別された機器情報"""

    device_type: DeviceType = DeviceType.UNKNOWN
    model: str = ""
    os_version: str = ""


def format_fortigate_display_model(model_code: str) -> str:
    """config-version 由来の機種コードを表示用モデル名に変換する。"""
    if not model_code:
        return "FortiGate"

    compact_match = re.fullmatch(r"(\d{2})([A-Z])(\d)", model_code)
    if compact_match:
        two_digits, series_letter, last_digit = compact_match.groups()
        model_suffix = f"{two_digits}0{last_digit}{series_letter}"
        return f"FortiGate {model_suffix}"

    return f"FortiGate {model_code}"


def identify_device(content: str) -> Optional[IdentifiedDevice]:
    """TextFSMを利用して設定ファイルの機器種別を識別する。"""
    sample = content[:65536]

    fortigate_rows = run_textfsm("identify/fortigate_header.textfsm", sample)
    if fortigate_rows:
        row = fortigate_rows[0]
        return IdentifiedDevice(
            device_type=DeviceType.FORTIGATE,
            model=format_fortigate_display_model(row.get("model_code", "")),
            os_version=row.get("os_version", ""),
        )

    paloalto_set_rows = run_textfsm("identify/paloalto_set_cli.textfsm", sample)
    if paloalto_set_rows:
        return IdentifiedDevice(
            device_type=DeviceType.PALOALTO,
            model="PA Series",
            os_version="",
        )

    paloalto_rows = run_textfsm("identify/paloalto_config_root.textfsm", sample)
    if paloalto_rows:
        row = paloalto_rows[0]
        return IdentifiedDevice(
            device_type=DeviceType.PALOALTO,
            model="PA Series",
            os_version=row.get("version", ""),
        )

    return None
