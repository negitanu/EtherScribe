# 公式 CLI に基づくネットワーク解析の精度と情報量の改善

## Why
現行 Junos 実装は peer-as を自 AS と混同し、静的ルートを行単位で重複させ、static LAG まで LACP 有効としている。既存機種の複数アドレス・IPv6・ルーティングコンテキストも不足している。

## What Changes
- Junos: BGP 継承・AS 分離、ルート属性統合、LACP 明示設定、複数アドレスと決定的順序を修正する。
- FortiOS: secondary IP / IPv6 / aggregate・LACP / VRF / static route 属性を抽出する。
- PAN-OS XML: IPv6 / MTU / virtual-router と IF の関連を抽出する。
- 加算的モデルと JSON/HTML/Excel の日本語項目、公式出典と合成回帰を追加する。

## Scope
Cisco 新規パーサ、PAN-OS set、Junos groups 展開は今回の範囲外。既存の対応済み機種を優先する。
