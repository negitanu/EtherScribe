# 検証記録（2026-10-25）

- Batfish revision: `2f6128b8f113c6406254c1557daad0c2b05fb739`
- 1,166 リソース + LICENSE 1 件 = 1,167 件、922,736 bytes（manifest を除く）。全 Git blob と SHA-256 を確認。
- Cisco 222 / ASA 64 / XR 93 / FortiOS 56 / Juniper 534 / Palo Alto 170 / FRR 27。
- `cd core && go test ./...`: 全成功。
- `./scripts/check-fixture-keys.sh`: 3 機種とも成功。
- `openspec validate add-parser-test-corpus --strict`: 成功。
- `git diff --check`: 成功。

## 検証の限界

外部原典は文法参照として保存したもの。Batfish の解析エンジンや Java テストは実行していない。通常の回帰は独自の合成 fixture のみを使用する。Cisco / PAN-OS set の対応拡張は行っていない。

Go テストは sandbox 内から標準ライブラリ testing が参照できなかったため、権限を拡張して実施し成功した。パーサ本体への変更はない。

コミット前検査: 上流原典の末尾空白・末尾空行もハッシュ一致を優先して保持する。空白検査は `.test/batfish/**` を除いた独自ファイルに適用する。
