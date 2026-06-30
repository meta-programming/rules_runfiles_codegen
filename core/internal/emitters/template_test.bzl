load("@bazel_skylib//lib:unittest.bzl", "asserts", "unittest")
load("//internal/emitters:template.bzl", "render")

def _render_test_impl(ctx):
    env = unittest.begin(ctx)
    
    # Test basic replacement
    asserts.equals(env, "hello world", render("hello {{name}}", name = "world"))
    
    # Test multiple replacements
    asserts.equals(env, "hello world, bye world", render("hello {{name}}, bye {{name}}", name = "world"))
    
    # Test multiple placeholders
    asserts.equals(env, "hello world, bye earth", render("hello {{name1}}, bye {{name2}}", name1 = "world", name2 = "earth"))
    
    # Test no placeholders (should pass if no kwargs)
    asserts.equals(env, "hello", render("hello"))
    
    return unittest.end(env)

render_test = unittest.make(_render_test_impl)

def template_test_suite(name):
    unittest.suite(
        name,
        render_test,
    )
