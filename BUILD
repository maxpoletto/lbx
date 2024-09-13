load("@io_bazel_rules_go//go:def.bzl", "go_binary")
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//go:def.bzl", "go_test")

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

go_library(
    name = "lbxclient",
    srcs = glob(["internal/client/*.go"]),
    visibility = ["//visibility:public"],
)

go_test(
    name = "lbxclient_test",
    srcs = glob(["internal/client/*.go"]),
    deps = [":lbxclient"],
    visibility = ["//visibility:public"],
)
