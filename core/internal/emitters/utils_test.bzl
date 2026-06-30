load("@bazel_skylib//lib:unittest.bzl", "asserts", "unittest")
load("//internal:emitters/utils.bzl", "escape_go_string", "escape_kotlin_string", "sanitize_kdoc")

def _escape_go_string_test_impl(ctx):
    env = unittest.begin(ctx)
    asserts.equals(env, "foo", escape_go_string("foo"))
    asserts.equals(env, "foo\\\\bar", escape_go_string("foo\\bar"))
    asserts.equals(env, "foo\\\"bar", escape_go_string("foo\"bar"))
    return unittest.end(env)

def _escape_kotlin_string_test_impl(ctx):
    env = unittest.begin(ctx)
    asserts.equals(env, "foo", escape_kotlin_string("foo"))
    asserts.equals(env, "foo\\\\bar", escape_kotlin_string("foo\\bar"))
    asserts.equals(env, "foo\\\"bar", escape_kotlin_string("foo\"bar"))
    asserts.equals(env, "foo\\$bar", escape_kotlin_string("foo$bar"))
    return unittest.end(env)

def _sanitize_kdoc_test_impl(ctx):
    env = unittest.begin(ctx)
    asserts.equals(env, "foo", sanitize_kdoc("foo"))
    asserts.equals(env, "foo* /bar", sanitize_kdoc("foo*/bar"))
    return unittest.end(env)

escape_go_string_test = unittest.make(_escape_go_string_test_impl)
escape_kotlin_string_test = unittest.make(_escape_kotlin_string_test_impl)
sanitize_kdoc_test = unittest.make(_sanitize_kdoc_test_impl)

def utils_test_suite(name):
    unittest.suite(
        name,
        escape_go_string_test,
        escape_kotlin_string_test,
        sanitize_kdoc_test,
    )
