# ハンズオン教材

教材本体は [WORKBOOK.md](WORKBOOK.md) に移しました。

このファイルは入口だけです。今後は以下の順番で進めてください。

1. [WORKBOOK.md](WORKBOOK.md) の章を読む
2. 対応する `internal/*` の TODO を埋める
3. `scripts/check-chapter.sh NN` で章別チェックを実行する
4. Linux TAP 上で実機確認する
5. 詰まったら `final` branch の同名ファイルを見る

例:

```sh
scripts/check-chapter.sh 02
```

