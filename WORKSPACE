load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.16.6/rules_go-0.16.6.tar.gz",
    sha256 = "ade51a315fa17347e5c31201fdc55aa5ffb913377aa315dceb56ee9725e620ee",
)

http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.16.0/bazel-gazelle-0.16.0.tar.gz"],
    sha256 = "7949fc6cc17b5b191103e97481cf8889217263acf52e00b560683413af204fcb",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains(
    go_version = "1.10.8",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

## Install dep for dependency management
http_file(
    name = "dep_darwin",
    executable = 1,
    sha256 = "1a7bdb0d6c31ecba8b3fd213a1170adf707657123e89dff234871af9e0498be2",
    urls = ["https://github.com/golang/dep/releases/download/v0.5.0/dep-darwin-amd64"],
)

http_file(
    name = "dep_linux",
    executable = 1,
    sha256 = "287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83",
    urls = ["https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64"],
)
