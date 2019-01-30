# Copyright 2018 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_context",
)
load(
    "@io_bazel_rules_go//go/private:rules/rule.bzl",
    "go_rule",
)

def _bindata_impl(ctx):
    go = go_context(ctx)
    out = go.declare_file(go, ext = ".go")
    arguments = ctx.actions.args()
    arguments.add_all([
        "-o",
        out,
        "-pkg",
        ctx.attr.package,
    ])
    if not ctx.attr.compress:
        arguments.add("-nocompress")
    if not ctx.attr.metadata:
        arguments.add("-nometadata")
    if not ctx.attr.memcopy:
        arguments.add("-nomemcopy")
    if not ctx.attr.modtime:
        arguments.add_all(["-modtime", "1437436800"])
    if ctx.attr.extra_args:
        arguments.add_all(ctx.attr.extra_args)
    srcs = sorted(depset([f.dirname for f in ctx.files.srcs] + [f.path for f in ctx.files.strip_file_srcs]).to_list())
    for f in ctx.files.strip_file_srcs:
        arguments.add("-prefix", f.dirname)
    arguments.add_all(srcs)
    ctx.actions.run(
        inputs = ctx.files.srcs + ctx.files.strip_file_srcs,
        outputs = [out],
        mnemonic = "GoBindata",
        executable = ctx.executable._bindata,
        arguments = [arguments],
    )
    return [
        DefaultInfo(
            files = depset([out]),
        ),
    ]

bindata = go_rule(
    _bindata_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "strip_file_srcs": attr.label_list(allow_files = True),
        "package": attr.string(mandatory = True),
        "compress": attr.bool(default = True),
        "metadata": attr.bool(default = False),
        "memcopy": attr.bool(default = True),
        "modtime": attr.bool(default = False),
        "extra_args": attr.string_list(),
        "_bindata": attr.label(
            executable = True,
            cfg = "host",
            default = "//vendor/github.com/kevinburke/go-bindata/go-bindata:go-bindata",
        ),
    },
)
