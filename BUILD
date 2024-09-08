load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "lbxd",
    srcs = ["cmd/server/main.go"],
    visibility = ["//visibility:public"],
)

go_binary(
    name = "lbx",
    srcs = ["cmd/cli/main.go"],
    visibility = ["//visibility:public"],
)
