# セキュリティ現状と SBOM 運用計画

調査日: 2026-11-05。アプリケーションの動作は変えていない。照会先は OSV（GHSA / Go vuln DB を含む）と Electron の公開リリース情報。NVD を一次ソースにはしていない。

この文書は Phase 0 の棚卸しである。Electron の上げは 2026-11-06 の変更で行った。残りの実装は別の OpenSpec 変更で行う。

## 実施状況（2026-11-06）

main の lockfile は Electron 45.2.3（Chromium 156.0.8078.80、Node 24.22.2）と electron-builder 26.18.1 である。公開済み Release v1.0.0 の ZIP / DMG は 33.4.11 のままなので、利用者が受け取る実行系が替わるのは次のタグからである。

選定は公式スケジュール（2026-11-06）による。サポート対象は 43（EOL 2027-01-05）、44（EOL 2027-03-02）、45（EOL 2027-04-27）。46 は alpha で、安定予定は 2027-01-05 なので採用していない。残サポートが最も長い最新安定として 45.2.3 を選んだ。npm の `latest` も 45.2.3 である。electron-builder は 26 系の最新 26.18.1（`app-builder-lib` 26.18.1、`builder-util-runtime` 9.7.1）。AppImage は作らないが、計画どおり 26.15.0 以降に揃えた。`vite-plugin-electron` は `vite.config.ts` が使っておらず、Electron の peer も無いので上げていない。

この変更で閉じたもの:

- `pkg:npm/electron@33.4.11` の OSV 46 件。2026-11-06 の再照会で `pkg:npm/electron@45.2.3` は 0 件。
- `app-builder-lib` の AppImage 向け GHSA-7g7r-gx96-252g（修正 26.15.0）と、`builder-util-runtime` の GHSA-p2f4-r6v6-j797（修正 9.7.0）。lockfile の該当パッケージは修正版以上で、同じ照会に出ない。
- `tar@6.2.1` の 12 件。lockfile は `tar@7.5.22` になり、OSV の lockfile スキャンに出ない。
- `extract-zip@2.0.1`。依存から外れ、Electron の展開は `@electron-internal/extract-zip@1.0.8`。
- `glob@10.5.0` の GHSA-3g3m-2fjr-6pw4。残る glob は 7.2.3 で、このアドバイザリの範囲外である。

まだ lockfile に残る npm の OSV は `minimatch@9.0.9` と `@10.2.6` の GHSA-g3hv-75wg-m249（7.0.0 以上 11.0.1 未満）だけである。`@electron/asar`、`@electron/universal`、`app-builder-lib`、`filelist` が使うビルド時の ReDoS で、配布 asar には入らない。`npm audit` は moderate 6 件で、依存関係の表示はすべてこの minimatch に行き着く。監査が提案する修正は electron-builder 24.13.3 へのダウングレードなので採用していない。

破壊的変更は Electron 34 から 45 の公式一覧を、`desktop/src` の呼び出しに照らした。

- 残している境界は `contextIsolation: true`、`nodeIntegration: false`、`sandbox: true`。IPC は `ipcMain.handle`、`ipcRenderer.invoke`、`ipcRenderer.on`、`contextBridge.exposeInMainWorld`。ダイアログは `showOpenDialog` と `showSaveDialog`。remote、BrowserView、`webContents.getPrinters`、旧 `protocol.register*Protocol`、同期 `safeStorage`、renderer からの clipboard は呼んでいない。
- Electron 45 は sandbox 済み preload から Node の `events` / `timers` / `url` シムと、`Buffer` / `setImmediate` / `clearImmediate` を削除した。`require` が読めるのは `electron` だけである。コンパイル後の `preload.js` は `require("electron")` だけで、この削除には当たらない。`ipcRenderer` の EventEmitter はネイティブ実装に替わったが、`on` と `invoke` の使い方は同じである。
- Electron 43 は `defaultPath` を省略したダイアログの初期ディレクトリを Downloads にする。このアプリはファイル名を `defaultPath` に渡すので、省略時の変更には当たらない。OS のファイルダイアログが最初に開くフォルダは、この VM では未確認である。ヘッドレスのスモークはダイアログを差し替えている。
- Electron 42 以降、`electron` の npm パッケージに postinstall は無い。バイナリは最初の `electron` 起動時に公式 zip を取得する。`npm ci` はその zip を置かない。パッケージングは electron-builder が別途取得する。`package.json` の `allowScripts` から `electron@33.4.11` を外した。ホストの Node は `>= 22.12.0` が要る。CI の `node-version: '22'` は、2026-11-06 の最新 22 なら満たす。この VM は 22.23.4 だった。
- electron-builder 26 は `afterPack` のパスが `desktop/` の外だと失敗する。入口を `desktop/build/sign-mac.cjs` にし、実体は従来の `scripts/sign-mac.cjs` を呼ぶ。darwin 以外は何もしない ad-hoc のままである。Developer ID、公証、Authenticode は付けていない。`mac.identity: null` と `win.signAndEditExecutable: false` もそのままである。
- lockfile の `hasInstallScript` は `esbuild@0.25.12`、任意の `fsevents@2.3.3`（darwin）、`electron-winstaller@5.4.0` である。最後のスクリプトは同梱の 7-Zip をホスト arch 向けにコピーするだけで、Squirrel 用である。公開ターゲットは ZIP / DMG なので、このスクリプトは配布物を作る経路では使わない。対応リスト 6 の CI 検査は未着手で、検査を入れるときはこの 3 つを想定する。

Linux では `./scripts/build.sh`（`electron-builder --dir`、Electron 45.2.3）が通り、その後の Xvfb スモークは終了コード 0 と `PASS real Electron IPC` だった。D-Bus 接続失敗は出るが、検査は完走する。electron-builder 26 が Mac の `.app` に入れる `LICENSE.electron.txt` と `LICENSES.chromium.html` は、`verify-release` の許可リストに加えた。ファイル選択から差分までの実機操作と、署名後の Gatekeeper / SmartScreen は未確認である。

対応リストの 1 はこの変更で閉じる。2 以降（ナビゲーションと権限の拒否、fuse、x/text、Go 版の固定、install script の CI、公開ノート、SBOM）は未着手である。ウィンドウ境界と fuse は、上げたあとに塞ぐ予定のまま残している。

## 結論

2026-11-05 時点で出荷していたデスクトップの実行系は Electron 33.4.11（Chromium 130.0.6723.191、Node 20.18.3）で、公式サポートは 2025-04-28 に終わっている。同日時点のサポート対象は 43 / 44 / 45 で、npm の `latest` は 45.2.3。OSV はこの 33.4.11 に Electron の GHSA を 46 件返す。大半は今の機能（offscreen、webview、拡張、カスタムプロトコル、ログイン項目など）を使っておらず、個別に電話を鳴らす対象ではない。対応の本体は、サポート中の系列へ上げることであり、2026-11-06 に main の lockfile を 45.2.3 へ更新した。公開物 v1.0.0 は未更新である。

Go エンジンは別である。v1.0.0 の CI は `go1.27.3` でビルドし、同じ版で `govulncheck ./...` を 2026-11-05 に実行した結果は脆弱性なし。OSV が `golang.org/x/text@v0.33.0` に出す 2 件は、本体が import していないパッケージである。

## 対象にした実体

| 面 | 実体 |
| --- | --- |
| Go | `core/go.mod` は Go 1.23、要求は `golang.org/x/text v0.33.0`、`replace` で `core/third_party/text`。本番 import は `encoding/japanese` と `transform` のみ（`core/internal/ingest/encoding.go`） |
| エンジンビルド | `scripts/build-engine.mjs` は `CGO_ENABLED=0`、`-trimpath`、`-buildvcs=false`、`GOPROXY=off` |
| npm | `desktop/package-lock.json` lockfileVersion 3。直接依存はすべて devDependency。421 個の一意なパッケージを OSV に照会 |
| 実行系 | 調査時点は Electron 33.4.11、electron-builder 25.1.8、Vite 6.4.5、esbuild 0.25.12、TypeScript 5.9.3。2026-11-06 以降の lockfile は実施状況のとおり |
| 配布 | Windows x64 ZIP、Mac arm64 ZIP / DMG。`app.asar` と `extraResources` のエンジン、Go と x/text のライセンス。`scripts/verify-release.mjs` が許可リストを検査 |
| CI | `.github/workflows/release.yml`。Actions はコミット SHA 固定。2026-11-05 に各タグの SHA が公式の checkout/setup-go/setup-node/upload-artifact v7 と download-artifact v8 と一致することを確認 |
| 公開物 | Release `v1.0.0`（2026-10-25）。タグ CI の Go は `go1.27.3` |

## A. 所見

深刻度は、今の配布物で利用者が踏むか、電話を鳴らす価値があるかで付けている。スキャナの CVSS をそのまま深刻度にはしていない。

### 高: サポート切れの Electron を配布している

2026-11-06 に main の lockfile は 45.2.3 へ上がり、この所見の lockfile 側は閉じた。公開済み v1.0.0 のバイナリは 33.4.11 のままである。以下は上げる前の記録である。

- 証拠（上げる前）: `desktop/package-lock.json` の `electron@33.4.11`。Electron 公式リリースは、この版を 2025-04-26 出荷、Chromium 130.0.6723.191 / Node 20.18.3 とし、33 系はサポート終了と明記している。サポート終了日は 2025-04-28。v1.0.0 の公開は 2026-10-25 で、終了から約 18 か月後である。
- OSV（2026-11-05）: `pkg:npm/electron@33.4.11` に 46 件。修正版は概ね 38 系以降で、33 系には戻っていない。
- 今のアプリが実際に使っている境界は強い。`desktop/src/main/main.ts` は `contextIsolation: true`、`nodeIntegration: false`、`sandbox: true`。描画側 CSP は `script-src 'self'`。プレビュー iframe は `sandbox=""`。生成 HTML は `html.EscapeString`、Excel は数式ではなく `inlineStr`。本番はローカルの `index.html` だけを開く。
- 46 件を、アプリがその API を呼んでいないという理由で `code_not_present` にまとめない。脆弱なコードは Electron のバイナリに含まれたままである。未使用の機能は、脆弱なコードが成果物から除かれたことではない。VEX の列挙値と、一件ごとの `analysis.detail` は下の「VEX」に従う。条件を detail に書けない ID は `in_triage` のままにする。
- `contextBridge` で Promise を返す形（`desktop/src/preload/preload.ts`）は GHSA-h7rp-cf8h-j98x（CVE-2026-70601）のパターンに一致する。アドバイザリは「信頼できない内容をそのウィンドウに読み込まないアプリは影響を受けない」と書く。現行の本番 UI はローカルの自前ページである。プレビューは別フレームで、スクリプトは sandbox で動かない。この一件を `not_affected` にするときの justification は `requires_environment`（そのウィンドウに信頼できない文書を載せていない）であり、`code_not_present` ではない。
- それでも高にする理由は、個別 CVE の再現ではなく、この Chromium には今後パッチが来ないことである。製品は第三者の設定ファイルを HTML としてこの Chromium に渡す。パーサの不具合は script タグが無くても届き得る。Electron 固有 GHSA 以外の Chromium CVE は、この調査では 1 件ずつ突合していない。

調査時点の振り分けは、単一の justification で埋めない。Phase 2 で ID ごとに detail を書いてから `not_affected` にする。

- 次の ID は、アドバイザリが特定の API や `webPreferences` を条件にしている。`desktop/src` はその API を呼んでいない。正当化の候補は `requires_configuration`（その設定が要る）か、呼ばない経路だと説明できるときの `code_not_reachable` である。detail にアドバイザリの条件と、リポジトリの該当ファイルを書く。`code_not_present` は使わない: GHSA-532v-xpq5-8h95、GHSA-8x5q-pvf5-64mp、GHSA-h3v5-h239-pg29、GHSA-pfmc-3mgc-p6fp、GHSA-9qh4-3jw8-366w、GHSA-4f78-qhmw-8j8m、GHSA-f2r8-jv7c-xqmp、GHSA-98mx-cjp9-gchm、GHSA-m55f-7gqj-fr98、GHSA-9899-m83m-qhpj、GHSA-fw47-9c62-h3vw、GHSA-4p4r-m79c-wq3v、GHSA-j84w-jfhq-vhvj、GHSA-v3j7-r9gq-3gjw、GHSA-v64r-4m7r-3mvq、GHSA-5rqw-r77c-jp79、GHSA-jfqx-fxh3-c62j、GHSA-mwmh-mq4g-g6gr、GHSA-f37v-82c4-4x64、GHSA-jjp3-mq3x-295m、GHSA-vv43-5jgx-7qv8、GHSA-3c8v-cfp5-9885、GHSA-5c9j-mhmv-5xgx、GHSA-xj5x-m3f3-5x3h、GHSA-xwr5-m59h-vwqr、GHSA-9w97-2464-8783、GHSA-9wfr-w7mm-pc7f、GHSA-3qxv-5xgr-q7w9、GHSA-jr5w-w6x2-84p9、GHSA-jm7p-cc5g-qwxx、GHSA-x8rc-wpg4-grpf、GHSA-2p8c-gr87-2676
- GHSA-vmqv-hx8q-j7mg（ASAR 完全性の回避）は、二つの fuse を有効にしたアプリだけが対象と書いている。正当化は `requires_configuration`。今は fuse が無い。コードがバイナリから消えたわけではない。fuse を付ける作業は、修正済み Electron へ上げた後に行う。
- 権限ハンドラへ渡る origin の取り違え（GHSA-8337-3p73-46f4、GHSA-9pf5-hg6p-4pwp、GHSA-r5p7-gp4j-qhrx、GHSA-x562-297c-g43g）は、ハンドラをインストールしていることが条件である。正当化は `requires_configuration`。拒否ハンドラを足すときは、その不具合が直った Electron 上で足す。

次の ID は 33.4.11 の記録である。45.2.3 の OSV 照会には出ない。`setWindowOpenHandler` と `setPermissionRequestHandler` が無いことは対応リスト 2 として残っており、この上げでは塞いでいない。当時の出力は、逃げた HTML と sandbox ではスクリプトも外部 URL も乗らない。

- GHSA-h7rp-cf8h-j98x、GHSA-ff2p-hmqr-hxm4: `contextBridge` と Promise。本番はローカルの自前 UI。
- GHSA-9f4c-93c8-jc8g、GHSA-p2rr-rvmm-c5fp、GHSA-gr2m-v5gq-v685、GHSA-hq2x-r82h-9wj4、GHSA-997q-gvvc-jwmx、GHSA-f3pv-wv63-48x8、GHSA-v93f-fgjr-hjrj: sandbox iframe からの window / 外部プロトコル。`setWindowOpenHandler` と `setPermissionRequestHandler` が無いので、既定では外部起動が許可され得る。生成 HTML は設定値をテキストノードへエスケープしており、今の出力に攻撃者のタグは乗らない。

### 中: 署名が無く、ウィンドウ境界も閉じていない

- 署名: `desktop/package.json` は Mac `identity: null`、Windows `signAndEditExecutable: false`。CI は `CSC_IDENTITY_AUTO_DISCOVERY=false`。`scripts/sign-mac.cjs` は ad-hoc（`codesign -`）だけで、公証もタイムスタンプも無い。electron-builder 26 以降の入口は `desktop/build/sign-mac.cjs` で、中身は同じスクリプトである。Release notes と `docs/release-process.md` に制約として書いてある。Gatekeeper と SmartScreen は発行元を確認できない。チェックサムは GitHub Release 上の `SHA256SUMS.txt` までで、署名された証明ではない。
- 開発 URL: `main.ts` は `VITE_DEV_SERVER_URL` があればパッケージ済みでも `loadURL` する。そのページには preload の API が載る。通常の配布物の起動では空である。
- ナビゲーションと権限の拒否が無い。上の潜伏項目と同じ穴である。
- Electron fuse（asar 完全性、`RunAsNode` 無効、inspect と `NODE_OPTIONS` の無効化）が無い。署名が無い状態ではローカル改ざんを暗号では止められない。fuse は署名の代わりにはならない。実行系は 45.2.3 になったので、対応リスト 3 として次に付けられる。この上げでは付けていない。

この VM では Apple / Windows の証明書も、署名した実機確認もできない。

### 低: ビルド時の依存と、到達しない Go モジュール

次の表は 2026-11-05 の lockfile である。Electron と electron-builder を上げたあとの残りは、上の実施状況にある（minimatch の 1 件だけが OSV に残る）。

OSV が当時の lockfile に返した Electron 以外は次だけである。直接依存の electron-builder 25.1.8、Vite 6.4.5、esbuild 0.25.12、TypeScript 5.9.3 自体は 0 件だった。

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
- 本番 import は Shift_JIS と `transform` だけ。`govulncheck`（Go 1.27.3、`core` で `./...`）は脆弱性なし。モジュール自体はランタイム SBOM の成分として残す（下の Phase 1）。`not_affected` にするときの justification は `code_not_reachable` であり、モジュールを SBOM から消すことでも `code_not_present` でもない。上げるなら v0.41.0 以上をまとめて vendor し、成分の版とハッシュを更新する。

Go 標準ライブラリ: v1.0.0 のエンジンジョブは `go1.27.3`。OSV の `stdlib` 1.27.3 は 0 件。`go-version: stable` のままなので、次のタグは別のパッチになり得る。ランタイム SBOM では、そのビルドの版を `pkg:golang/stdlib@<版>` の library 成分にする。言語行 `go 1.23` は最小言語版であり、標準ライブラリの版ではない。

2026-11-05 時点の `allowScripts` は electron と esbuild だけを許可する意図だが、CI の `npm ci` はこれを強制しない。当時の lockfile で `hasInstallScript` があるのは `electron`、`esbuild`、任意の `fsevents` だけだった。2026-11-06 の lockfile では内訳が変わっている。上の実施状況を参照する。将来の推移的依存がスクリプトを足しても CI は気づかない。

公開ジョブのリリースノートは `docs/releases/v1.0.0.md` に固定されている。次のタグでも同じファイルを読む。

### 既に効いているもの

- 描画プロセスは contextIsolation、sandbox、nodeIntegration 無効、CSP、iframe sandbox。
- プレビュー読み取りは、そのセッションで生成した HTML パスだけ（`previewPaths`）。
- 秘匿はフィールド名（`password`、`community`、`secret`、`private_key`、`psk_value` と接尾辞）と `ENC` トークン。PSK / HA パスワード / SNMP community は生値ではなく設定有無の真偽値。出力前に JSON の走査で伏せる。
- エンジンは stdio の子プロセス。待受ポートは無い。デスクトップの本番コードに外向き通信は無い。
- 依存のダウンロードを狭めている。Go は `GOPROXY=off` とローカル replace。npm は lockfile の `npm ci`。Actions は SHA 固定。ジョブ権限は `contents: read`、公開ジョブだけ `contents: write`。`persist-credentials: false`。
- `scripts/audit-release.py` が追跡ファイルと履歴の鍵・トークン形、機微パス、原典 blob を見る。
- `extraResources/licenses` は Go と x/text の LICENSE / PATENTS だけを許可している。electron-builder 26 は、Mac の zip / dmg に入るのが `.app` だけなので、Electron の `LICENSE` と Chromium の `LICENSES.chromium.html` を `Contents/Resources` へ移す。`verify-release` はその 2 ファイル（`LICENSE.electron.txt` と `LICENSES.chromium.html`）を darwin の許可リストに入れ、空ファイルは拒否する。Go のライセンス置き場には混ぜない。

## B. 対応リスト

### 電話から頼める（Linux のクラウドエージェントが PR を作り、既存の Windows / Mac CI が梱包を見る）

1. 完了（2026-11-06、main の lockfile）。Electron 45.2.3 と electron-builder 26.18.1。Linux の `npm ci`、`electron-builder --dir`、Xvfb の IPC スモークは確認した。Windows / Mac の `verify-release` と GUI の最終クリック確認は、CI と下の人間作業に残る。
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
   - Go は `cmd/etherscribe` から辿ったモジュールグラフと、ビルドに使った Go の版。`core/third_party/text` のソースファイルを 1 ファイルずつ成分にしない。`go.sum` のテスト専用（`go-cmp`、`x/tools` など）も成分にしない。ランタイム SBOM が持つモジュール成分（次項）と、このビルド入力は同じ purl を指す。
2. 配布ランタイム（Release に付けるもの）
   - アプリ版。
   - Electron 実行系（npm 版、Chromium、Node）。purl は `pkg:npm/electron@<版>`。
   - `app.asar` のハッシュ。中の node_modules 展開はしない。verify の許可リスト上、中身は UI と main/preload と package.json だけである。
   - `extraResources/engine` のファイル成分（バイナリの SHA-256）。`go version -m` の原文はこのファイル成分の証拠に残す。モジュール成分の代わりにはしない。
   - `go version -m` が列挙する各モジュールを、ファイル成分とは別の `type: library` にする。標準ライブラリと、replace された `golang.org/x/text` を必ず含める。サブパッケージ（`unicode/norm` など）には分けない。モジュール 1 つが成分 1 つである。各成分は次を持つ。
     - `name` と `version`。replace のディレクトリ名ではなく、ビルド info のモジュールパスと版。現行なら `golang.org/x/text` と `v0.33.0`、標準ライブラリはビルドに使った版（v1.0.0 の実績は `1.27.3`）。
     - `purl`。`pkg:golang/golang.org/x/text@v0.33.0`、標準ライブラリは `pkg:golang/stdlib@v1.27.3`。週次の再スキャンはこの purl を OSV へ照会する。
     - `hashes` の SHA-256 と、プロパティ `cdx:go:module-hash`。値は Go のモジュール dirhash（`h1:` + base64）。`hashes[].content` には、その base64 を復号したダイジェストを十六進で書く。`go version -m` に `h1:` が無い replace モジュールは、リンクしたツリー（`x/text` なら `core/third_party/text`）から `dirhash.Hash1` を計算して同じ二つを埋める。このハッシュはモジュールのファイル一覧のものであり、エンジンバイナリの SHA-256 ではない。
   - 既に同梱している 3 つのライセンスファイルをライセンス証拠として。
   - Windows ZIP、Mac ZIP、Mac DMG のハッシュ。既存の `SHA256SUMS.txt` と同じ値にする。

### いつ作るか

両方。ゲートの厳しさは分ける。

- タグの公開（Phase 1 を先にやる）。利用者が受け取る物の SBOM が先。PR が赤くなる仕組みより、Release の横にファイルがある方が一人の配布では役に立つ。
- PR（Phase 2）。差分の新規ランタイム問題だけを失敗にする。Electron は 45.2.3 へ上がったので、ゲートを有効にできる状態になった。33.4.11 の 46 件は lockfile から外れている。ゲート自体はこの上げには入れていない。残る minimatch で失敗させない。

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

Dependabot のバージョン更新は `npm` と `github-actions` だけ有効にする。Go は `replace` のローカルツリーなので、gomod の自動更新には期待しない。x/text の追従は、Phase 3 がランタイム SBOM の purl を再照会するか、明示的な vendor 更新 PR で行う。

脆弱性の扱いと、バージョン更新の無視は別の台帳にする。

- 台帳 A はアドバイザリ単位の triage と VEX である。一件の GHSA / CVE を `docs/security/vex.cdx.json` に書く。GitHub の Dependabot alert を閉じるときも、その alert だけを閉じる。同じパッケージの次の CVE は新しい alert として残る。
- 台帳 B は `dependabot.yml` の `ignore` である。キーは `dependency-name` と、任意の `versions` または `update-types` だけである。脆弱性 ID を ignore のキーにしない。このスキーマに GHSA や CVE の欄は無い。パッケージ名だけを ignore して、その依存の今後の CVE まで見えなくしない。ignore を書くときは、版の範囲か `update-types`（例: `version-update:semver-patch`）まで指定し、コメントに「バージョン更新の騒音であり、CVE の判断ではない」と書く。
- VEX 済みのビルド専用 CVE を、`ignore` で黙らせない。tar や minimatch を依存名だけで ignore しない。

### VEX

CycloneDX 1.6 の `vulnerabilities[].analysis` を使う。`state` が `not_affected` のとき、`justification` は仕様の列挙だけを使う。OpenVEX の `vulnerable_code_not_present` と `vulnerable_code_not_in_execute_path` は書かない。

| justification | 仕様上の意味 | このリポジトリで使ってよい場面 |
| --- | --- | --- |
| `code_not_present` | 脆弱なコードが成果物から除かれている | そのパッケージがランタイム SBOM に無く、配布 asar にも無いとき。例: tar、extract-zip、glob、minimatch の JS。Electron の未使用 API には使わない |
| `code_not_reachable` | コードは成果物にあるが、実行時に呼ばれない | `govulncheck` が到達なしとした Go のシンボル。例: モジュール成分として残す `golang.org/x/text` の norm / precis。detail に import と govulncheck の版を書く |
| `requires_configuration` | 悪用に特定の設定のオンまたはオフが要る | アドバイザリが API、`webPreferences`、fuse を条件にしており、その設定が `desktop/src` に無いとき。例: GHSA-vmqv-hx8q-j7mg |
| `requires_dependency` | 悪用に、存在しない依存が要る | 条件の依存がランタイム SBOM の purl に無いとき。未使用機能の言い換えには使わない |
| `requires_environment` | 悪用に、存在しない環境が要る | 例: GHSA-h7rp-cf8h-j98x は信頼できない文書をそのウィンドウへ載せることが条件で、本番はローカルの自前 UI である。AppImage 向け GHSA は、公開物が ZIP/DMG であるとき |
| `protected_at_runtime` | 実行時に悪用が止まる | アドバイザリが実行時の防御を条件にしており、その防御がコードにあるとき |
| `protected_by_mitigating_control` | 緩和策が当たっている | 生成 HTML のエスケープと iframe の `sandbox=""` のように、出力側の制御で発動条件を満たさないとき。制御のファイル名を detail に書く |
| `protected_by_compiler` | コンパイラフラグが要る | このリポジトリの現行判断では使っていない |
| `protected_at_perimeter` | 境界で止まっている | このリポジトリの現行判断では使っていない |

`not_affected` の各エントリは、`analysis.detail` に次の三つを書く。アドバイザリが述べる影響条件。このリポジトリのどのファイルがその条件を満たすか、満たさないか。選んだ justification がその条件に対して主張すること。三つを書けない ID は `not_affected` にしない。`in_triage` のまま週次の issue に残す。

Electron の GHSA を機能未使用の一括理由で `code_not_present` にしない。週次スキャンは、モジュールと Electron の purl が SBOM にある限り、その成分を照会し続ける。VEX は通知を止める判断であり、成分を SBOM から消す操作ではない。

Codex レビューでは、追加された VEX を、アドバイザリの影響条件、上記の列挙、リポジトリの該当行の三つが一致するときだけ受け入れる。

### 段階

Phase 0。この文書。ツールは入れない。

Phase 1。タグの公開ジョブがランタイム CycloneDX を作り、Release に載せる。失敗条件は二つある。3 配布物のハッシュが `SHA256SUMS.txt` と一致しないこと。そして `go version -m` の各モジュール、少なくとも `pkg:golang/stdlib@<ビルドの版>` と `pkg:golang/golang.org/x/text@<リンクした版>` が、purl と SHA-256 を持つ library 成分として欠けていること。エンジンファイルのハッシュだけがある SBOM は失敗にする。CVE の有無では落とさない。可能なら artifact attestation も付ける。ビルド入力の SBOM は Actions artifact に残すだけで、Release には混ぜない。

Phase 2。Electron をサポート中の系列へ上げた後に、PR ゲートを足す。`govulncheck`、`osv-scanner`、Dependency Review。失敗は「VEX に無く、実行系（Electron の purl か、ランタイム SBOM に載る Go モジュールの purl）の新規問題」だけ。`docs/security/vex.cdx.json` をこの段階でコミットする。justification は上の表以外を受け付けない。

Phase 3。週 1 回、最新 Release のランタイム SBOM を `osv-scanner --sbom` で読み直す。照会対象は SBOM 内の purl である。`pkg:golang/golang.org/x/text@...` が成分にあるので、x/text の新しいアドバイザリはこの再スキャンに乗る。バイナリハッシュや `go version -m` の文章だけではヒットしたことにしない。モジュール成分が欠けた SBOM はスキャン成功と扱わず、欠落を issue にする。`not_affected` の detail が無い High 以上だけ、脆弱性の issue を開く。毎日にしない。利用者の手元にある版を見るのが目的で、main の lockfile だけを見ても署名前の作業ツリーとずれる。

### まだやらないこと

- Dependency-Track や Guac の自前ホスト。常時サーバ、認証、バックアップが、週次の 1 ジョブより重い。顧客がポータルを要求してからでよい。
- SPDX と CycloneDX の二重生成。
- `npm audit` と osv-scanner の併用。CI の `npm ci --no-audit` は、監査を osv-scanner に寄せる前提で維持してよい。
- 33.4.11 にあった Electron の 46 件で PR を赤くすること。それらは 45.2.3 の lockfile からは外れた。PR ゲート自体は Phase 2 のままで、この上げの中には入れていない。残る minimatch で main を赤くしない。
- 生成 SBOM の git コミット。
- vendored x/text のサブパッケージを成分にばらすこと。モジュール 1 成分（purl とハッシュ）は Phase 1 の必須であり、この禁止には入らない。
- クラウドエージェントでの署名、公証、実設定を使った秘匿の確認。
