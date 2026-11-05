# セキュリティ現状と SBOM 運用計画

調査日: 2026-11-05。アプリケーションの動作は変えていない。照会先は OSV（GHSA / Go vuln DB を含む）と Electron の公開リリース情報。NVD を一次ソースにはしていない。

この文書は Phase 0 の棚卸しである。実装は別の OpenSpec 変更で行う。

## 結論

出荷しているデスクトップの実行系は Electron 33.4.11（Chromium 130.0.6723.191、Node 20.18.3）で、公式サポートは 2025-04-28 に終わっている。2026-11-05 時点のサポート対象は 43 / 44 / 45 で、npm の `latest` は 45.2.3。OSV はこの 33.4.11 に Electron の GHSA を 46 件返す。大半は今の機能（offscreen、webview、拡張、カスタムプロトコル、ログイン項目など）を使っておらず、個別に電話を鳴らす対象ではない。対応の本体は、サポート中の系列へ上げること。

Go エンジンは別である。v1.0.0 の CI は `go1.27.3` でビルドし、同じ版で `govulncheck ./...` を 2026-11-05 に実行した結果は脆弱性なし。OSV が `golang.org/x/text@v0.33.0` に出す 2 件は、本体が import していないパッケージである。

## 対象にした実体

| 面 | 実体 |
| --- | --- |
| Go | `core/go.mod` は Go 1.23、要求は `golang.org/x/text v0.33.0`、`replace` で `core/third_party/text`。本番 import は `encoding/japanese` と `transform` のみ（`core/internal/ingest/encoding.go`） |
| エンジンビルド | `scripts/build-engine.mjs` は `CGO_ENABLED=0`、`-trimpath`、`-buildvcs=false`、`GOPROXY=off` |
| npm | `desktop/package-lock.json` lockfileVersion 3。直接依存はすべて devDependency。421 個の一意なパッケージを OSV に照会 |
| 実行系 | Electron 33.4.11、electron-builder 25.1.8、Vite 6.4.5、esbuild 0.25.12、TypeScript 5.9.3 |
| 配布 | Windows x64 ZIP、Mac arm64 ZIP / DMG。`app.asar` と `extraResources` のエンジン、Go と x/text のライセンス。`scripts/verify-release.mjs` が許可リストを検査 |
| CI | `.github/workflows/release.yml`。Actions はコミット SHA 固定。2026-11-05 に各タグの SHA が公式の checkout/setup-go/setup-node/upload-artifact v7 と download-artifact v8 と一致することを確認 |
| 公開物 | Release `v1.0.0`（2026-10-25）。タグ CI の Go は `go1.27.3` |

## A. 所見

深刻度は、今の配布物で利用者が踏むか、電話を鳴らす価値があるかで付けている。スキャナの CVSS をそのまま深刻度にはしていない。

### 高: サポート切れの Electron を配布している

- 証拠: `desktop/package-lock.json` の `electron@33.4.11`。Electron 公式リリースは、この版を 2025-04-26 出荷、Chromium 130.0.6723.191 / Node 20.18.3 とし、33 系はサポート終了と明記している。サポート終了日は 2025-04-28。v1.0.0 の公開は 2026-10-25 で、終了から約 18 か月後である。
- OSV（2026-11-05）: `pkg:npm/electron@33.4.11` に 46 件。修正版は概ね 38 系以降で、33 系には戻っていない。
- 今のアプリが実際に使っている境界は強い。`desktop/src/main/main.ts` は `contextIsolation: true`、`nodeIntegration: false`、`sandbox: true`。描画側 CSP は `script-src 'self'`。プレビュー iframe は `sandbox=""`。生成 HTML は `html.EscapeString`、Excel は数式ではなく `inlineStr`。本番はローカルの `index.html` だけを開く。
- 46 件のうち、機能がコードに無いものは非該当として VEX する（下の初期判断）。`contextBridge` で Promise を返す形（`desktop/src/preload/preload.ts`）は GHSA-h7rp-cf8h-j98x（CVE-2026-70601）のパターンに一致する。アドバイザリ自身が「信頼できない内容をそのウィンドウに読み込まないアプリは影響を受けない」と書いている。現行の本番 UI はこれに当たる。プレビューは別フレームで、スクリプトは sandbox で動かない。
- それでも高にする理由は、個別 CVE の再現ではなく、この Chromium には今後パッチが来ないことである。製品は第三者の設定ファイルを HTML としてこの Chromium に渡す。パーサの不具合は script タグが無くても届き得る。Electron 固有 GHSA 以外の Chromium CVE は、この調査では 1 件ずつ突合していない。

初期の非該当（機能が無い、またはアドバイザリの前提を満たさない）。Phase 2 の VEX に `not_affected` / `vulnerable_code_not_present` で書く。

- offscreen、webview、DevTools、拡張、USB/HID、カスタムプロトコル、`net.fetch`、ログイン項目、クリップボード、PowerMonitor、Squirrel.Mac: GHSA-532v-xpq5-8h95、GHSA-8x5q-pvf5-64mp、GHSA-h3v5-h239-pg29、GHSA-pfmc-3mgc-p6fp、GHSA-9qh4-3jw8-366w、GHSA-4f78-qhmw-8j8m、GHSA-f2r8-jv7c-xqmp、GHSA-98mx-cjp9-gchm、GHSA-m55f-7gqj-fr98、GHSA-9899-m83m-qhpj、GHSA-fw47-9c62-h3vw、GHSA-4p4r-m79c-wq3v、GHSA-j84w-jfhq-vhvj、GHSA-v3j7-r9gq-3gjw、GHSA-v64r-4m7r-3mvq、GHSA-5rqw-r77c-jp79、GHSA-jfqx-fxh3-c62j、GHSA-mwmh-mq4g-g6gr、GHSA-f37v-82c4-4x64、GHSA-jjp3-mq3x-295m、GHSA-vv43-5jgx-7qv8、GHSA-3c8v-cfp5-9885、GHSA-5c9j-mhmv-5xgx、GHSA-xj5x-m3f3-5x3h、GHSA-xwr5-m59h-vwqr、GHSA-9w97-2464-8783、GHSA-9wfr-w7mm-pc7f、GHSA-3qxv-5xgr-q7w9、GHSA-jr5w-w6x2-84p9、GHSA-jm7p-cc5g-qwxx、GHSA-x8rc-wpg4-grpf、GHSA-2p8c-gr87-2676
- GHSA-vmqv-hx8q-j7mg（ASAR 完全性の回避）は、`embeddedAsarIntegrityValidation` と `onlyLoadAppFromAsar` の fuse を有効にしたアプリだけが対象。今は fuse が無く、この CVE の対象外。fuse を付ける作業は、修正済み Electron へ上げた後に行う。
- 権限ハンドラをインストールしていないため、ハンドラへ渡る origin の取り違え（GHSA-8337-3p73-46f4、GHSA-9pf5-hg6p-4pwp、GHSA-r5p7-gp4j-qhrx、GHSA-x562-297c-g43g）も現行コードでは非該当。拒否ハンドラを足すときは、その不具合が直った Electron 上で足す。

潜伏（パターンはある。現行の逃げた HTML と sandbox ではスクリプトも外部 URL も乗らない）。上げる作業と一緒に塞ぐ。

- GHSA-h7rp-cf8h-j98x、GHSA-ff2p-hmqr-hxm4: `contextBridge` と Promise。本番はローカルの自前 UI。
- GHSA-9f4c-93c8-jc8g、GHSA-p2rr-rvmm-c5fp、GHSA-gr2m-v5gq-v685、GHSA-hq2x-r82h-9wj4、GHSA-997q-gvvc-jwmx、GHSA-f3pv-wv63-48x8、GHSA-v93f-fgjr-hjrj: sandbox iframe からの window / 外部プロトコル。`setWindowOpenHandler` と `setPermissionRequestHandler` が無いので、既定では外部起動が許可され得る。生成 HTML は設定値をテキストノードへエスケープしており、今の出力に攻撃者のタグは乗らない。

### 中: 署名が無く、ウィンドウ境界も閉じていない

- 署名: `desktop/package.json` は Mac `identity: null`、Windows `signAndEditExecutable: false`。CI は `CSC_IDENTITY_AUTO_DISCOVERY=false`。`scripts/sign-mac.cjs` は ad-hoc（`codesign -`）だけで、公証もタイムスタンプも無い。Release notes と `docs/release-process.md` に制約として書いてある。Gatekeeper と SmartScreen は発行元を確認できない。チェックサムは GitHub Release 上の `SHA256SUMS.txt` までで、署名された証明ではない。
- 開発 URL: `main.ts` は `VITE_DEV_SERVER_URL` があればパッケージ済みでも `loadURL` する。そのページには preload の API が載る。通常の配布物の起動では空である。
- ナビゲーションと権限の拒否が無い。上の潜伏項目と同じ穴である。
- Electron fuse（asar 完全性、`RunAsNode` 無効、inspect と `NODE_OPTIONS` の無効化）が無い。署名が無い状態ではローカル改ざんを暗号では止められない。fuse は署名の代わりにはならないが、上げた後の次の段である。

この VM では Apple / Windows の証明書も、署名した実機確認もできない。

### 低: ビルド時の依存と、到達しない Go モジュール

OSV が lockfile に返した Electron 以外は次だけである。直接依存の electron-builder 25.1.8、Vite 6.4.5、esbuild 0.25.12、TypeScript 5.9.3 自体は 0 件。

| パッケージ | 件数 | 判断 |
| --- | --- | --- |
| `tar@6.2.1` | 12 | `@electron/rebuild`、`app-builder-lib`、`cacache`、`node-gyp` が使う。配布 asar の許可リストに入らない。修正は 7.5 系。ビルドが展開するのはチェックサム付きの上流アーカイブであり、利用者の設定ファイルではない |
| `extract-zip@2.0.1` | 2 | `electron` のインストールが公式 zip を展開するときに使う。OSV は `last_affected: 2.0.1` で、修正版のイベントが無い。symlink 経由の書き込み。入力は `@electron/get` が取得する公式配布物 |
| `glob@10.5.0` | 1（GHSA-3g3m-2fjr-6pw4、9.0.0 以上 14.0.2 未満） | ビルドツールのファイル走査。製品の実行パスではない |
| `minimatch@9.0.9` と `@10.2.6` | 1（GHSA-g3hv-75wg-m249、7.0.0 以上 11.0.1 未満） | 同上。ReDoS 級。配布物の外 |
| `app-builder-lib@25.1.8` | GHSA-7g7r-gx96-252g | AppImage の `LD_LIBRARY_PATH`。修正は 26.15.0。このリポジトリの公開物は Windows ZIP と Mac ZIP/DMG で、AppImage は作っていない |
| `builder-util-runtime@9.2.10` | GHSA-p2f4-r6v6-j797 | リダイレクト時に `PRIVATE-TOKEN` や大文字混じりの `Authorization` が残る。修正は 9.7.0。公開は `electron-builder --publish never` と、別ジョブの `gh release upload`。electron-builder にトークンを渡していない |

`golang.org/x/text@v0.33.0`（上流の最新は v0.43.0）:

- GO-2026-5970 / CVE-2026-56852。`unicode/norm` の不正 UTF-8 で無限ループ。修正 v0.39.0。
- GO-2026-6629 / CVE-2026-56851。`secure/precis` の Nickname が短いバッファで panic。修正 v0.41.0。
- 本番 import は Shift_JIS と `transform` だけ。`govulncheck`（Go 1.27.3、`core` で `./...`）は脆弱性なし。モジュール全体をソーススキャンすると 2 件出る。電話を鳴らすのは後者のスキャナをゲートにした場合だけなので、VEX と govulncheck の到達性で止める。上げるなら v0.41.0 以上をまとめて vendor し、警告そのものを消す。

Go 標準ライブラリ: v1.0.0 のエンジンジョブは `go1.27.3`。OSV の `stdlib` 1.27.3 は 0 件。`go-version: stable` のままなので、次のタグは別のパッチになり得る。SBOM にはそのとき入った `go version -m` の版を書く。言語行 `go 1.23` は最小言語版であり、標準ライブラリの版ではない。

`desktop/package.json` の `allowScripts` は electron と esbuild だけを許可する意図だが、CI の `npm ci` はこれを強制しない。lockfile で `hasInstallScript` があるのは `electron`、`esbuild`、任意の `fsevents` だけで、意図とは一致している。将来の推移的依存がスクリプトを足しても CI は気づかない。

公開ジョブのリリースノートは `docs/releases/v1.0.0.md` に固定されている。次のタグでも同じファイルを読む。

### 既に効いているもの

- 描画プロセスは contextIsolation、sandbox、nodeIntegration 無効、CSP、iframe sandbox。
- プレビュー読み取りは、そのセッションで生成した HTML パスだけ（`previewPaths`）。
- 秘匿はフィールド名（`password`、`community`、`secret`、`private_key`、`psk_value` と接尾辞）と `ENC` トークン。PSK / HA パスワード / SNMP community は生値ではなく設定有無の真偽値。出力前に JSON の走査で伏せる。
- エンジンは stdio の子プロセス。待受ポートは無い。デスクトップの本番コードに外向き通信は無い。
- 依存のダウンロードを狭めている。Go は `GOPROXY=off` とローカル replace。npm は lockfile の `npm ci`。Actions は SHA 固定。ジョブ権限は `contents: read`、公開ジョブだけ `contents: write`。`persist-credentials: false`。
- `scripts/audit-release.py` が追跡ファイルと履歴の鍵・トークン形、機微パス、原典 blob を見る。
- `extraResources/licenses` は Go と x/text の LICENSE / PATENTS だけを許可している。Electron / Chromium のライセンスはフレームワーク同梱に依存しており、この許可リストにはコピーしていない。

## B. 対応リスト

### 電話から頼める（Linux のクラウドエージェントが PR を作り、既存の Windows / Mac CI が梱包を見る）

1. Electron を、作業時点でサポート中の最新安定（調査時点では 45 系、npm `latest` 45.2.3）へ上げる。electron-builder は 26 系（AppImage の修正は 26.15.0 以降。公開物は AppImage ではないが、同じ上げの中で lockfile を更新する）。`npm ci` と両 OS の `verify-release` を通す。GUI の最終クリック確認は下の人間作業。
2. パッケージ済みでは `VITE_DEV_SERVER_URL` を無視する。`setWindowOpenHandler` で拒否する。`setPermissionRequestHandler` で拒否する。`will-navigate` をローカルファイル（開発時だけ開発サーバ）に限る。これは潜伏項目の塞ぎであり、Electron を上げた同じ変更に載せる。
3. 上げた後に fuse を付ける。`OnlyLoadAppFromAsar`、`EnableEmbeddedAsarIntegrityValidation`、`RunAsNode` 無効、inspect と `NODE_OPTIONS` 無効。`verify-release` で fuse を見る。fuse を先に 33 系へ付けると、ASAR の GHSA の対象側に入るので、順序は上げてから。
4. vendored `x/text` を v0.41.0 以上（調査時点の最新は v0.43.0）へ置き、`go test ./...` と `govulncheck ./...` を見る。到達しない 2 件をソーススキャンから消すため。
5. `release.yml` の `go-version: stable` を、そのときサポートされているパッチ（v1.0.0 実績は 1.27.3）に固定する。SBOM の再現に使う。
6. CI で `hasInstallScript` が `electron`、`esbuild`、`fsevents` 以外なら失敗させる。`allowScripts` フィールドに頼らない。
7. 公開ジョブのノートを、タグの版に対応する `docs/releases/` にする。
8. 下の Phase 1 を別 PR で入れる。この文書の時点では生成も添付もしていない。

### 人間と証明書が要る

1. Apple Developer ID、公証、staple。クラウド VM にも電話にも秘密鍵は置かない。
2. Windows Authenticode。組織の SmartScreen を通すならコード署名証明書が要る。
3. 署名後に、Mac で Gatekeeper、Windows で SmartScreen を自分の端末で一度見る。CI の ad-hoc `codesign --verify` はこれの代わりにならない。
4. Electron を上げたビルドを、Mac と Windows でファイル選択、生成、プレビュー、差分まで操作する。Linux エージェントは Electron の Mac/Windows バイナリを起動してこの確認はできない。
5. 実機の設定に、今のフィールド名では伏せられない秘密が無いかは、設定ファイルをリポジトリへ置かずに手元で見る。

## C. SBOM の運用

対象ワークフローは、電話で依頼し、Cursor のクラウドエージェントが PR を作り、Codex が差分を見て、GitHub のタグが Mac / Windows のバイナリを Release する、という一人の配布である。Linux エージェントがほとんどの編集をする。署名と実機 GUI だけが人間の手元に残る。

### 形式

一次は CycloneDX JSON（1.6 系）。

- 同じ文書型で SBOM と VEX を扱える。電話で見る「この CVE は非該当」を機械可読にできる。
- 成分の種別で、ロックファイル、Go モジュール、asar 内のファイル、外側のエンジンバイナリ、配布 zip/dmg を分けられる。
- 公開リポジトリなので、GitHub の Dependency submission と Dependency Review に載せられる。
- SPDX は、顧客や法務が SPDX を指定してから一方向に変換する。今は両方作らない。ライセンス本文は既に `extraResources/licenses` にあるので、SBOM はそのファイルをライセンス証拠として指す。

### どの成果物に SBOM を付けるか

二つに分ける。混ぜると 421 パッケージが毎回鳴る。

1. ビルド入力（PR の作業用、Release の本体にはしない）
   - `desktop/package-lock.json` 全体。Electron は devDependency だが実行系そのものなので、`omit dev` で落としてはいけない。
   - Go は `cmd/etherscribe` から辿ったモジュールグラフと、ビルドに使った `go version`。`core/third_party/text` の全ファイルや、`go.sum` のテスト専用（`go-cmp`、`x/tools` など）を成分にしない。`govulncheck` が「脆弱性なし」とした事実と、モジュール版 v0.33.0 がバイナリの build info に出ることの両方を残す。
2. 配布ランタイム（Release に付けるもの）
   - アプリ版。
   - Electron 実行系（npm 版、Chromium、Node）。
   - `app.asar` のハッシュ。中の node_modules 展開はしない。verify の許可リスト上、中身は UI と main/preload と package.json だけである。
   - `extraResources/engine` のバイナリハッシュと `go version -m`。
   - 既に同梱している 3 つのライセンスファイルをライセンス証拠として。
   - Windows ZIP、Mac ZIP、Mac DMG のハッシュ。既存の `SHA256SUMS.txt` と同じ値にする。

### いつ作るか

両方。ゲートの厳しさは分ける。

- タグの公開（Phase 1 を先にやる）。利用者が受け取る物の SBOM が先。PR が赤くなる仕組みより、Release の横にファイルがある方が一人の配布では役に立つ。
- PR（Phase 2）。差分の新規ランタイム問題だけを失敗にする。既存の 46 件で main を赤くしない。Electron を上げてからゲートを有効にする。

Actions の artifact はエンジン 7 日、デスクトップ 14 日で消える。正本は Release の資産であり、リポジトリに生成 SBOM をコミットしない。

### どこに置くか

- GitHub Release 資産: `EtherScribe-<version>-sbom.cdx.json` を 3 つの配布物と `SHA256SUMS.txt` の隣に置く。公開ジョブは既に `contents: write` を持っている。
- リポジトリにコミットしてよいのは人手の VEX だけ（Phase 2 で `docs/security/vex.cdx.json`）。生成物はコミットしない。
- 証明: 公開ジョブに `id-token: write` と `attestations: write` を足し、zip / dmg / SBOM へ GitHub artifact attestation を付ける。これは Actions がそのファイルを作ったことの証明で、Apple 公証や Authenticode の代わりではない。証明書が要らず、電話から頼める。

### 日常の見方

最小の組は次の三つだけ。`npm audit`、Snyk、Trivy、Grype は重ねない。OSV が npm と Go の照会先として既に足りている。

| 道具 | 役割 |
| --- | --- |
| `govulncheck ./...` | Go の到達性。vendored x/text の「モジュールは該当、バイナリは非該当」を維持する |
| `osv-scanner` の lockfile モード | npm。PR では新規の実行系（Electron）だけを失敗候補にし、tar / glob / minimatch は VEX |
| Dependency Review Action | 公開リポジトリで追加料金なし。PR が lockfile に足したパッケージだけを Codex と電話に見せる |

Dependabot は `npm` と `github-actions` だけ有効にする。Go は `replace` のローカルツリーなので、Dependabot の gomod 更新に期待しない。x/text の追従は Phase 3 の再スキャンか、明示的な vendor 更新 PR にする。Dependabot のセキュリティ PR は、VEX 済みのビルド専用 CVE では作らせない（無視ルールに ID を書く）。

### VEX

CycloneDX VEX。状態は `not_affected` を主に使う。正当化は次のどちらか。

- `vulnerable_code_not_present`: 機能が無い、配布物に入らない、AppImage を作っていない、electron-builder にトークンを渡していない。
- `vulnerable_code_not_in_execute_path`: x/text の norm / precis。根拠は `govulncheck` の「脆弱性なし」と、本番 import が `encoding/japanese` と `transform` だけであること。

`affected` のまま放置して通知だけ弱くする、は使わない。鳴らない理由を文書に残す。新しい High で VEX が無いものだけを、週次ジョブが issue にする。Codex レビューでは VEX の追加を、アドバイザリの「影響を受ける条件」とこのリポジトリのコードの両方に照らして見る。

初期判断は上の A に書いた。機械可読ファイルは Phase 2 で、この表を写す。

### 段階

Phase 0。この文書。ツールは入れない。

Phase 1。タグの公開ジョブがランタイム CycloneDX を作り、Release に載せる。失敗条件は「SBOM が 3 配布物のハッシュと `SHA256SUMS.txt` と一致しない」ことであり、CVE の有無では落とさない。可能なら artifact attestation も付ける。ビルド入力の SBOM は Actions artifact に残すだけで、Release には混ぜない。

Phase 2。Electron をサポート中の系列へ上げた後に、PR ゲートを足す。`govulncheck`、`osv-scanner`、Dependency Review。失敗は「VEX に無く、実行系（Electron か、エンジンが到達する Go）の新規問題」だけ。`docs/security/vex.cdx.json` をこの段階でコミットする。

Phase 3。週 1 回、最新 Release のランタイム SBOM を `osv-scanner --sbom` で読み直す。非該当でない High 以上だけ issue を開く。毎日にしない。利用者の手元にある版を見るのが目的で、main の lockfile だけを見ても署名前の作業ツリーとずれる。

### まだやらないこと

- Dependency-Track や Guac の自前ホスト。常時サーバ、認証、バックアップが、週次の 1 ジョブより重い。顧客がポータルを要求してからでよい。
- SPDX と CycloneDX の二重生成。
- `npm audit` と osv-scanner の併用。CI の `npm ci --no-audit` は、監査を osv-scanner に寄せる前提で維持してよい。
- 46 件すべてで PR を赤くすること。上げる PR の前にゲートを入れると、電話がノイズで埋まる。
- 生成 SBOM の git コミット。
- vendored x/text の全パッケージを実行系成分にすること。
- クラウドエージェントでの署名、公証、実設定を使った秘匿の確認。
