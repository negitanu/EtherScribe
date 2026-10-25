# 設計

公式 PAN-OS 11.1 Configure CLI hierarchy の network dhcp と shared/vsys log-settings syslog を XML 型へ対応付ける。DHCP は client と区別し、server pointer の存在する IF のみ DHCPServers へ入れる。複数プールは IPPools に原形保持し、従来 StartIP/EndIP は先頭の範囲だけに限定する。lease は timeout 原値と型、unlimited を保持し、未確認の時間単位へ変換しない。予約は親サーバーと関連する子表にする。IPv4 relay は独立した DHCPRelays に保持する。

SyslogConfigured は nullable bool（未解析/設定なし/設定あり）。Palo Alto の profile は設定を意味し、転送実行や有効状態を推定しない。shared と vsys は分離し、transport/format/facility/port/name を保持する。FortiOS は syslogd〜syslogd4 を対象に設定有無を求める。

メーカー判定は選択セクションから独立した DeviceInfo.DeviceType を使用する。既知の非 FortiGate では logging の FortiAnalyzer フィールドを JSON/表の両方から除き、FortiOS 専用モデルしか実装されていない HA 節は省略する。メーカー不明の内部モデルは従来の汎用出力を維持する。DHCP は各社で共有可能な項目に正規化し、Palo Alto のリレーを独立表として表示する。

通常テストは core/testdata の合成 XML のみ。未対応 DHCP オプション・IPv6 relay・カスタム Syslog format は値を含まない coverage に残す。
