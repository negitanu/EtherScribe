#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""台帳に値を出さないフィールドと、ENC 文字列の伏せ字。"""

import re
from typing import Any

REDACTED = "<redacted>"
SECRET_FIELDS = frozenset({"password", "community", "secret", "private_key", "psk_value"})
_ENC = re.compile(r"ENC\s+\S+")


def is_secret_field(name: str) -> bool:
    """パスワード・SNMP コミュニティなど、値そのものが資格情報になり得るフィールドか。"""
    return name in SECRET_FIELDS or name.endswith(("_password", "_secret"))


def scrub_secrets(value: Any) -> str:
    """表示用文字列から ENC で始まる秘密値を除く。"""
    text = "" if value is None else str(value)
    return _ENC.sub(REDACTED, text)


def secret_presence(value: str) -> str:
    """秘密そのものは出さず、設定されていることだけを示す。"""
    return "設定あり" if value else ""
