---
audio: 2.mp3
title: 最速でポッドキャスサイトを構築する
date: 2024-09-06T18:00:00+09:00
description: podbardを使って最速でポッドキャスサイトを構築する方法を解説します。
---

このページでは、podbardで最短でポッドキャスサイトを構築する方法を説明します。

## インストール
まずは、`podbard` コマンドをインストールします。Homebrewまたはgo installでインストールできます。

```console
# Homebrew
$ brew install Songmu/tap/podbard

# go install
$ go install github.com/Songmu/podbard/cmd/podbard@latest
```

## サイトの雛形作成
次に、サイトの雛形を作成します。

```console
$ podbard init <dirname>
```

## 雛形の調整

###  設定ファイル `podbard.yaml`
`podbard.yaml` を開いて適宜調整してください。コメントアウトされていない項目は必須か推奨項目です。artwork指定は消しても大丈夫ですが、Apple Podcastsに登録する場合は必須です。

### 不要なサンプルファイルの削除及び差し替え
`audio/sample.mp3`, `episode/1.md` は不要なので削除してください。また、`static/images/artwork.jpg` はダミーの画像なので、適切な画像に差し替えてください。

## 音声ファイルの配置
`audio/` ディレクトリ直下に配信する音声ファイルを配置してください。MP3またはM4Aをサポートしています。`audio/abc/` のようなサブディレクトリ階層はサポートしていないので、直下にフラットに配置してください。

## エピソードの作成
音声に対応するエピソードファイルを作成します。エピソードファイルは、`episode/` ディレクトリ直下にMarkdownファイル形式で保存します。このファイルは、`podbard episode` サブコマンドで以下のようにベースを生成できます。

```console
$ podbard episode audio/1.mp3
```

このとき、`episode/1.md`という以下のようなMarkdownファイルが生成されます。

```markdown
---
audio: 1.mp3
title: "1"
date: 2024-09-06T21:29:28+09:00
description: "1"
---

<!-- write your episode here in markdown -->
```

ターミナル操作の場合にはエディタが自動起動して編集まで行えます。エディタを自動起動しない `--no-edit` オプションもあります。`episode`サブコマンドにはその他多くのオプションがありますが、ここでは取り上げません。

このファイルを適宜編集してエピソードの情報を記述してください。本文部分がShow Noteになります。

## サイトのビルド
あとは、サイトをビルドするだけです。サイトのビルドは `podbard build` コマンドで行います。

```console
$ podbard build
```

サイトは `public/` ディレクトリに出力されます。このディレクトリを適切なホスティング環境にdeployすることで、ポッドキャストサイトが完成します。

このサイト自体もpodbardで作られており、GitHub ActionsでビルドしてGitHub Pagesへdeployしています。[具体的なワークフローの設定](https://github.com/Songmu/podbard/blob/main/.github/workflows/deploy-pages.yaml)も参考にしてください。

## まとめ

ここでは、必要最小限でサイトを構築する方法を説明しました。詳細な使い方やカスタマイズ方法、様々なユースケースの対応については、次回以降に解説していきます。
