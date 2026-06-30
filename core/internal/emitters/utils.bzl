# Escaping and sanitization helpers for code emitters.

def escape_go_string(s):
    """Escapes a string for use in a double-quoted Go string literal.

    According to the Go Language Specification (https://go.dev/ref/spec#String_literals):
    "Inside the double quotes, any character except newline and unescaped double quote
    may appear. The signs backslash \\ and double quote \" must be escaped as \\\\ and \\\"
    respectively."

    Since runfile paths do not contain newlines, escaping backslashes and double
    quotes is sufficient to ensure the string literal is well-formed and safe.
    """
    # Escape backslashes first, then double quotes
    return s.replace("\\", "\\\\").replace('"', '\\"')

def escape_kotlin_string(s):
    """Escapes a string for use in a double-quoted Kotlin string literal.

    According to the Kotlin Documentation on String Literals and Templates
    (https://kotlinlang.org/docs/strings.html#string-literals and
    https://kotlinlang.org/docs/strings.html#string-templates):
    - Backslashes and double quotes must be escaped (\\\\ and \\\").
    - The dollar sign ($) must be escaped (\\$) to prevent it from being
      interpreted as the start of a string interpolation expression.
    """
    # Escape backslashes, double quotes, and dollar signs (to prevent interpolation)
    return s.replace("\\", "\\\\").replace('"', '\\"').replace("$", "\\$")

def sanitize_kdoc(doc):
    """Sanitizes a docstring for use inside a Kotlin KDoc block comment.

    Prevents comment injection by escaping the closing '*/' token.
    """
    # Replace */ with * / to prevent early termination of the block comment
    return doc.replace("*/", "* /")
