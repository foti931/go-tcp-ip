# go-tcp-ip

Go と Linux TAP デバイスで TCP/IP プロトコルスタックを自作するハンズオン教材です。

この `main` branch には、ハンズオン用のスターターコードを置いています。型、定数、関数シグネチャ、TAP 読み取りループ、TODO コメントは用意済みです。

読者が全部のソースコードをゼロから設計する必要はありません。各章では、TODO が付いた小さな関数を埋めていきます。

完成版は `final` branch にあります。

```sh
git fetch origin
git checkout final
```

## 進め方

1. [docs/WORKBOOK.md](docs/WORKBOOK.md) を読む
2. `internal/*` の TODO を章ごとに埋める
3. `scripts/check-chapter.sh 02` のように章別チェックを実行する
4. Linux TAP 上で `ping`, `nc` などの実機確認をする
5. 詰まったら `final` branch の同名 package を参照する

## 最終ゴール

Linux host から TAP デバイス越しに、以下が動くことを目標にします。

```sh
ping 192.168.100.2
nc -u 192.168.100.2 9000
nc 192.168.100.2 8080
```

## なぜ main に完成コードを置かないか

この教材の目的は「TCP/IP を packet level で理解すること」です。完成コードを最初から置くと、読者は写経ではなく眺めるだけになりがちです。

ただし、完全な白紙から始める教材でもありません。`main` には以下を置いています。

- package 構成
- struct / const / error 定義
- 関数シグネチャ
- TAP の open 処理
- read/write loop
- 各章の TODO と実装ヒント
- 最初から通る最小テスト
- 章ごとに失敗/成功を確認できる build tag 付きテスト
- 章ごとのチェックコマンド `scripts/check-chapter.sh`

そのため:

- `main`: ハンズオン用スターターコード
- `final`: 完成リファレンス実装

に分けています。

## 必要環境

- Linux
- Go 1.23 以上
- `iproute2`
- `ping`, `arping`, `nc`
- 任意: Wireshark または tcpdump

TAP デバイスを使う実行確認は Linux 前提です。macOS / Windows では Linux VM を使ってください。

## TAP 作成

```sh
sudo ip tuntap add dev tap0 mode tap user "$USER"
sudo ip addr add 192.168.100.1/24 dev tap0
sudo ip link set tap0 up
```

## 章別チェック

例:

```sh
scripts/check-chapter.sh 02
scripts/check-chapter.sh 03
scripts/check-chapter.sh 04
```

各章の TODO を埋める前は失敗します。失敗内容が、その章で満たすべき仕様です。

自作スタック側は以下に固定します。

- MAC: `02:00:00:00:00:02`
- IPv4: `192.168.100.2`
- UDP echo: `9000`
- TCP echo: `8080`

## 参考 RFC

- RFC 826 ARP
- RFC 791 IPv4
- RFC 792 ICMP
- RFC 768 UDP
- RFC 793 TCP
