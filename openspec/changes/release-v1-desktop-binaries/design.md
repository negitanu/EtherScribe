# 設計

GitHub Actions の検証ジョブで追跡ソース・原典ハッシュ・版・Go 回帰を確認し、CGO_ENABLED=0 / trimpath / buildvcs=false で Windows amd64 と Darwin arm64 のエンジンをクロスビルドする。各 OS ランナーへ最小限のエンジン artifact を渡し、npm ci と UI ビルド後 electron-builder でパッケージ化する。

配布内容は dist / dist-electron / package.json と extraResources のエンジンだけの許可リストにする。原典テスト資材、core/testdata、実設定、ソースマップ、作業者のホームパス、環境変数ファイルを含めない。app.asar と外側のファイル一覧、エンジンの版と RPC の応答を CI で検査する。

通常ブランチ/PR は検証と artifact 作成のみ。v* タグの場合は package.json / package-lock.json / Go 版との一致を必須とし、両 platform 成功後に集約 artifact を Release へアップロードする。先に draft を作成し、全ファイルのアップロード後に公開する。既存 Release を上書きしない。公開ジョブだけ contents:write とし、ビルドジョブの権限は read に限定する。

署名証明書は新設しない。macOS は ad-hoc 署名であり Apple notarization なし、Windows も Authenticode なし。その状態をダウンロード案内へ明記する。実行時にユーザーの設定を永続保存しない既存方針を維持する。
