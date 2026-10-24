"""Selection semantics shared by parameter-sheet exporters, the Web UI and the CLI.

An empty selection retains the historical meaning of all sections. Network is
one HTML section and three Excel sheets; legacy names map accordingly.
"""

from typing import Dict, List, NamedTuple, Optional

COMMON_ALIASES = {
    "device": frozenset(("device_info", "system_settings", "ha", "logging")),
    "policy": frozenset(("policies", "nat")),
}
FORMAT_ALIASES = {
    "html": {key: frozenset(("network",)) for key in ("interfaces", "routes", "dhcp")},
    "excel": {"network": frozenset(("interfaces", "routes", "dhcp"))},
    "json": {},
}


def section_selected(sections, key, output_format):
    if not sections:
        return True
    if key in sections:
        return True
    aliases = FORMAT_ALIASES[output_format]
    return any(
        key in COMMON_ALIASES.get(name, ()) or key in aliases.get(name, ()) for name in sections
    )


class SectionInfo(NamedTuple):
    key: str
    group: str
    label: str
    description: str


class Preset(NamedTuple):
    key: str
    label: str
    description: str
    sections: List[str]


SECTION_GROUPS: Dict[str, str] = {
    "global": "グローバル",
    "analysis": "診断・構造図",
    "investigation": "通信調査",
    "vdom": "VDOM / vsys",
}

SECTION_CATALOG: List[SectionInfo] = [
    SectionInfo("device_info", "global", "機器概要", "ホスト名・機種・OS・ライセンス・件数サマリー"),
    SectionInfo("system_settings", "global", "システム設定", "管理アクセス・DNS・NTP・管理者アカウント"),
    SectionInfo("ha", "global", "HA", "モード・優先度・ハートビート・監視 IF"),
    SectionInfo("logging", "global", "ログ・監視", "Syslog・SNMP・FortiManager / Panorama"),
    SectionInfo(
        "security_analysis", "analysis", "セキュリティ診断", "重要度・根拠・改善案を表示（設定の静的診断）"
    ),
    SectionInfo(
        "topology",
        "analysis",
        "推定ネットワーク構造",
        "全 VDOM の統合図・リンク追跡（Excel は構造・接続一覧）",
    ),
    SectionInfo(
        "flow_analysis",
        "investigation",
        "通信候補の確認",
        "HTML で IP・TCP/UDP ポートを入力して根拠を確認",
    ),
    SectionInfo("network", "vdom", "ネットワーク", "IF・ゾーン・静的ルート・DHCP・OSPF / BGP"),
    SectionInfo("objects", "vdom", "オブジェクト", "アドレス・サービス・グループ（shared 含む）"),
    SectionInfo("policies", "vdom", "ポリシー", "FW ポリシー・Local-in・条件・UTM 紐付け"),
    SectionInfo("nat", "vdom", "NAT", "VIP / DNAT・IP Pool / SNAT"),
    SectionInfo("vpn", "vdom", "VPN", "IPsec Phase1/2・SSL-VPN / GlobalProtect"),
    SectionInfo(
        "security_profiles", "vdom", "セキュリティプロファイル", "AV・Web・App・IPS・SSL インスペクション"
    ),
]

SECTION_KEYS: List[str] = [section.key for section in SECTION_CATALOG]

PRESETS: Dict[str, Preset] = {
    preset.key: preset
    for preset in (
        Preset("overview", "概要", "ホスト・IF・ポリシー。初見の把握", ["device_info", "network", "policies"]),
        Preset(
            "connectivity",
            "接続",
            "統合ネットワーク図・IF・経路・NAT・VPN",
            ["device_info", "topology", "network", "nat", "vpn"],
        ),
        Preset(
            "policy",
            "ポリシー監査",
            "オブジェクト・ルール・NAT・UTM",
            ["device_info", "objects", "policies", "nat", "security_profiles"],
        ),
        Preset(
            "ops",
            "運用",
            "オブジェクト・VPN・UTM・HA・ログを含む。引き継ぎ・障害対応",
            [
                "device_info",
                "system_settings",
                "network",
                "objects",
                "policies",
                "nat",
                "vpn",
                "security_profiles",
                "ha",
                "logging",
            ],
        ),
        Preset("all", "すべて", "全項目。設計書・納品・差分比較のベース", list(SECTION_KEYS)),
    )
}

DEFAULT_PRESET = "ops"


def resolve_sections(preset: Optional[str] = None, sections: Optional[str] = None) -> Optional[List[str]]:
    """CLI の --preset / --sections を、エクスポータに渡すセクション配列へ解決する。

    --sections を優先する。どちらも無指定なら None（全セクション）を返す。
    """
    if sections:
        keys = [key.strip() for key in sections.split(",") if key.strip()]
        known = set(SECTION_KEYS) | set(COMMON_ALIASES)
        unknown = [key for key in keys if key not in known]
        if unknown:
            raise ValueError(
                f"未知のセクションです: {', '.join(unknown)}（--list-sections で一覧を確認できます）"
            )
        return keys
    if preset:
        if preset not in PRESETS:
            raise ValueError(f"未知のプリセットです: {preset}")
        return list(PRESETS[preset].sections)
    return None
