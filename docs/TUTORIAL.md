# Go + Linux TAP で作る TCP/IP プロトコルスタック

この教材は、完成コードを写経しながら段階的に増やす構成です。最終コードは `cmd/stack` と `internal/*` にあり、各章の「完全な Go コード」では、その章で読むべきファイルと到達点を示します。

章ごとの差分の見方:

| 章 | 追加する主な package | 到達点 |
| --- | --- | --- |
| 1 | `tap`, `hexdump` | TAP から Ethernet フレームを読んで表示 |
| 2 | `ethernet` | EtherType で ARP / IPv4 に振り分け |
| 3 | `arp` | `arping 192.168.100.2` に応答 |
| 4 | `ipv4` | IPv4 header と checksum を処理 |
| 5 | `icmp` | `ping 192.168.100.2` が成功 |
| 6 | `udp` | `nc -u 192.168.100.2 9000` が echo |
| 7 | `tcp` state | TCP 3-way handshake |
| 8 | `tcp` payload | TCP echo |
| 9 | `tcp` FIN | `nc` 終了時の close |
| 10 | docs | Wireshark でデバッグ |
| 11 | docs | 発展課題 |

## 第0章: この教材で作るもの

### 1. この章で作るもの

OS の TCP/IP スタックではなく、TAP デバイスから Ethernet フレームを直接読んで応答する学習用スタックの全体像を確認します。

### 2. 背景知識

Go の `net` package は便利ですが、TCP/IP の大部分は Linux kernel が処理します。この教材では `net.Listen`, `net.Dial`, `net.TCPConn`, `net.UDPConn`, `net/http` を使わず、Go の `os.File.Read` / `Write` で L2 フレームを扱います。

TAP は Ethernet フレームを読み書きできる仮想 NIC です。Linux ホストから見ると `tap0` はネットワークインターフェースで、自作スタックから見ると `/dev/net/tun` に紐づいたファイルです。

### 3. パケット構造の説明

全体は次の入れ子です。

```text
Ethernet
  ARP
  IPv4
    ICMP
    UDP
    TCP
```

### 4. 実装方針

- byte slice を `encoding/binary.BigEndian` で読む
- 各 protocol package に `Parse` と `Marshal` を置く
- main は TAP を開いて `stack.Run` を呼ぶだけにする
- checksum と状態遷移はテスト可能な関数に寄せる

### 5. 完全な Go コード

完成コード全体がこの章の最終形です。入口は [cmd/stack/main.go](/Users/tshimobayashi/sources/go-tcp-tp/cmd/stack/main.go)、プロトコル処理は [internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) です。

### 6. 実行方法

README の TAP 作成後に `go run ./cmd/stack` を実行します。

### 7. 動作確認

最終的に `ping`, UDP echo, TCP echo の 3 つを確認します。

### 8. Wiresharkで見るポイント

最初は `tap0` に ARP が流れることだけ確認します。

### 9. よくあるバグ

- OS 側 IP と自作スタック側 IP を同じにしてしまう
- TAP ではなく TUN を作ってしまう
- Ethernet header 14 bytes を飛ばし忘れる

### 10. 演習問題

1. `tap0` と通常の NIC の違いを自分の言葉で説明してください。
2. なぜ `net.Listen` を使うと TCP 実装の学習にならないのか説明してください。

## 第1章: TAPデバイスを作ってGoから読む

### 1. この章で作るもの

`tap0` を作成し、Go から `/dev/net/tun` を開いて Ethernet フレームを読みます。`ping 192.168.100.2` を実行すると、まず ARP Request が読めることを確認します。

### 2. 背景知識

Linux の TAP は L2 デバイスです。`IFF_TAP | IFF_NO_PI` を指定すると、先頭に Linux 固有の packet information を付けず、Ethernet frame そのものを読めます。

### 3. パケット構造の説明

この章では中身を解釈しません。受信単位は Ethernet frame です。

```text
Dst MAC(6) Src MAC(6) EtherType(2) Payload(...)
```

### 4. 実装方針

1. `ip tuntap` で `tap0` を作る
2. `192.168.100.1/24` を Linux 側に設定する
3. Go で `/dev/net/tun` を開く
4. ioctl `TUNSETIFF` で `tap0` に接続する
5. `os.File.Read` で frame を読む

### 5. 完全な Go コード

この章の中心は [internal/tap/tap_linux.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/tap/tap_linux.go) と [internal/hexdump/hexdump.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/hexdump/hexdump.go) です。最終版の `cmd/stack/main.go` は echo まで進んでいますが、写経時は `Read` して hex dump するだけで構いません。

最小の読み取りループ:

```go
buf := make([]byte, 1600)
for {
    n, err := dev.Read(buf)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(hexdump.Format(buf[:n]))
}
```

### 6. 実行方法

```sh
sudo ip tuntap add dev tap0 mode tap user "$USER"
sudo ip addr add 192.168.100.1/24 dev tap0
sudo ip link set tap0 up
go run ./cmd/stack
```

### 7. 動作確認

別ターミナルで:

```sh
ping 192.168.100.2
```

まだ応答は返しませんが、ARP frame が読めます。

### 8. Wiresharkで見るポイント

`arp` filter を使い、`Who has 192.168.100.2? Tell 192.168.100.1` が見えることを確認します。

### 9. よくあるバグ

- `IFF_NO_PI` を付け忘れて先頭 4 bytes が余分に付く
- `tap0` が down のまま
- `/dev/net/tun` の権限がない

### 10. 演習問題

1. 受信 frame の先頭 14 bytes を手で読み、宛先 MAC、送信元 MAC、EtherType を特定してください。
2. `ip link set tap0 down` にすると何が変わるか確認してください。

## 第2章: Ethernetフレームをパースする

### 1. この章で作るもの

Ethernet header を `Parse` し、EtherType が `0x0806` なら ARP、`0x0800` なら IPv4 に振り分けます。

### 2. 背景知識

Ethernet は LAN 上の配送を担当します。IP address ではなく MAC address を使います。

### 3. パケット構造の説明

```text
0               6               12      14
+---------------+---------------+-------+---------+
| Dst MAC       | Src MAC       | Type  | Payload |
+---------------+---------------+-------+---------+
```

### 4. 実装方針

`internal/ethernet` に `MAC`, `Frame`, `Parse`, `Marshal` を実装します。長さが 14 bytes 未満なら error にします。

### 5. 完全な Go コード

[internal/ethernet/ethernet.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/ethernet/ethernet.go) と [internal/ethernet/ethernet_test.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/ethernet/ethernet_test.go) がこの章の完成コードです。

### 6. 実行方法

`stack.HandleFrame` 内で `ethernet.Parse` を呼びます。まだ応答は返さず、EtherType を log するだけでも構いません。

### 7. 動作確認

`ping 192.168.100.2` 実行時に EtherType `0x0806` が見えます。

### 8. Wiresharkで見るポイント

Ethernet II の Dst, Src, Type を確認します。

### 9. よくあるバグ

- EtherType を little endian で読んでしまう
- Payload の開始位置を 12 bytes と誤解する
- short frame を panic で落としてしまう

### 10. 演習問題

1. Broadcast MAC `ff:ff:ff:ff:ff:ff` を判定する関数を追加してください。
2. 未対応 EtherType を log する処理を追加してください。

## 第3章: ARP Replyを返す

### 1. この章で作るもの

`Who has 192.168.100.2?` に対して、自作スタックの MAC `02:00:00:00:00:02` を返します。

### 2. 背景知識

同じ L2 network で IPv4 packet を送るには、相手 IP に対応する MAC address が必要です。ARP は IPv4 address から MAC address を解決します。

### 3. パケット構造の説明

ARP for Ethernet/IPv4 は 28 bytes です。

```text
HardwareType(2) ProtocolType(2) HLen(1) PLen(1) Operation(2)
SenderMAC(6) SenderIP(4) TargetMAC(6) TargetIP(4)
```

### 4. 実装方針

ARP Request の `TargetIP` が `192.168.100.2` なら Reply を作ります。Reply では Sender が自作スタック、Target が Linux ホストになります。

### 5. 完全な Go コード

[internal/arp/arp.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/arp/arp.go)、[internal/arp/arp_test.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/arp/arp_test.go)、[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の `handleARP` がこの章の完成コードです。

### 6. 実行方法

```sh
go run ./cmd/stack
```

### 7. 動作確認

```sh
arping -I tap0 192.168.100.2
ip neigh show dev tap0
```

`02:00:00:00:00:02` が表示されれば成功です。

### 8. Wiresharkで見るポイント

ARP Reply の Sender MAC/IP と Target MAC/IP が Request と逆になっていることを確認します。

### 9. よくあるバグ

- Operation を `1` のまま返す
- Ethernet の宛先 MAC を broadcast のまま返す
- ARP payload の target と sender を逆にしない

### 10. 演習問題

1. 自分宛でない ARP Request を無視するテストを追加してください。
2. `ip neigh flush dev tap0` 後に再度 ARP が発生することを確認してください。

## 第4章: IPv4ヘッダを読む

### 1. この章で作るもの

IPv4 header を parse し、checksum を検証し、Protocol 番号で ICMP / TCP / UDP に振り分けます。

### 2. 背景知識

IPv4 は L3 の配送を担当します。Ethernet は隣の機器まで、IPv4 は宛先 IP まで、という役割分担です。

### 3. パケット構造の説明

最小 header は 20 bytes です。

```text
Version/IHL TOS TotalLen ID Flags/Fragment TTL Protocol Checksum SrcIP DstIP
```

Protocol は `1=ICMP`, `6=TCP`, `17=UDP` です。

### 4. 実装方針

`IHL * 4` で header 長を計算し、`Total Length` が受信 byte 数を超えないことを確認します。checksum は header の one's complement sum です。

### 5. 完全な Go コード

[internal/ipv4/ipv4.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/ipv4/ipv4.go)、[internal/ipv4/checksum.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/ipv4/checksum.go)、[internal/ipv4/ipv4_test.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/ipv4/ipv4_test.go) がこの章の完成コードです。

### 6. 実行方法

ARP 解決後に `ping` を実行し、EtherType `0x0800` の payload を `ipv4.Parse` します。

### 7. 動作確認

ICMP Echo Request の Protocol が `1`、DstIP が `192.168.100.2` であることを log で確認します。

### 8. Wiresharkで見るポイント

IPv4 header checksum が valid になっていることを確認します。

### 9. よくあるバグ

- IHL を byte 数ではなく word 数のまま使う
- checksum field を zero にせず checksum を作る
- Total Length ではなく受信 frame 全体を IP payload として扱う

### 10. 演習問題

1. TTL が 0 の packet を捨てる処理を追加してください。
2. 自分宛でない DstIP を無視するテストを追加してください。

## 第5章: ICMP Echo Replyを返してpingを通す

### 1. この章で作るもの

ICMP Echo Request に Echo Reply を返し、`ping 192.168.100.2` を成功させます。

### 2. 背景知識

ping は ICMP Echo を使います。TCP や UDP の port は使いません。

### 3. パケット構造の説明

```text
Type(1) Code(1) Checksum(2) Identifier(2) Sequence(2) Payload(...)
```

Echo Request は Type `8`、Echo Reply は Type `0` です。

### 4. 実装方針

Request の Identifier、Sequence、Payload をそのまま Reply にコピーします。IPv4 と Ethernet は送信元と宛先を入れ替えます。

### 5. 完全な Go コード

[internal/icmp/icmp.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/icmp/icmp.go)、[internal/icmp/icmp_test.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/icmp/icmp_test.go)、[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の `handleICMP` がこの章の完成コードです。

### 6. 実行方法

```sh
go run ./cmd/stack
ping 192.168.100.2
```

### 7. 動作確認

`64 bytes from 192.168.100.2` が表示されます。

### 8. Wiresharkで見るポイント

Echo Request と Echo Reply の Identifier と Sequence が一致していることを確認します。

### 9. よくあるバグ

- ICMP checksum を更新し忘れる
- IPv4 Protocol を `1` にし忘れる
- IPv4 Src/Dst は入れ替えたが Ethernet Src/Dst を入れ替えていない

### 10. 演習問題

1. Echo Request 以外の ICMP を無視するテストを追加してください。
2. ping payload の中身が Reply に残っていることを Wireshark で確認してください。

## 第6章: UDP Echo Serverを作る

### 1. この章で作るもの

UDP port `9000` に届いた payload をそのまま返します。

### 2. 背景知識

UDP は connectionless です。送信元 port と宛先 port、長さ、checksum だけを持ち、再送や順序制御はしません。

### 3. パケット構造の説明

```text
SrcPort(2) DstPort(2) Length(2) Checksum(2) Payload(...)
```

IPv4 上の UDP checksum は pseudo header を含みます。

### 4. 実装方針

`DstPort == 9000` の datagram だけ処理します。Reply は port を入れ替え、payload をコピーして返します。

### 5. 完全な Go コード

[internal/udp/udp.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/udp/udp.go)、[internal/udp/udp_test.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/udp/udp_test.go)、[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の `handleUDP` がこの章の完成コードです。

### 6. 実行方法

```sh
go run ./cmd/stack
nc -u 192.168.100.2 9000
```

### 7. 動作確認

`hello` と入力して Enter を押し、同じ文字列が返ることを確認します。

### 8. Wiresharkで見るポイント

`udp.port == 9000` で絞り、Request と Reply の port が逆になっていることを確認します。

### 9. よくあるバグ

- UDP Length を header だけの長さにしてしまう
- pseudo header の Protocol を `17` にし忘れる
- checksum `0` の扱いを誤る

### 10. 演習問題

1. port `9001` に来た datagram を無視するテストを追加してください。
2. UDP checksum 検証を無効にした場合、壊れた packet がどう見えるか確認してください。

## 第7章: TCP 3-way handshakeを実装する

### 1. この章で作るもの

TCP port `8080` で `LISTEN -> SYN_RECEIVED -> ESTABLISHED` まで進め、`nc 192.168.100.2 8080` の接続成立を確認します。

### 2. 背景知識

TCP は信頼性のために sequence number と acknowledgement number を使います。SYN は payload を持たなくても sequence number を 1 消費します。

### 3. パケット構造の説明

```text
SrcPort DstPort Seq Ack DataOffset Flags Window Checksum Urgent Options Payload
```

この章で使う flag は SYN と ACK です。

### 4. 実装方針

1. SYN を受け取る
2. `Ack = clientSeq + 1` の SYN-ACK を返す
3. ACK を受け取り `ESTABLISHED` にする

### 5. 完全な Go コード

[internal/tcp/tcp.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/tcp/tcp.go)、[internal/tcp/state.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/tcp/state.go)、[internal/tcp/checksum.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/tcp/checksum.go)、[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の `handleTCP` がこの章の完成コードです。

### 6. 実行方法

```sh
go run ./cmd/stack
nc 192.168.100.2 8080
```

### 7. 動作確認

Wireshark で SYN, SYN-ACK, ACK の 3 packet が見えれば接続成立です。

### 8. Wiresharkで見るポイント

SYN-ACK の ACK number が `client SYN seq + 1` になっていることを確認します。

### 9. よくあるバグ

- SYN の sequence 消費を忘れる
- TCP checksum の pseudo header を入れ忘れる
- ACK number と Seq number を逆に考える

### 10. 演習問題

1. SYN 以外を LISTEN 状態で受けたら無視するテストを追加してください。
2. 初期 sequence number を固定値から時刻由来に変えてください。

## 第8章: TCP Echo Serverを作る

### 1. この章で作るもの

`ESTABLISHED` 後に payload を受け取り、同じ payload を PSH/ACK で返します。

### 2. 背景知識

TCP は byte stream です。payload 長の分だけ次に期待する remote sequence number が進みます。

### 3. パケット構造の説明

TCP payload は header の data offset 以降です。options があるため、固定の 20 bytes と決め打ちしないことが重要です。

### 4. 実装方針

`seg.Seq == remoteSeq` なら payload を受理し、`remoteSeq += len(payload)` に更新します。返信では `Ack = remoteSeq`、payload は受信内容そのものです。

### 5. 完全な Go コード

[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の `StateEstablished` 分岐がこの章の完成コードです。

### 6. 実行方法

```sh
nc 192.168.100.2 8080
```

### 7. 動作確認

入力した文字がそのまま返れば成功です。

### 8. Wiresharkで見るポイント

echo reply の Seq は server 側の現在値、Ack は client payload の末尾 + 1 になっていることを確認します。

### 9. よくあるバグ

- payload 長だけ ACK を進めない
- ACK-only packet にも echo を返してしまう
- options を考慮せず payload 開始位置を間違える

### 10. 演習問題

1. 空 payload の ACK-only packet に返信しないことを確認してください。
2. 受信 payload を大文字に変換して返す改造をしてください。

## 第9章: TCP接続終了を扱う

### 1. この章で作るもの

FIN を受け取り、ACK と FIN を返して `nc` 終了時に接続を閉じます。

### 2. 背景知識

FIN も SYN と同じく sequence number を 1 消費します。TCP close は本来複数状態を持ちますが、この教材では `CLOSE_WAIT`, `LAST_ACK`, `CLOSED/LISTEN` に簡略化します。

### 3. パケット構造の説明

FIN は TCP Flags の bit です。payload がなくても `Ack = finSeq + 1` にします。

### 4. 実装方針

`ESTABLISHED` で FIN を受けたら `remoteSeq = seg.Seq + 1` とし、FIN/ACK を返します。相手の ACK を受けたら `LISTEN` に戻します。

### 5. 完全な Go コード

[internal/stack/stack.go](/Users/tshimobayashi/sources/go-tcp-tp/internal/stack/stack.go) の TCP close 分岐がこの章の完成コードです。

### 6. 実行方法

`nc` を `Ctrl-D` または `Ctrl-C` で終了します。

### 7. 動作確認

Wireshark で FIN, ACK, FIN, ACK の流れを確認します。

### 8. Wiresharkで見るポイント

FIN を含む packet の次の ACK が sequence + 1 になっていることを確認します。

### 9. よくあるバグ

- FIN の sequence 消費を忘れる
- close 後に状態を LISTEN に戻さず、次の接続を受けられない
- RST を返す OS 側の挙動と混同する

### 10. 演習問題

1. close 後にもう一度 `nc` で接続できることを確認してください。
2. server 側から先に FIN を送る設計に変えると状態がどう変わるか考えてください。

## 第10章: Wiresharkでデバッグする

### 1. この章で作るもの

実装は増やさず、`tap0` を流れる packet を見て典型的なバグを切り分ける方法を学びます。

### 2. 背景知識

自作 stack のバグは log だけでは分かりにくいです。Wireshark は checksum、sequence、MAC address、protocol number を人間が読める形で表示します。

### 3. パケット構造の説明

Ethernet から TCP payload まで、層ごとに展開して確認します。

### 4. 実装方針

Wireshark の filter を使い、1 protocol ずつ見る対象を狭めます。

### 5. 完全な Go コード

この章でコード追加はありません。観察対象は完成コード全体です。

### 6. 実行方法

```sh
sudo wireshark
```

または:

```sh
sudo tcpdump -i tap0 -n -vv -XX
```

### 7. 動作確認

`arp`, `icmp`, `udp.port == 9000`, `tcp.port == 8080` でそれぞれ表示します。

### 8. Wiresharkで見るポイント

- checksum error: pseudo header、長さ、checksum field の zero 化
- ack number mismatch: SYN/FIN/payload の sequence 消費
- wrong MAC address: Ethernet と ARP の sender/target 入れ替え
- wrong protocol: IPv4 Protocol の `1`, `6`, `17`

### 9. よくあるバグ

- Wireshark の checksum offload 表示と本当の checksum error を混同する
- host 側の ARP cache が古い
- `tap0` 以外の interface を見ている

### 10. 演習問題

1. ICMP checksum をわざと壊して、Wireshark の表示を確認してください。
2. TCP ACK number をわざと 1 ずらして、Linux 側の反応を観察してください。

## 第11章: 発展課題

### 1. この章で作るもの

本教材の最小 stack を出発点に、より実用的な機能を検討します。

### 2. 背景知識

本物の TCP/IP stack は、多数の timer、queue、再送、経路制御、セキュリティ処理を持ちます。この教材は packet level の理解を優先して省略しています。

### 3. パケット構造の説明

発展課題ごとに新しい protocol header や状態が増えます。DHCP は UDP、DNS は UDP/TCP、HTTP は TCP の上に乗ります。

### 4. 実装方針

小さな純粋関数と状態遷移テストを増やしながら進めます。

### 5. 完全な Go コード

発展課題は未実装です。既存 package に責務を追加するか、新しい package を切ってください。

### 6. 実行方法

課題ごとに専用 command や test を追加します。

### 7. 動作確認

Wireshark と `go test ./...` の両方で確認します。

### 8. Wiresharkで見るポイント

新しい protocol の header と、下位層の checksum / length が正しいことを確認します。

### 9. よくあるバグ

- TCP 再送 timer が状態と競合する
- 複数接続で 4-tuple を key にしない
- window と buffer 管理を混同する

### 10. 演習問題

1. DHCP client を実装して IP address を固定値ではなく取得してください。
2. DNS query parser を実装してください。
3. TCP 再送 timer を追加してください。
4. TCP window を実装してください。
5. 複数 TCP 接続を 4-tuple で管理してください。
6. HTTP server を TCP echo の上に作ってください。
7. IPv6 の Ethernet Type と header parser を追加してください。
8. packet parser に fuzzing を追加してください。
9. checksum の property-based testing を追加してください。

